package controller

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (c *Controller) Run(_ context.Context, logE *logrus.Entry, input *Input) error {
	// read Jsonnet
	logE.Debug("reading a jsonnet file")
	b, err := afero.ReadFile(c.fs, input.File)
	if err != nil {
		return fmt.Errorf("read a jsonnet file: %w", err)
	}
	// parse Jsonnet
	logE.Debug("parsing a jsonnet file")
	ja, err := jsonnet.SnippetToAST(input.File, string(b))
	if err != nil {
		return fmt.Errorf("parse a jsonnet file: %w", err)
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
	for _, file := range files {
		logE := logE.WithField("file", file)
		logE.Debug("handling a file")
		if err := c.handleFile(logE, editor, ja, input, file); err != nil {
			return fmt.Errorf("handle a file: %w", err)
		}
	}
	return nil
}

func (c *Controller) handleFile(logE *logrus.Entry, editor *Editor, ja ast.Node, input *Input, file string) error {
	logE.Debug("reading a tf file")
	b, err := afero.ReadFile(c.fs, file)
	if err != nil {
		return fmt.Errorf("read a file: %w", err)
	}
	// parse *.tf
	logE.Debug("parsing a tf file")
	blocks, err := parse(b, file)
	if err != nil {
		return fmt.Errorf("parse a HCL file: %w", err)
	}
	if len(blocks) == 0 {
		logE.Debug("no resource or module block is found")
		return nil
	}
	for _, block := range blocks {
		logE := logE.WithFields(logrus.Fields{
			"block_type":    block.BlockType,
			"resource_type": block.ResourceType,
			"name":          block.Name,
		})
		logE.Debug("handling a block")
		if err := c.handleBlock(logE, editor, ja, input, file, block); err != nil {
			return fmt.Errorf("handle a block: %w", err)
		}
	}
	return nil
}

func (c *Controller) handleBlock(logE *logrus.Entry, editor *Editor, ja ast.Node, input *Input, file string, block *Block) error {
	// evaluate Jsonnet
	dest, err := c.evaluate(block, ja)
	if err != nil {
		return fmt.Errorf("evaluate Jsonnet: %w", err)
	}
	logE.WithField("new_name", dest).Debug("evaluate Jsonnet")
	if dest == "" || dest == block.Name {
		return nil
	}
	// generate moved blocks
	fileName := input.Dest
	if fileName == "same" {
		fileName = filepath.Base(block.File)
	}
	movedFile := filepath.Join(filepath.Dir(block.File), fileName)
	logE.WithField("moved_file", movedFile).Debug("generating a moved block")
	if input.DryRun {
		logE.WithField("moved_file", movedFile).Info("[DRY RUN] generate a moved block")
	} else {
		if err := c.writeMovedBlock(block, dest, movedFile); err != nil {
			return fmt.Errorf("write a moved block: %w", err)
		}
	}

	// rename resources
	logE.Debug("moving a block")
	if err := editor.Move(logE, &MoveBlockOpt{
		From:     block.Address(),
		To:       block.NewAddress(dest),
		FilePath: file,
		Update:   true,
	}); err != nil {
		return fmt.Errorf("move a block: %w", err)
	}
	return nil
}
