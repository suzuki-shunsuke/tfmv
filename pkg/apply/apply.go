package apply

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
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

func (a *Applier) Apply(logE *logrus.Entry, input *types.Input, dirs map[string]*types.Dir) error {
	editor := &Editor{
		stderr: a.stderr,
		dryRun: input.DryRun,
	}
	for _, dir := range dirs {
		if err := a.handleDir(logE, editor, input, dir); err != nil {
			return err
		}
	}
	return nil
}

// handleDir modifies files in a given directory.
func (a *Applier) handleDir(logE *logrus.Entry, editor *Editor, input *types.Input, dir *types.Dir) error {
	// fix references
	if err := a.fixRef(logE, dir, input); err != nil {
		return err
	}
	for _, block := range dir.Blocks {
		// change resource addresses by hcledit
		// generate moved blocks
		logE := logE.WithFields(logrus.Fields{
			"address":     block.TFAddress,
			"new_address": block.NewTFAddress,
			"file":        block.File,
		})
		if err := a.handleBlock(logE, editor, input, block); err != nil {
			return err
		}
	}
	return nil
}

func applyFixes(body string, blocks []*types.Block) string {
	for _, b := range blocks {
		body = b.Fix(body)
	}
	return body
}

func (a *Applier) fixRef(logE *logrus.Entry, dir *types.Dir, input *types.Input) error {
	files := dir.Files
	if len(input.Args) != 0 {
		arr, err := afero.Glob(a.fs, filepath.Join(dir.Path, "*.tf"))
		if err != nil {
			return fmt.Errorf("find a file: %w", err)
		}
		files = arr
	}
	for _, file := range files {
		fields := logrus.Fields{"file": file}
		b, err := afero.ReadFile(a.fs, file)
		if err != nil {
			return fmt.Errorf("read a file: %w", logerr.WithFields(err, fields))
		}
		orig := string(b)
		s := applyFixes(orig, dir.Blocks)
		if orig == s {
			continue
		}
		f, err := a.fs.Stat(file)
		if err != nil {
			return fmt.Errorf("get a file stat: %w", logerr.WithFields(err, fields))
		}
		if input.DryRun {
			logE.WithFields(fields).Debug("[DRY RUN] fixing references")
		} else {
			logE.WithFields(fields).Debug("fixing references")
			if err := afero.WriteFile(a.fs, file, []byte(s), f.Mode()); err != nil {
				return fmt.Errorf("write a file: %w", logerr.WithFields(err, fields))
			}
		}
	}
	return nil
}

// handleBlock generates a moved block and renames a block.
func (a *Applier) handleBlock(logE *logrus.Entry, editor *Editor, input *types.Input, block *types.Block) error {
	// generate moved blocks
	if !block.IsData() {
		if input.DryRun {
			logE.WithField("moved_file", block.MovedFile).Debug("[DRY RUN] generate a moved block")
		} else {
			logE.WithField("moved_file", block.MovedFile).Debug("writing a moved block")
			if err := a.writeMovedBlock(block, block.NewName, block.MovedFile); err != nil {
				return fmt.Errorf("write a moved block: %w", err)
			}
		}
	}

	// rename resources
	if err := editor.Move(logE, &MoveBlockOpt{
		From:     block.HCLAddress,
		To:       block.NewHCLAddress,
		FilePath: block.File,
		Update:   true,
	}); err != nil {
		return fmt.Errorf("move a block: %w", err)
	}
	return nil
}
