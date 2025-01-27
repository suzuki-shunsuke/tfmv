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

// Init initializes the Controller.
func (c *Controller) Init(fs afero.Fs, stdout, stderr io.Writer) {
	c.fs = fs
	c.stdout = stdout
	c.stderr = stderr
}

type Input struct {
	// Jsonnet is a jsonnet option.
	Jsonnet string
	// MovedFile is -moved option.
	MovedFile string
	// Replace is a replace option.
	Replace string
	// Regexp is a regexp option.
	Regexp string
	// Include is an include option.
	Include *regexp.Regexp
	// Exclude is an exclude option.
	Exclude *regexp.Regexp
	// Args is a list of arguments.
	Args []string
	// Recursive is a recursive option.
	Recursive bool
	// DryRun is a dry-run option.
	DryRun bool
}

// Dir represents a Terraform Module directory.
type Dir struct {
	// Path is a directory path.
	Path string
	// Files a list of file paths (not file names).
	Files []string
	// Blocks is a list of renamed Terraform blocks.
	Blocks []*Block
}

// Summary represents a summary of changes.
// It is used to output a summary of changes.
type Summary struct {
	// Changes is a list of changes.
	Changes []*Change `json:"changes"`
}

// FromDirs updates the Summary from a list of directories.
func (s *Summary) FromDirs(dirs map[string]*Dir) {
	s.Changes = []*Change{}
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

// Change represents a change of a Terraform block.
type Change struct {
	// Dir is a Terraform module directory path.
	Dir string `json:"dir"`
	// Address is a current Terraform address.
	Address string `json:"address"`
	// NewAddress is a new Terraform address.
	NewAddress string `json:"new_address"`
}
