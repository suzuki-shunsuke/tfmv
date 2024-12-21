package controller

import (
	"fmt"
	"os"
)

func (c *Controller) writeMovedBlock(block *Block, dest, movedFile string) error {
	file, err := c.fs.OpenFile(movedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:mnd
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer file.Close()
	fmt.Fprintf(file, `moved {
  from = %s.%s
  to   = %s.%s
}
`, block.ResourceType, block.Name, block.ResourceType, dest)
	return nil
}
