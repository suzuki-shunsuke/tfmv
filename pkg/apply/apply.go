package apply

import (
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/tfmv/pkg/domain"
)

type Applier struct {
	fs     afero.Fs
	stderr io.Writer
}

func New(fs afero.Fs, stderr io.Writer) *Applier {
	return &Applier{
		fs:     fs,
		stderr: stderr,
	}
}

func (a *Applier) Apply(logger *slog.Logger, input *domain.Input, dirs map[string]*domain.Dir) error {
	editor := &Editor{
		stderr: a.stderr,
		dryRun: input.DryRun,
	}
	for _, dir := range dirs {
		if err := a.handleDir(logger, editor, input, dir); err != nil {
			return err
		}
	}
	return nil
}

// handleDir modifies files in a given directory.
func (a *Applier) handleDir(logger *slog.Logger, editor *Editor, input *domain.Input, dir *domain.Dir) error {
	// fix references
	if err := a.fixRef(logger, dir, input); err != nil {
		return err
	}
	for _, block := range dir.Blocks {
		// change resource addresses by hcledit
		// generate moved blocks
		logger := logger.With(
			"address", block.TFAddress,
			"new_address", block.NewTFAddress,
			"file", block.File,
		)
		if err := a.handleBlock(logger, editor, input, block); err != nil {
			return err
		}
	}
	return nil
}

func applyFixes(body string, blocks []*domain.Block) string {
	for _, b := range blocks {
		body = b.Fix(body)
	}
	return body
}

func (a *Applier) fixRef(logger *slog.Logger, dir *domain.Dir, input *domain.Input) error {
	files := dir.Files
	if len(input.Args) != 0 {
		arr, err := afero.Glob(a.fs, filepath.Join(dir.Path, "*.tf"))
		if err != nil {
			return fmt.Errorf("find a file: %w", err)
		}
		files = arr
	}
	for _, file := range files {
		b, err := afero.ReadFile(a.fs, file)
		if err != nil {
			return fmt.Errorf("read a file: %w", slogerr.With(err, "file", file))
		}
		orig := string(b)
		s := applyFixes(orig, dir.Blocks)
		if orig == s {
			continue
		}
		f, err := a.fs.Stat(file)
		if err != nil {
			return fmt.Errorf("get a file stat: %w", slogerr.With(err, "file", file))
		}
		if input.DryRun {
			logger.Debug("[DRY RUN] fixing references", "file", file)
		} else {
			logger.Debug("fixing references", "file", file)
			if err := afero.WriteFile(a.fs, file, []byte(s), f.Mode()); err != nil {
				return fmt.Errorf("write a file: %w", slogerr.With(err, "file", file))
			}
		}
	}
	return nil
}

// handleBlock generates a moved block and renames a block.
func (a *Applier) handleBlock(logger *slog.Logger, editor *Editor, input *domain.Input, block *domain.Block) error {
	// generate moved blocks
	if !block.IsData() {
		if input.DryRun {
			logger.Debug("[DRY RUN] generate a moved block", "moved_file", block.MovedFile)
		} else {
			logger.Debug("writing a moved block", "moved_file", block.MovedFile)
			if err := a.writeMovedBlock(block, block.MovedFile); err != nil {
				return fmt.Errorf("write a moved block: %w", err)
			}
		}
	}

	// rename resources
	if err := editor.Move(logger, &MoveBlockOpt{
		From:     block.HCLAddress,
		To:       block.NewHCLAddress,
		FilePath: block.File,
		Update:   true,
	}); err != nil {
		return fmt.Errorf("move a block: %w", err)
	}
	return nil
}
