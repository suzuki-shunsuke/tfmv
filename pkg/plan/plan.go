package plan

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/suzuki-shunsuke/tfmv/pkg/rename"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

type Planner struct {
	fs afero.Fs
}

func NewPlanner(fs afero.Fs) *Planner {
	return &Planner{
		fs: fs,
	}
}

func (c *Planner) Plan(logE *logrus.Entry, input *types.Input) (map[string]*types.Dir, error) {
	renamer, err := rename.New(logE, c.fs, input)
	if err != nil {
		return nil, fmt.Errorf("initialize a renamer: %w", err)
	}

	// find *.tf
	logE.Debug("finding tf files")
	files, err := c.findFiles(input)
	if err != nil {
		return nil, fmt.Errorf("find a file: %w", err)
	}
	if len(files) == 0 {
		logE.Warn("no tf file is found")
		return nil, nil //nolint:nilnil
	}
	logE.WithField("num_of_files", len(files)).Debug("found tf files")

	// read *.tf
	dirs := map[string]*types.Dir{}
	for _, file := range files {
		logE := logE.WithField("file", file)
		logE.Debug("handling a file")
		dirPath := filepath.Dir(file)
		dir, ok := dirs[dirPath]
		if !ok {
			dir = &types.Dir{Path: dirPath}
			dirs[dirPath] = dir
		}
		dir.Files = append(dir.Files, file)
		blocks, err := c.handleFile(logE, renamer, input, file)
		if err != nil {
			return nil, fmt.Errorf("handle a file: %w", logerr.WithFields(err, logrus.Fields{
				"file": file,
			}))
		}
		dir.Blocks = append(dir.Blocks, blocks...)
	}
	return dirs, nil
}

// handleFile reads and parses a file and returns renamed blocks.
// handleFile doesn't actually edit a file.
func (c *Planner) handleFile(logE *logrus.Entry, renamer rename.Renamer, input *types.Input, file string) ([]*types.Block, error) {
	logE.Debug("reading a tf file")
	b, err := afero.ReadFile(c.fs, file)
	if err != nil {
		return nil, fmt.Errorf("read a file: %w", err)
	}
	logE.Debug("parsing a tf file")
	blocks, err := parse(b, file, input.Include, input.Exclude)
	if err != nil {
		return nil, fmt.Errorf("parse a HCL file: %w", err)
	}
	if len(blocks) == 0 {
		logE.Debug("no resource or module block is found")
		return nil, nil
	}
	arr := []*types.Block{}
	movedFile := getMovedFile(file, input.MovedFile)
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

// getMovedFile returns a file path where moved blocks are written.
func getMovedFile(file, dest string) string {
	if dest == "same" {
		dest = filepath.Base(file)
	}
	return filepath.Join(filepath.Dir(file), dest)
}
