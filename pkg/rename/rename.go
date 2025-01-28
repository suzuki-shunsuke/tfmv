package rename

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

// Renamer is an interface to rename a block address.
type Renamer interface {
	Rename(block *types.Block) (string, error)
}

// New creates a Renamer.
func New(logE *logrus.Entry, fs afero.Fs, input *types.Input) (Renamer, error) {
	if input.Replace != "" {
		return NewReplaceRenamer(input.Replace)
	}
	if input.Jsonnet != "" {
		return NewJsonnetRenamer(logE, fs, input.Jsonnet)
	}
	if input.Regexp != "" {
		return NewRegexpRenamer(input.Regexp)
	}
	return nil, errors.New("one of --jsonnet or --replace or --regexp must be specified")
}
