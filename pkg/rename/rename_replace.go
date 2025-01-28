package rename

import (
	"fmt"
	"strings"

	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

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
func (r *ReplaceRenamer) Rename(block *types.Block) (string, error) {
	return strings.ReplaceAll(block.Name, r.old, r.new), nil
}
