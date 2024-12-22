package controller

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

type Renamer interface {
	Rename(block *Block) (string, error)
}

func NewRenamer(logE *logrus.Entry, fs afero.Fs, input *Input) (Renamer, error) {
	if input.Replace != "" {
		return NewReplaceRenamer(input.Replace)
	}
	if input.File != "" {
		return NewJsonnetRenamer(logE, fs, input.File)
	}
	if input.Regexp != "" {
		return NewRegexpRenamer(input.Regexp)
	}
	return nil, errors.New("one of --jsonnet or --replace or --regexp must be specified")
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

type RegexpRenamer struct {
	regexp *regexp.Regexp
	new    string
}

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

func (r *RegexpRenamer) Rename(block *Block) (string, error) {
	return r.regexp.ReplaceAllString(block.Name, r.new), nil
}
