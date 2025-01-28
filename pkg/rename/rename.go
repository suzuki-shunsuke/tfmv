package rename

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/logrus-error/logerr"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

// Renamer is an interface to rename a block address.
type Renamer interface {
	RenameName(block *types.Block) (string, error)
	RenameResourceType(block *types.Block) (string, error)
}

// New creates a Renamer.
func New(logE *logrus.Entry, fs afero.Fs, input *types.Input) (Renamer, error) {
	if input.Replace != "" {
		return NewReplaceRenamer(input.Replace)
	}
	if input.Jsonnet != "" {
		return NewJsonnetRenamer(logE, fs, input.Jsonnet)
	}
	if input.Regexp != "" {
		return NewRegexpRenamer(input.Regexp)
	}
	return nil, errors.New("one of --jsonnet or --replace or --regexp must be specified")
}

func Rename(renamer Renamer, block *types.Block, typ string) error {
	switch typ {
	case "name":
		return renameName(renamer, block)
	case "type":
		return renameResourceType(renamer, block)
	default:
		return fmt.Errorf("unsupported type: %s", typ)
	}
}

func renameResourceType(renamer Renamer, block *types.Block) error {
	n, err := renamer.RenameResourceType(block)
	if err != nil {
		return fmt.Errorf("get a new name: %w", err)
	}
	if n == "" || n == block.ResourceType {
		return nil
	}
	if !hclsyntax.ValidIdentifier(n) {
		return logerr.WithFields(errors.New("the new resource type is an invalid HCL identifier"), logrus.Fields{ //nolint:wrapcheck
			"address":           block.TFAddress,
			"new_resource_type": n,
		})
	}
	block.SetNewResourceType(n)
	return nil
}

func renameName(renamer Renamer, block *types.Block) error {
	newName, err := renamer.RenameName(block)
	if err != nil {
		return fmt.Errorf("get a new name: %w", err)
	}
	if newName == "" || newName == block.Name {
		return nil
	}
	if !hclsyntax.ValidIdentifier(newName) {
		return logerr.WithFields(errors.New("the new name is an invalid HCL identifier"), logrus.Fields{ //nolint:wrapcheck
			"address":  block.TFAddress,
			"new_name": newName,
		})
	}
	block.SetNewName(newName)
	return nil
}
