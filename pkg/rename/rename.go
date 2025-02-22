package rename

import (
	"errors"
	"fmt"
	"strings"

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
	RenameAddress(block *types.Block) (string, error)
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

// Rename returns true if the block is renamed.
func Rename(renamer Renamer, block *types.Block, typ string) (bool, error) {
	switch typ {
	case "name":
		return renameName(renamer, block)
	case "type":
		return renameResourceType(renamer, block)
	case "address":
		return renameResourceType(renamer, block)
	default:
		return false, fmt.Errorf("unsupported type: %s", typ)
	}
}

func renameResourceType(renamer Renamer, block *types.Block) (bool, error) {
	n, err := renamer.RenameResourceType(block)
	if err != nil {
		return false, fmt.Errorf("get a new name: %w", err)
	}
	if n == "" || n == block.ResourceType {
		return false, nil
	}
	if !hclsyntax.ValidIdentifier(n) {
		return false, logerr.WithFields(errors.New("the new resource type is an invalid HCL identifier"), logrus.Fields{ //nolint:wrapcheck
			"address":           block.TFAddress,
			"new_resource_type": n,
		})
	}
	block.SetNewResourceType(n)
	return true, nil
}

func renameName(renamer Renamer, block *types.Block) (bool, error) {
	newName, err := renamer.RenameName(block)
	if err != nil {
		return false, fmt.Errorf("get a new name: %w", err)
	}
	if newName == "" || newName == block.Name {
		return false, nil
	}
	if !hclsyntax.ValidIdentifier(newName) {
		return false, logerr.WithFields(errors.New("the new name is an invalid HCL identifier"), logrus.Fields{ //nolint:wrapcheck
			"address":  block.TFAddress,
			"new_name": newName,
		})
	}
	block.SetNewName(newName)
	return true, nil
}

func renameAddress(renamer Renamer, block *types.Block) (bool, error) {
	newAddress, err := renamer.RenameAddress(block)
	if err != nil {
		return false, fmt.Errorf("get a new name: %w", err)
	}
	if newAddress == "" || newAddress == block.TFAddress {
		return false, nil
	}
	arr := strings.Split(newAddress, ".")
	for _, a := range arr {
		if !hclsyntax.ValidIdentifier(a) {
			return false, logerr.WithFields(errors.New("the new address is an invalid HCL identifier"), logrus.Fields{ //nolint:wrapcheck
				"address":     block.TFAddress,
				"new_address": newAddress,
			})
		}
	}
	if arr[0] != block.BlockType {
		return false, logerr.WithFields(errors.New("block type can't be changed"), logrus.Fields{ //nolint:wrapcheck
			"address":     block.TFAddress,
			"new_address": newAddress,
		})
	}
	if len(strings.Split(block.TFAddress, ".")) != len(arr) {
		return false, logerr.WithFields(errors.New("invalid change"), logrus.Fields{ //nolint:wrapcheck
			"address":     block.TFAddress,
			"new_address": newAddress,
		})
	}
	switch len(arr) {
	case 2: //nolint:mnd
		// module "foo"
		block.SetNewName(arr[1])
		return true, nil
	case 3: //nolint:mnd
		// resource "null_resource" "foo"
		block.SetNewAddress(arr[1], arr[2])
		return true, nil
	}
	return true, nil
}
