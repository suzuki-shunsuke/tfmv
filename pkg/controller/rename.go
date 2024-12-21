package controller

import (
	"fmt"
	"io"

	"github.com/minamijoyo/hcledit/editor"
	"github.com/sirupsen/logrus"
)

type Editor struct {
	stderr io.Writer
	dryRun bool
}

type MoveBlockOpt struct {
	// From is a source address.
	From string
	// To is a new address.
	To       string
	FilePath string
	Stdin    io.Reader
	Stdout   io.Writer
	// If `Update` is true, the Terraform Configuration is updated in-place.
	Update bool
}

func (e *Editor) Move(logE *logrus.Entry, opt *MoveBlockOpt) error {
	filter := editor.NewBlockRenameFilter(opt.From, opt.To)
	cl := editor.NewClient(&editor.Option{
		InStream:  opt.Stdin,
		OutStream: opt.Stdout,
		ErrStream: e.stderr,
	})

	if e.dryRun {
		logE.Info("[DRY RUN] move a block")
		return nil
	}
	logE.Info("move a block")

	if err := cl.Edit(opt.FilePath, opt.Update, filter); err != nil {
		return fmt.Errorf("move a block in %s from %s to %s: %w", opt.FilePath, opt.From, opt.To, err)
	}
	return nil
}
