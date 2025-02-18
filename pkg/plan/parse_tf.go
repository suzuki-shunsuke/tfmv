package plan

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

func parse(src []byte, filePath string, include, exclude *regexp.Regexp) ([]*types.Block, error) {
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
		b, err := parseBlock(filePath, block, include, exclude)
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

func parseBlock(filePath string, block *hclsyntax.Block, include, exclude *regexp.Regexp) (*types.Block, error) {
	if _, ok := types.Types()[block.Type]; !ok {
		return nil, nil //nolint:nilnil
	}
	b := &types.Block{
		File:      filePath,
		BlockType: block.Type,
	}
	switch len(block.Labels) {
	case 1:
		b.Name = block.Labels[0]
	case 2: //nolint:mnd
		b.ResourceType = block.Labels[0]
		b.Name = block.Labels[1]
	default:
		return nil, nil //nolint:nilnil
	}
	if err := b.Init(); err != nil {
		return nil, fmt.Errorf("initialize block attributes: %w", err)
	}
	if exclude != nil && exclude.MatchString(b.TFAddress) {
		return nil, nil //nolint:nilnil
	}
	if include != nil && !include.MatchString(b.TFAddress) {
		return nil, nil //nolint:nilnil
	}
	return b, nil
}
