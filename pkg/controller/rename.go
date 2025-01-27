package controller

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Renamer is an interface to rename a block address.
type Renamer interface {
	Rename(block *Block) (string, error)
}

// NewRenamer creates a Renamer.
func NewRenamer(logE *logrus.Entry, fs afero.Fs, input *Input) (Renamer, error) {
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

// ReplaceRenamer is a Renamer which renames addresses by replacing a fixed string `old` to `new`.
type ReplaceRenamer struct {
	old string
	new string
}

// NewReplaceRenamer creates a ReplaceRenamer.
// s must be a string "<old>/<new>".
func NewReplaceRenamer(s string) (*ReplaceRenamer, error) {
	o, n, ok := strings.Cut(s, "/")
	if !ok {
		return nil, fmt.Errorf("--replace must include /: %s", s)
	}
	return &ReplaceRenamer{old: o, new: n}, nil
}

// Rename renames a block address.
func (r *ReplaceRenamer) Rename(block *Block) (string, error) {
	return strings.ReplaceAll(block.Name, r.old, r.new), nil
}

// RegexpRenamer is a Renamer which renames addresses by replacing a regular expression `regexp` to `new`.
type RegexpRenamer struct {
	regexp *regexp.Regexp
	new    string
}

// NewRegexpRenamer creates a RegexpRenamer.
func NewRegexpRenamer(s string) (*RegexpRenamer, error) {
	o, n, ok := strings.Cut(s, "/")
	if !ok {
		return nil, fmt.Errorf("--regexp must include /: %s", s)
	}
	r, err := regexp.Compile(o)
	if err != nil {
		return nil, fmt.Errorf("compile a regular expression: %w", err)
	}
	return &RegexpRenamer{regexp: r, new: n}, nil
}

// Rename renames a block address.
func (r *RegexpRenamer) Rename(block *Block) (string, error) {
	return r.regexp.ReplaceAllString(block.Name, r.new), nil
}
