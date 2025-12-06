package rename

import (
	"errors"
	"log/slog"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/domain"
)

// Renamer is an interface to rename a block address.
type Renamer interface {
	Rename(block *domain.Block) (string, error)
}

// New creates a Renamer.
func New(logger *slog.Logger, fs afero.Fs, input *domain.Input) (Renamer, error) {
	if input.Replace != "" {
		return NewReplaceRenamer(input.Replace)
	}
	if input.Jsonnet != "" {
		return NewJsonnetRenamer(logger, fs, input.Jsonnet)
	}
	if input.Regexp != "" {
		return NewRegexpRenamer(input.Regexp)
	}
	return nil, errors.New("one of --jsonnet or --replace or --regexp must be specified")
}
