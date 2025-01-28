package plan

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

func (c *Planner) findFiles(input *types.Input) ([]string, error) {
	if len(input.Args) != 0 {
		return input.Args, nil
	}
	if input.Recursive {
		return c.walkFiles()
	}
	return afero.Glob(c.fs, "*.tf") //nolint:wrapcheck
}

func (c *Planner) walkFiles() ([]string, error) {
	// find *.tf
	ignoreDirs := map[string]struct{}{
		".git":         {},
		".terraform":   {},
		"node_modules": {},
	}
	files := []string{}
	if err := fs.WalkDir(afero.NewIOFS(c.fs), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if _, ok := ignoreDirs[d.Name()]; ok {
			return fs.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tf") {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk a directory: %w", err)
	}
	return files, nil
}
