package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	flag "github.com/spf13/pflag"
	"github.com/suzuki-shunsuke/tfmv/pkg/controller"
	"github.com/suzuki-shunsuke/tfmv/pkg/log"
)

const help = `tfmv - Rename Terraform resources, data sources, and modules and generate moved blocks.
https://github.com/suzuki-shunsuke/tfmv

Usage:
	tfmv [<options>] [file ...]

One of --jsonnet (-j), --replace (-r), or --regexp must be specified.

Options:
	--help, -h       Show help
	--version, -v    Show sort-issue-template version
	--replace, -r    Replace strings in block names. The format is <old>/<new>. e.g. -/_
	--jsonnet, -j    Jsonnet file path
	--regexp         Replace strings in block names by regular expression. The format is <regular expression>/<new>. e.g. '\bfoo\b/bar'
	--recursive, -R  If this is set, tfmv finds files recursively
	--include        A regular expression to filter resources. Only resources that match the regular expression are renamed
	--exclude        A regular expression to filter resources. Only resources that don't match the regular expression are renamed
	--dry-run        Dry Run
	--log-level      Log level
	--log-color      Log color. "auto", "always", "never" are available
	--moved, -m      A file name where moved blocks are written. If this is "same", the file is same with renamed resources`

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

func (r *Runner) Run() error {
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

	include, err := getRegexFilter(flg.Include)
	if err != nil {
		return fmt.Errorf("--include is an invalid regular expression: %w", err)
	}

	exclude, err := getRegexFilter(flg.Exclude)
	if err != nil {
		return fmt.Errorf("--exclude is an invalid regular expression: %w", err)
	}

	ctrl := &controller.Controller{}
	ctrl.Init(afero.NewOsFs(), r.Stdout, r.Stderr)
	return ctrl.Run(r.LogE, &controller.Input{ //nolint:wrapcheck
		File:      flg.Jsonnet,
		Dest:      flg.Moved,
		Recursive: flg.Recursive,
		DryRun:    flg.DryRun,
		Args:      flg.Args,
		Replace:   flg.Replace,
		Include:   include,
		Exclude:   exclude,
		Regexp:    flg.Regexp,
	})
}

func getRegexFilter(s string) (*regexp.Regexp, error) {
	if s == "" {
		return nil, nil //nolint:nilnil
	}
	return regexp.Compile(s) //nolint:wrapcheck
}

type Flag struct {
	Jsonnet   string
	Moved     string
	LogLevel  string
	LogColor  string
	Replace   string
	Regexp    string
	Include   string
	Exclude   string
	Args      []string
	Help      bool
	Version   bool
	Recursive bool
	DryRun    bool
}

func parseFlags(f *Flag) {
	flag.StringVarP(&f.Jsonnet, "jsonnet", "j", "", "Jsonnet file path")
	flag.StringVarP(&f.Moved, "moved", "m", "moved.tf", "The destination file name")
	flag.StringVarP(&f.Replace, "replace", "r", "", "Replace strings in block names. The format is <old>/<new>. e.g. -/_")
	flag.StringVar(&f.Regexp, "regexp", "", "Replace strings in block names by regular expression. The format is <regular expression>/<new>. e.g. '\bfoo\b/bar'")
	flag.StringVar(&f.Include, "include", "", "A regular expression to filter resources")
	flag.StringVar(&f.Exclude, "exclude", "", "A regular expression to filter resources")
	flag.StringVar(&f.LogLevel, "log-level", "info", "The log level")
	flag.StringVar(&f.LogColor, "log-color", "auto", "The log color")
	flag.BoolVarP(&f.Help, "help", "h", false, "Show help")
	flag.BoolVarP(&f.Version, "version", "v", false, "Show version")
	flag.BoolVarP(&f.Recursive, "recursive", "R", false, "If this is set, tfmv finds files recursively")
	flag.BoolVar(&f.DryRun, "dry-run", false, "Dry Run")
	flag.Parse()
	f.Args = flag.Args()
}
