package types

import (
	"fmt"
	"regexp"
)

const (
	wordResource = "resource"
	wordData     = "data"
	wordModule   = "module"
)

// Block represents a Terraform resource, data, or module block.
type Block struct {
	// File is a file path
	File string `json:"file"`
	// BlockType is one of "resource", "data", or "module"
	BlockType string `json:"block_type"`
	// ResourceType is a resource type such as "aws_instance"
	ResourceType string `json:"resource_type"`
	// Name is a resource name.
	Name string `json:"name"`
	// NewName is a new resource name.
	NewName string `json:"-"`
	// MovedFile is a file path where moved blocks are written.
	MovedFile string `json:"-"`
	// Regexp is a regular expression to capture a resource reference.
	Regexp *regexp.Regexp `json:"-"`
	// TFAdress is a Terraform address such as "aws_instance.foo"
	TFAddress string `json:"-"`
	// HCLAddress is a HCL address such as "resource.aws_instance.foo"
	// hcledit uses this address rather than TFAddress.
	HCLAddress string `json:"-"`
	// NewTFAddress is a new Terraform address.
	NewTFAddress string `json:"-"`
	// NewHCLAddress is a new HCL address.
	NewHCLAddress string `json:"-"`
}

// isResource returns true if blockType is "resource".
func isResource(blockType string) bool {
	return blockType == wordResource
}

// isResource returns true if the block type is "resource".
func (b *Block) IsResource() bool {
	return isResource(b.BlockType)
}

func (b *Block) IsData() bool {
	return b.BlockType == wordData
}

// Types returns a map of block types.
func Types() map[string]struct{} {
	return map[string]struct{}{
		wordResource: {},
		wordData:     {},
		wordModule:   {},
	}
}

// hclAddress returns a HCL address.
func hclAddress(blockType, resourceType, name string) string {
	switch blockType {
	case wordResource:
		return fmt.Sprintf("resource.%s.%s", resourceType, name)
	case wordData:
		return fmt.Sprintf("data.%s.%s", resourceType, name)
	case wordModule:
		return "module." + name
	}
	return ""
}

// tfAddress returns a Terraform address.
func tfAddress(blockType, resourceType, name string) string {
	switch blockType {
	case wordResource:
		return fmt.Sprintf("%s.%s", resourceType, name)
	case wordData:
		return fmt.Sprintf("data.%s.%s", resourceType, name)
	case wordModule:
		return "module." + name
	}
	return ""
}

// Regestr returns a regular expression to capture a resource reference.
func (b *Block) Regstr() string {
	// A name must start with a letter or underscore and may contain only letters, digits, underscores, and dashes.
	switch b.BlockType {
	case wordResource:
		return fmt.Sprintf(`\b%s\.%s\b`, b.ResourceType, b.Name)
	case wordData:
		return fmt.Sprintf(`\bdata\.%s\.%s\b`, b.ResourceType, b.Name)
	case wordModule:
		return fmt.Sprintf(`\bmodule\.%s\b`, b.Name)
	}
	return ""
}

// SetNewName sets updates a new name, a new HCL address, and a new Terraform address.
func (b *Block) SetNewName(newName string) {
	b.NewName = newName
	b.NewHCLAddress = hclAddress(b.BlockType, b.ResourceType, newName)
	b.NewTFAddress = tfAddress(b.BlockType, b.ResourceType, newName)
}

// Init initializes a block attributes.
func (b *Block) Init() error {
	b.TFAddress = tfAddress(b.BlockType, b.ResourceType, b.Name)
	b.HCLAddress = hclAddress(b.BlockType, b.ResourceType, b.Name)
	reg, err := regexp.Compile(b.Regstr())
	if err != nil {
		return fmt.Errorf("compile a regular expression to capture a resource reference: %w", err)
	}
	b.Regexp = reg
	return nil
}

// Fix replaces resource references with a new Terraform address.
func (b *Block) Fix(body string) string {
	return b.Regexp.ReplaceAllString(body, b.NewTFAddress)
}
