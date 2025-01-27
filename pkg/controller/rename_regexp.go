package controller

import (
	"fmt"
	"regexp"
	"strings"
)

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
