package controller

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Renamer interface {
	Rename(block *Block) (string, error)
}

type ReplaceRenamer struct {
	old string
	new string
}

func NewReplaceRenamer(s string) (*ReplaceRenamer, error) {
	o, n, ok := strings.Cut(s, "/")
	if !ok {
		return nil, fmt.Errorf("--repace must include /: %s", s)
	}
	return &ReplaceRenamer{old: o, new: n}, nil
}

func (r *ReplaceRenamer) Rename(block *Block) (string, error) {
	return strings.ReplaceAll(block.Name, r.old, r.new), nil
}

func NewRenamer(logE *logrus.Entry, fs afero.Fs, input *Input) (Renamer, error) {
	if input.Replace != "" {
		return NewReplaceRenamer(input.Replace)
	}
	if input.File != "" {
		return NewJsonnetRenamer(logE, fs, input.File)
	}
	return nil, errors.New("either --jsonnet or --replace must be specified")
}
