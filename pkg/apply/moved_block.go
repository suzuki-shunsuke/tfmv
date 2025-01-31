package apply

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

var filePermission os.FileMode = 0o644 //nolint:gochecknoglobals

func (a *Applier) writeMovedBlock(block *types.Block, movedFile string) error {
	if block.IsData() {
		return nil
	}

	content := fmt.Sprintf(`moved {
  from = %s
  to   = %s
}
`, block.TFAddress, block.NewTFAddress)

	f, err := a.fs.Stat(movedFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("check a file exists: %w", err)
		}
		// create a file
		if err := afero.WriteFile(a.fs, movedFile, []byte(content), filePermission); err != nil {
			return fmt.Errorf("create a moved block file: %w", err)
		}
		return nil
	}
	// update a file
	file, err := a.fs.OpenFile(movedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, f.Mode())
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
	}
	defer file.Close()
	fmt.Fprint(file, "\n"+content)
	return nil
}
