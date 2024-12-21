package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/minamijoyo/hcledit/editor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Controller struct {
	fs     afero.Fs
	stdout io.Writer
	stderr io.Writer
}

func (c *Controller) Init(fs afero.Fs, stdout, stderr io.Writer) {
	c.fs = fs
	c.stdout = stdout
	c.stderr = stderr
}

type Input struct {
	File      string
	Dest      string
	Recursive bool
	Args      []string
}

const wordResource = "resource"

func (c *Controller) Run(_ context.Context, logE *logrus.Entry, input *Input) error {
	// read Jsonnet
	b, err := afero.ReadFile(c.fs, input.File)
	if err != nil {
		return fmt.Errorf("read a jsonnet file: %w", err)
	}
	// parse Jsonnet
	ja, err := jsonnet.SnippetToAST(input.File, string(b))
	if err != nil {
		return fmt.Errorf("parse a jsonnet file: %w", err)
	}
	// find *.tf
	files, err := c.findFiles(input)
	if err != nil {
		return fmt.Errorf("find a file: %w", err)
	}
	// read *.tf
	editor := &Editor{
		stderr: c.stderr,
	}
	for _, file := range files {
		if err := c.handleFile(logE, editor, ja, input, file); err != nil {
			return fmt.Errorf("handle a file: %w", err)
		}
	}
	return nil
}

func (c *Controller) handleFile(logE *logrus.Entry, editor *Editor, ja ast.Node, input *Input, file string) error {
	b, err := afero.ReadFile(c.fs, file)
	if err != nil {
		return fmt.Errorf("read a file: %w", err)
	}
	// parse *.tf
	blocks, err := Parse(b, file)
	if err != nil {
		return fmt.Errorf("parse a HCL file: %w", err)
	}
	for _, block := range blocks {
		if err := c.handleBlock(logE, editor, ja, input, block); err != nil {
			return fmt.Errorf("handle a block: %w", err)
		}
	}
	return nil
}

func (c *Controller) handleBlock(logE *logrus.Entry, editor *Editor, ja ast.Node, input *Input, block *Block) error {
	// evaluate Jsonnet
	dest, err := c.evaluate(block, ja)
	if err != nil {
		return fmt.Errorf("evaluate Jsonnet: %w", err)
	}
	if dest == "" || dest == block.Name {
		return nil
	}
	// generate moved blocks
	fileName := input.Dest
	if fileName == "same" {
		fileName = filepath.Base(block.File)
	}
	movedFile := filepath.Join(filepath.Dir(block.File), fileName)
	if err := c.writeMovedBlock(block, dest, movedFile); err != nil {
		return fmt.Errorf("write a moved block: %w", err)
	}
	// rename resources
	if err := editor.Move(logE, &MoveBlockOpt{
		From:     block.Address(),
		To:       block.NewAddress(dest),
		FilePath: movedFile,
		Stdout:   c.stdout,
	}); err != nil {
		return fmt.Errorf("move a block: %w", err)
	}
	return nil
}

func (c *Controller) findFiles(input *Input) ([]string, error) {
	if input.Args != nil {
		return input.Args, nil
	}
	if input.Recursive {
		return c.walkFiles()
	}
	return afero.Glob(c.fs, "*.tf") //nolint:wrapcheck
}

func (c *Controller) walkFiles() ([]string, error) {
	// find *.tf
	ignoreDirs := map[string]struct{}{
		".git":         {},
		".terraform":   {},
		"node_modules": {},
	}
	files := []string{}
	if err := fs.WalkDir(afero.NewIOFS(c.fs), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if _, ok := ignoreDirs[d.Name()]; ok {
			return fs.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tf") {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk a directory: %w", err)
	}
	return files, nil
}

func (c *Controller) evaluate(block *Block, ja ast.Node) (string, error) {
	b, err := json.Marshal(block)
	if err != nil {
		return "", fmt.Errorf("marshal a block: %w", err)
	}
	vm := NewVM(string(b))
	result, err := vm.Evaluate(ja)
	if err != nil {
		return "", fmt.Errorf("evaluate Jsonnet: %w", err)
	}
	var dest string
	if err := json.Unmarshal([]byte(result), &dest); err != nil {
		return "", fmt.Errorf("unmarshal as a JSON: %w", err)
	}
	return dest, nil
}

func (c *Controller) writeMovedBlock(block *Block, dest, movedFile string) error {
	file, err := c.fs.OpenFile(movedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:mnd
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer file.Close()
	fmt.Fprintf(file, `moved {
  from = %s.%s
  to   = %s.%s
}`, block.ResourceType, block.Name, block.ResourceType, dest)
	return nil
}

type Block struct {
	File         string `json:"file"`
	BlockType    string `json:"block_type"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
}

func (b *Block) IsResouce() bool {
	return b.BlockType == wordResource
}

func (b *Block) Address() string {
	if b.IsResouce() {
		return fmt.Sprintf("resource.%s.%s", b.ResourceType, b.Name)
	}
	return "module." + b.Name
}

func (b *Block) NewAddress(name string) string {
	if b.IsResouce() {
		return fmt.Sprintf("resource.%s.%s", b.ResourceType, name)
	}
	return "module." + name
}

func Parse(src []byte, filePath string) ([]*Block, error) {
	file, diags := hclsyntax.ParseConfig(src, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("convert file body to body type")
	}
	blocks := make([]*Block, 0, len(body.Blocks))
	for _, block := range body.Blocks {
		if block.Type != wordResource && block.Type != "module" {
			continue
		}
		b := &Block{
			File:      filePath,
			BlockType: block.Type,
		}
		switch len(block.Labels) {
		case 1:
			b.Name = block.Labels[0]
		case 2: //nolint:mnd
			b.ResourceType = block.Labels[0]
			b.Name = block.Labels[1]
		default:
			continue
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

type Editor struct {
	stderr io.Writer
	dryRun bool
}

type MoveBlockOpt struct {
	// From is a source address.
	From string
	// To is a new address.
	To       string
	FilePath string
	Stdin    io.Reader
	Stdout   io.Writer
	// If `Update` is true, the Terraform Configuration is updated in-place.
	Update bool
}

func (e *Editor) Move(logE *logrus.Entry, opt *MoveBlockOpt) error {
	filter := editor.NewBlockRenameFilter(opt.From, opt.To)
	cl := editor.NewClient(&editor.Option{
		InStream:  opt.Stdin,
		OutStream: opt.Stdout,
		ErrStream: e.stderr,
	})

	if e.dryRun {
		logE.Info("[DRY RUN] move a block")
		return nil
	}
	logE.Info("move a block")

	if err := cl.Edit(opt.FilePath, opt.Update, filter); err != nil {
		return fmt.Errorf("move a block in %s from %s to %s: %w", opt.FilePath, opt.From, opt.To, err)
	}
	return nil
}
