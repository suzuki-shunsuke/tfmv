package apply

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

func (a *Applier) writeMovedBlock(block *types.Block, dest, movedFile string) error {
	if block.IsData() {
		return nil
	}
	file, err := a.fs.OpenFile(movedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) //nolint:mnd
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer file.Close()
	content := fmt.Sprintf(`moved {
  from = %s.%s
  to   = %s.%s
}
`, block.ResourceType, block.Name, block.ResourceType, dest)
	if !block.IsResource() {
		content = fmt.Sprintf(`moved {
  from = module.%s
  to   = module.%s
}
`, block.Name, dest)
	}
	if f, err := afero.Exists(a.fs, movedFile); err == nil && f {
		content = "\n" + content
	}
	fmt.Fprint(file, content)
	return nil
}
