package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
)

func (c *Controller) Run(logE *logrus.Entry, input *Input) error {
	// read Jsonnet
	renamer, err := NewRenamer(logE, c.fs, input)
	if err != nil {
		return err
	}
	// find *.tf
	logE.Debug("finding tf files")
	files, err := c.findFiles(input)
	if err != nil {
		return fmt.Errorf("find a file: %w", err)
	}
	if len(files) == 0 {
		logE.Warn("no tf file is found")
		return nil
	}
	logE.WithField("num_of_files", len(files)).Debug("found tf files")
	// read *.tf
	editor := &Editor{
		stderr: c.stderr,
		dryRun: input.DryRun,
	}
	dirs := map[string]*Dir{}
	for _, file := range files {
		logE := logE.WithField("file", file)
		logE.Debug("handling a file")
		dirPath := filepath.Dir(file)
		dir, ok := dirs[dirPath]
		if !ok {
			dir = &Dir{Path: dirPath}
			dirs[dirPath] = dir
		}
		dir.Files = append(dir.Files, file)
		blocks, err := c.handleFile(logE, renamer, input, file)
		if err != nil {
			return fmt.Errorf("handle a file: %w", logerr.WithFields(err, logrus.Fields{
				"file": file,
			}))
		}
		dir.Blocks = append(dir.Blocks, blocks...)
	}
	if err := c.summarize(dirs); err != nil {
		logerr.WithError(logE, err).Warn("output changed summary")
	}
	for _, dir := range dirs {
		if err := c.handleDir(logE, editor, input, dir); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) summarize(dirs map[string]*Dir) error {
	summary := &Summary{}
	summary.FromDirs(dirs)
	encoder := json.NewEncoder(c.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(summary); err != nil {
		return fmt.Errorf("encode a summary as JSON: %w", err)
	}
	return nil
}

func (c *Controller) handleDir(logE *logrus.Entry, editor *Editor, input *Input, dir *Dir) error {
	// fix references
	if err := c.fixRef(logE, dir, input); err != nil {
		return err
	}
	for _, block := range dir.Blocks {
		// change resource addresses by hcledit
		// generate moved blocks
		logE := logE.WithFields(logrus.Fields{
			"address":     block.TFAddress,
			"new_address": block.NewTFAddress,
			"file":        block.File,
		})
		if err := c.handleBlock(logE, editor, input, block); err != nil {
			return err
		}
	}
	return nil
}

func getMovedFile(file, dest string) string {
	if dest == "same" {
		dest = filepath.Base(file)
	}
	return filepath.Join(filepath.Dir(file), dest)
}

func (c *Controller) handleFile(logE *logrus.Entry, renamer Renamer, input *Input, file string) ([]*Block, error) {
	logE.Debug("reading a tf file")
	b, err := afero.ReadFile(c.fs, file)
	if err != nil {
		return nil, fmt.Errorf("read a file: %w", err)
	}
	// parse *.tf
	logE.Debug("parsing a tf file")
	blocks, err := parse(b, file, input.Include, input.Exclude)
	if err != nil {
		return nil, fmt.Errorf("parse a HCL file: %w", err)
	}
	if len(blocks) == 0 {
		logE.Debug("no resource or module block is found")
		return nil, nil
	}
	arr := []*Block{}
	movedFile := getMovedFile(file, input.Dest)
	for _, block := range blocks {
		logE := logE.WithFields(logrus.Fields{
			"block_type":    block.BlockType,
			"resource_type": block.ResourceType,
			"name":          block.Name,
		})
		logE.Debug("handling a block")
		block.MovedFile = movedFile
		newName, err := renamer.Rename(block)
		if err != nil {
			return nil, fmt.Errorf("get a new name: %w", err)
		}
		if newName == "" || newName == block.Name {
			continue
		}
		if !hclsyntax.ValidIdentifier(newName) {
			return nil, logerr.WithFields(errors.New("the new name is an invalid HCL identifier"), logrus.Fields{ //nolint:wrapcheck
				"address":  block.TFAddress,
				"new_name": newName,
			})
		}
		block.SetNewName(newName)
		arr = append(arr, block)
	}
	return arr, nil
}

func (c *Controller) handleBlock(logE *logrus.Entry, editor *Editor, input *Input, block *Block) error {
	// generate moved blocks
	if block.BlockType != wordData {
		if input.DryRun {
			logE.WithField("moved_file", block.MovedFile).Debug("[DRY RUN] generate a moved block")
		} else {
			logE.WithField("moved_file", block.MovedFile).Debug("writing a moved block")
			if err := c.writeMovedBlock(block, block.NewName, block.MovedFile); err != nil {
				return fmt.Errorf("write a moved block: %w", err)
			}
		}
	}

	// rename resources
	if err := editor.Move(logE, &MoveBlockOpt{
		From:     block.HCLAddress,
		To:       block.NewHCLAddress,
		FilePath: block.File,
		Update:   true,
	}); err != nil {
		return fmt.Errorf("move a block: %w", err)
	}
	return nil
}

func applyFixes(body string, blocks []*Block) string {
	for _, b := range blocks {
		body = b.Fix(body)
	}
	return body
}

func (c *Controller) fixRef(logE *logrus.Entry, dir *Dir, input *Input) error {
	files := dir.Files
	if len(input.Args) != 0 {
		arr, err := afero.Glob(c.fs, filepath.Join(dir.Path, "*.tf"))
		if err != nil {
			return fmt.Errorf("find a file: %w", err)
		}
		files = arr
	}
	for _, file := range files {
		fields := logrus.Fields{"file": file}
		b, err := afero.ReadFile(c.fs, file)
		if err != nil {
			return fmt.Errorf("read a file: %w", logerr.WithFields(err, fields))
		}
		orig := string(b)
		s := applyFixes(orig, dir.Blocks)
		if orig == s {
			continue
		}
		f, err := c.fs.Stat(file)
		if err != nil {
			return fmt.Errorf("get a file stat: %w", logerr.WithFields(err, fields))
		}
		if input.DryRun {
			logE.WithFields(fields).Debug("[DRY RUN] fixing references")
		} else {
			logE.WithFields(fields).Debug("fixing references")
			if err := afero.WriteFile(c.fs, file, []byte(s), f.Mode()); err != nil {
				return fmt.Errorf("write a file: %w", logerr.WithFields(err, fields))
			}
		}
	}
	return nil
}
