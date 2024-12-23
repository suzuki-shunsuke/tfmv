package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/suzuki-shunsuke/tfmv/pkg/cli"
	"github.com/suzuki-shunsuke/tfmv/pkg/log"
)

var (
	version = ""
	commit  = "" //nolint:gochecknoglobals
	date    = "" //nolint:gochecknoglobals
)

func main() {
	logE := log.New(version)
	if err := core(logE); err != nil {
		logerr.WithError(logE, err).Fatal("tfmv failed")
	}
}

func core(logE *logrus.Entry) error {
	runner := cli.Runner{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		LDFlags: &cli.LDFlags{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
		LogE: logE,
	}
	return runner.Run() //nolint:wrapcheck
}
