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
	"github.com/suzuki-shunsuke/tfmv/pkg/log"
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

func (r *Runner) Run(ctx context.Context) error {
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
	log.SetColor(flg.LogColor, r.LogE)
	log.SetLevel(flg.LogLevel, r.LogE)
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
		DryRun:    flg.DryRun,
		Args:      flg.Args,
		Replace:   flg.Replace,
	})
}

type Flag struct {
	Jsonnet   string
	Moved     string
	LogLevel  string
	LogColor  string
	Replace   string
	Args      []string
	Help      bool
	Version   bool
	Recursive bool
	DryRun    bool
}

func parseFlags(f *Flag) {
	flag.StringVarP(&f.Jsonnet, "jsonnet", "j", "", "Jsonnet file path")
	flag.StringVarP(&f.Moved, "moved", "m", "moved.tf", "The destination file name")
	flag.StringVarP(&f.Replace, "replace", "r", "", "Replace strings in block names. The format is <new>/<old>. e.g. -/_")
	flag.StringVar(&f.LogLevel, "log-level", "info", "The log level")
	flag.StringVar(&f.LogColor, "log-color", "auto", "The log color")
	flag.BoolVarP(&f.Help, "help", "h", false, "Show help")
	flag.BoolVarP(&f.Version, "version", "v", false, "Show version")
	flag.BoolVarP(&f.Recursive, "recursive", "R", false, "If this is set, tfmv finds files recursively")
	flag.BoolVar(&f.DryRun, "dry-run", false, "Dry Run")
	flag.Parse()
	f.Args = flag.Args()
}

const help = `tfmv - Rename Terraform resources and modules and generate moved blocks.
https://github.com/suzuki-shunsuke/tfmv

Usage:
	tfmv [--jsonnet <Jsonnet file path>] [--recursive] [--moved <file name|same>] [file ...]

Options:
	--help, -h       Show help
	--version, -v    Show sort-issue-template version
	--jsonnet, -j    Jsonnet file path
	--recursive, -R  If this is set, tfmv finds files recursively
	--replace, -r    Replace strings in block names. The format is <new>/<old>. e.g. -/_
	--dry-run        Dry Run
	--log-level      Log level
	--log-color      Log color. "auto", "always", "never" are available
	--moved, -m      The destination file name. If this is "same", the file is same with the resource`
