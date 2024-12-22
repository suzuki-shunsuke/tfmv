package controller

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Renamer interface {
	Rename(block *Block) (string, error)
}

type ReplaceRenamer struct {
	old string
	new string
}

func NewReplaceRenamer(s string) (*ReplaceRenamer, error) {
	o, n, ok := strings.Cut(s, "/")
	if !ok {
		return nil, fmt.Errorf("--repace must include /: %s", s)
	}
	return &ReplaceRenamer{old: o, new: n}, nil
}

func (r *ReplaceRenamer) Rename(block *Block) (string, error) {
	return strings.ReplaceAll(block.Name, r.old, r.new), nil
}

func NewRenamer(logE *logrus.Entry, fs afero.Fs, input *Input) (Renamer, error) {
	if input.Replace != "" {
		return NewReplaceRenamer(input.Replace)
	}
	if input.File != "" {
		return NewJsonnetRenamer(logE, fs, input.File)
	}
	return nil, errors.New("either --jsonnet or --replace must be specified")
}

func (c *Controller) Run(_ context.Context, logE *logrus.Entry, input *Input) error {
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
			return fmt.Errorf("handle a file: %w", err)
		}
		dir.Blocks = append(dir.Blocks, blocks...)
	}
	for _, dir := range dirs {
		if err := c.handleDir(logE, editor, input, dir); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) handleDir(logE *logrus.Entry, editor *Editor, input *Input, dir *Dir) error {
	// fix references
	if err := c.fixRef(logE, dir); err != nil {
		return err
	}
	for _, block := range dir.Blocks {
		// change resource addressses by hcledit
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
	blocks, err := parse(b, file)
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
		block.SetNewName(newName)
		arr = append(arr, block)
	}
	return arr, nil
}

func (c *Controller) handleBlock(logE *logrus.Entry, editor *Editor, input *Input, block *Block) error {
	// generate moved blocks
	if input.DryRun {
		logE.WithField("moved_file", block.MovedFile).Info("[DRY RUN] generate a moved block")
	} else {
		logE.WithField("moved_file", block.MovedFile).Info("writing a moved block")
		if err := c.writeMovedBlock(block, block.NewName, block.MovedFile); err != nil {
			return fmt.Errorf("write a moved block: %w", err)
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
