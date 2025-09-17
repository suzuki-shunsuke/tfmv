package types //nolint:revive

import "regexp"

type Input struct {
	// Jsonnet is a jsonnet option.
	Jsonnet string
	// MovedFile is -moved option.
	MovedFile string
	// Replace is a replace option.
	Replace string
	// Regexp is a regexp option.
	Regexp string
	// Include is an include option.
	Include *regexp.Regexp
	// Exclude is an exclude option.
	Exclude *regexp.Regexp
	// Args is a list of arguments.
	Args []string
	// Recursive is a recursive option.
	Recursive bool
	// DryRun is a dry-run option.
	DryRun bool
}

// Dir represents a Terraform Module directory.
type Dir struct {
	// Path is a directory path.
	Path string
	// Files a list of file paths (not file names).
	Files []string
	// Blocks is a list of renamed Terraform blocks.
	Blocks []*Block
}
