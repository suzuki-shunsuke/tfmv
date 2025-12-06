package main

import (
	"os"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/tfmv/pkg/cli"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	if code := core(); code != 0 {
		os.Exit(code)
	}
}

func core() int {
	logger := slogutil.New(&slogutil.InputNew{
		Name:    "tfmv",
		Version: version,
	})
	runner := cli.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		LDFlags: &cli.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
		Logger: logger,
	}
	if err := runner.Run(); err != nil {
		slogerr.WithError(logger.Logger, err).Error("tfmv failed")
		return 1
	}
	return 0
}
