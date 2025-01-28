package controller

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/suzuki-shunsuke/tfmv/pkg/apply"
	"github.com/suzuki-shunsuke/tfmv/pkg/plan"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

func (c *Controller) Run(logE *logrus.Entry, input *types.Input) error {
	planner := plan.NewPlanner(c.fs)
	dirs, err := planner.Plan(logE, input)
	if err != nil {
		return fmt.Errorf("plan changes: %w", err)
	}

	if err := c.summarize(dirs); err != nil {
		logerr.WithError(logE, err).Warn("output changed summary")
	}

	applier := apply.New(c.fs, c.stderr)
	if err := applier.Apply(logE, input, dirs); err != nil {
		return fmt.Errorf("apply changes: %w", err)
	}
	return nil
}

// summarize outputs a summary of changes as JSON to stdout.
func (c *Controller) summarize(dirs map[string]*types.Dir) error {
	summary := &Summary{}
	summary.FromDirs(dirs)
	encoder := json.NewEncoder(c.stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(summary); err != nil {
		return fmt.Errorf("encode a summary as JSON: %w", err)
	}
	return nil
}
