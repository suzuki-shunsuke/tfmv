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
	Regexp    string
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

type Summary struct {
	Changes []*Change `json:"changes"`
}

func (s *Summary) FromDirs(dirs map[string]*Dir) {
	for _, dir := range dirs {
		for _, block := range dir.Blocks {
			s.Changes = append(s.Changes, &Change{
				Dir:        dir.Path,
				Address:    block.TFAddress,
				NewAddress: block.NewTFAddress,
			})
		}
	}
}

type Change struct {
	Dir        string `json:"dir"`
	Address    string `json:"address"`
	NewAddress string `json:"new_address"`
}
