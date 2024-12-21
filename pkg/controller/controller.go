package controller

import (
	"fmt"
	"io"

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
	Args      []string
	Recursive bool
	DryRun    bool
}

const wordResource = "resource"

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
