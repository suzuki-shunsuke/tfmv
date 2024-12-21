package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/controller"
)

type Runner struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	LDFlags *LDFlags
	LogE    *logrus.Entry
}

type LDFlags struct {
	Version string
	Commit  string
	Date    string
}

func (r *Runner) Run(ctx context.Context, args ...string) error {
	flg := &Flag{}
	parseFlags(flg)
	if flg.Version {
		fmt.Fprintln(r.Stdout, r.LDFlags.Version)
		return nil
	}
	if flg.Help {
		fmt.Fprintln(r.Stdout, help)
		return nil
	}
	if flg.Dest != "same" {
		if !strings.HasSuffix(flg.Dest, ".tf") {
			return errors.New("--dest name must be either 'same' or a file name with the suffix .tf")
		}
		if filepath.Base(flg.Dest) != flg.Dest {
			return errors.New("--dest name must be either 'same' or a file name with the suffix .tf")
		}
	}
	ctrl := &controller.Controller{}
	ctrl.Init(afero.NewOsFs(), r.Stdout, r.Stderr)
	return ctrl.Run(ctx, r.LogE, &controller.Input{ //nolint:wrapcheck
		File:      flg.File,
		Dest:      flg.Dest,
		Recursive: flg.Recursive,
		Args:      args,
	})
}

type Flag struct {
	File      string
	Dest      string
	Help      bool
	Version   bool
	Recursive bool
}

func parseFlags(f *Flag) {
	flag.StringVar(&f.File, "file", "", "Jsonnet file path")
	flag.StringVar(&f.Dest, "dest", "moved.tf", "The destination file name")
	flag.BoolVar(&f.Help, "help", false, "Show help")
	flag.BoolVar(&f.Version, "version", false, "Show version")
	flag.BoolVar(&f.Recursive, "r", false, "If this is set, tfmv finds files recursively")
	flag.Parse()
}

const help = `tfmv - Rename Terraform resources and modules and generate moved blocks.
https://github.com/suzuki-shunsuke/tfmv

Usage:
	tfmv [-help] [-version] [-file <Jsonnet file path>] [-r] [-dest <file name|same>] [file ...]

Options:
	-help		Show help
	-version	Show sort-issue-template version
	-file		Jsonnet file path
	-r			If this is set, tfmv finds files recursively
	-dest		The destination file name. If this is "same", the file is same with the resource`
