package controller

import (
	"fmt"
	"io"
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
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
	Replace   string
	Args      []string
	Recursive bool
	DryRun    bool
}

const wordResource = "resource"

type Block struct {
	File          string         `json:"file"`
	BlockType     string         `json:"block_type"`
	ResourceType  string         `json:"resource_type"`
	Name          string         `json:"name"`
	NewName       string         `json:"-"`
	MovedFile     string         `json:"-"`
	Regexp        *regexp.Regexp `json:"-"`
	TFAddress     string         `json:"-"`
	HCLAddress    string         `json:"-"`
	NewTFAddress  string         `json:"-"`
	NewHCLAddress string         `json:"-"`
}

func (b *Block) IsResource() bool {
	return b.BlockType == wordResource
}

func hclAddress(blockType, resourceType, name string) string {
	if blockType == wordResource {
		return fmt.Sprintf("resource.%s.%s", resourceType, name)
	}
	return "module." + name
}

func tfAddress(blockType, resourceType, name string) string {
	if blockType == wordResource {
		return fmt.Sprintf("%s.%s", resourceType, name)
	}
	return "module." + name
}

func (b *Block) Regstr() string {
	// A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes.
	if b.IsResource() {
		return fmt.Sprintf(`\b%s\.%s\b`, b.ResourceType, b.Name)
	}
	return fmt.Sprintf(`\bmodule\.%s\b`, b.Name)
}

type Dir struct {
	Path   string
	Files  []string
	Blocks []*Block
}

type Fix struct {
	Regexp     *regexp.Regexp
	NewAddress string
}

func (b *Block) Fix(body string) string {
	return b.Regexp.ReplaceAllString(body, b.NewTFAddress)
}

func applyFixes(body string, blocks []*Block) string {
	for _, b := range blocks {
		body = b.Fix(body)
	}
	return body
}

type RenamedBlock struct {
	Dir        string
	Address    string
	NewAddress string
}

func (c *Controller) fixRef(logE *logrus.Entry, dir *Dir) error {
	for _, file := range dir.Files {
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
		logE.WithFields(fields).Info("fixing references")
		if err := afero.WriteFile(c.fs, file, []byte(s), f.Mode()); err != nil {
			return fmt.Errorf("write a file: %w", logerr.WithFields(err, fields))
		}
	}
	return nil
}

func (b *Block) SetNewName(newName string) {
	b.NewName = newName
	b.NewHCLAddress = hclAddress(b.BlockType, b.ResourceType, newName)
	b.NewTFAddress = tfAddress(b.BlockType, b.ResourceType, newName)
}

func (b *Block) Init() error {
	b.TFAddress = tfAddress(b.BlockType, b.ResourceType, b.Name)
	b.HCLAddress = hclAddress(b.BlockType, b.ResourceType, b.Name)
	reg, err := regexp.Compile(b.Regstr())
	if err != nil {
		return fmt.Errorf("compile a regular expression to capture a resource reference: %w", err)
	}
	b.Regexp = reg
	return nil
}
