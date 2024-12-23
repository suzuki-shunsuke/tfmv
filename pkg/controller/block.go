package controller

import (
	"fmt"
	"regexp"
)

const (
	wordResource = "resource"
	wordData     = "data"
	wordModule   = "module"
)

type Block struct {
	File          string         `json:"file"`
	BlockType     string         `json:"block_type"`
	ResourceType  string         `json:"resource_type"`
	Name          string         `json:"name"`
	NewName       string         `json:"-"`
	MovedFile     string         `json:"-"`
	Regexp        *regexp.Regexp `json:"-"`
	TFAddress     string         `json:"-"`
	HCLAddress    string         `json:"-"`
	NewTFAddress  string         `json:"-"`
	NewHCLAddress string         `json:"-"`
}

func isResource(blockType string) bool {
	return blockType == wordResource
}

func (b *Block) IsResource() bool {
	return isResource(b.BlockType)
}

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

func (b *Block) SetNewName(newName string) {
	b.NewName = newName
	b.NewHCLAddress = hclAddress(b.BlockType, b.ResourceType, newName)
	b.NewTFAddress = tfAddress(b.BlockType, b.ResourceType, newName)
}

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

func (b *Block) Fix(body string) string {
	return b.Regexp.ReplaceAllString(body, b.NewTFAddress)
}
