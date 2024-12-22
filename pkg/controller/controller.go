package controller

import (
	"io"
	"regexp"

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
	Replace   string
	Include   *regexp.Regexp
	Exclude   *regexp.Regexp
	Args      []string
	Recursive bool
	DryRun    bool
}

type Dir struct {
	Path   string
	Files  []string
	Blocks []*Block
}
