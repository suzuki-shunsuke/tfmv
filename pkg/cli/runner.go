package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
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
	if flg.Moved != "same" {
		if !strings.HasSuffix(flg.Moved, ".tf") {
			return errors.New("--moved name must be either 'same' or a file name with the suffix .tf")
		}
		if filepath.Base(flg.Moved) != flg.Moved {
			return errors.New("--moved name must be either 'same' or a file name with the suffix .tf")
		}
	}
	ctrl := &controller.Controller{}
	ctrl.Init(afero.NewOsFs(), r.Stdout, r.Stderr)
	return ctrl.Run(ctx, r.LogE, &controller.Input{ //nolint:wrapcheck
		File:      flg.Jsonnet,
		Dest:      flg.Moved,
		Recursive: flg.Recursive,
		Args:      args,
	})
}

type Flag struct {
	Jsonnet   string
	Moved     string
	Help      bool
	Version   bool
	Recursive bool
}

func parseFlags(f *Flag) {
	flag.StringVarP(&f.Jsonnet, "jsonnet", "j", "", "Jsonnet file path")
	flag.StringVarP(&f.Moved, "moved", "m", "moved.tf", "The destination file name")
	flag.BoolVarP(&f.Help, "help", "h", false, "Show help")
	flag.BoolVarP(&f.Version, "version", "v", false, "Show version")
	flag.BoolVarP(&f.Recursive, "recursive", "r", false, "If this is set, tfmv finds files recursively")
	flag.Parse()
}

const help = `tfmv - Rename Terraform resources and modules and generate moved blocks.
https://github.com/suzuki-shunsuke/tfmv

Usage:
	tfmv [--jsonnet <Jsonnet file path>] [--recursive] [--moved <file name|same>] [file ...]

Options:
	--help, -h			Show help
	--version, -v		Show sort-issue-template version
	--jsonnet, -j		Jsonnet file path
	--recursive, -r		If this is set, tfmv finds files recursively
	--moved, -m			The destination file name. If this is "same", the file is same with the resource`
