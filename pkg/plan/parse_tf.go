package plan

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

func parse(src []byte, filePath string, input *types.Input) ([]*types.Block, error) {
	file, diags := hclsyntax.ParseConfig(src, filePath, hcl.Pos{Byte: 0, Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, diags
	}
	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil, errors.New("convert file body to body type")
	}
	blocks := make([]*types.Block, 0, len(body.Blocks))
	for _, block := range body.Blocks {
		b, err := parseBlock(filePath, block, input)
		if err != nil {
			return nil, err
		}
		if b == nil {
			continue
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

func parseBlock(filePath string, block *hclsyntax.Block, input *types.Input) (*types.Block, error) {
	if excludeBefore(block, input) {
		return nil, nil //nolint:nilnil
	}
	b := &types.Block{
		File:      filePath,
		BlockType: block.Type,
	}
	switch len(block.Labels) {
	case 1:
		// module "foo"
		b.Name = block.Labels[0]
	case 2: //nolint:mnd
		// resource "null_resource" "foo"
		b.ResourceType = block.Labels[0]
		b.Name = block.Labels[1]
	default:
		return nil, nil //nolint:nilnil
	}
	if err := b.Init(); err != nil {
		return nil, fmt.Errorf("initialize block attributes: %w", err)
	}
	if excludeAfter(b, input) {
		return nil, nil //nolint:nilnil
	}
	return b, nil
}

// excludeBefore returns true if the block should be excluded.
func excludeBefore(block *hclsyntax.Block, input *types.Input) bool {
	if _, ok := types.Types()[block.Type]; !ok {
		return true
	}
	// If --type is "type", module blocks are excluded.
	if input.Type == "type" && block.Type == "module" {
		return true
	}
	return false
}

// excludeAfter returns true if the block should be excluded.
func excludeAfter(b *types.Block, input *types.Input) bool {
	if input.Exclude != nil && input.Exclude.MatchString(b.TFAddress) {
		return true
	}
	if input.Include != nil && !input.Include.MatchString(b.TFAddress) {
		return true
	}
	return false
}
