package controller

import (
	"io"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
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

// Summary represents a summary of changes.
// It is used to output a summary of changes.
type Summary struct {
	// Changes is a list of changes.
	Changes []*Change `json:"changes"`
}

// FromDirs updates the Summary from a list of directories.
func (s *Summary) FromDirs(dirs map[string]*types.Dir) {
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
