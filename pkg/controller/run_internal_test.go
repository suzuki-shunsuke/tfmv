package controller_test

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/controller"
)

func TestController_Run(t *testing.T) { //nolint:funlen
	t.Parallel()
	tests := []struct {
		name   string
		files  map[string]string
		stdout io.Writer
		stderr io.Writer
		input  *controller.Input
		isErr  bool
	}{
		{
			name: "no changed file",
			files: map[string]string{
				"main.tf": `resource "null_resource" "example_1" {}
`,
			},
			stdout: &bytes.Buffer{},
			stderr: &bytes.Buffer{},
			input: &controller.Input{
				Args:    []string{"main.tf"},
				Replace: "-/_",
				DryRun:  true,
			},
		},
		{
			name: "replace",
			files: map[string]string{
				"testdata/main.tf": `resource "null_resource" "example-1" {}
`,
			},
			stdout: &bytes.Buffer{},
			stderr: &bytes.Buffer{},
			input: &controller.Input{
				Args:    []string{"testdata/main.tf"},
				Replace: "-/_",
				DryRun:  true,
			},
		},
		{
			name: "regexp",
			files: map[string]string{
				"testdata/main.tf": `resource "null_resource" "example-1" {}
`,
			},
			stdout: &bytes.Buffer{},
			stderr: &bytes.Buffer{},
			input: &controller.Input{
				Args:   []string{"testdata/main.tf"},
				Regexp: "^example-/test-",
				DryRun: true,
			},
		},
		{
			name: "jsonnet",
			files: map[string]string{
				"testdata/main.tf": `resource "null_resource" "example-1" {}
`,
				"main.jsonnet": `std.native("strings.Replace")(std.extVar('input').name, "-", "_", -1)[0]
`,
			},
			stdout: &bytes.Buffer{},
			stderr: &bytes.Buffer{},
			input: &controller.Input{
				Args:   []string{"testdata/main.tf"},
				File:   "main.jsonnet",
				DryRun: true,
			},
		},
		{
			name: "no renamer",
			files: map[string]string{
				"testdata/main.tf": `resource "null_resource" "example-1" {}
`,
			},
			stdout: &bytes.Buffer{},
			stderr: &bytes.Buffer{},
			input:  &controller.Input{},
			isErr:  true,
		},
		{
			name:   "no file is found",
			files:  map[string]string{},
			stdout: &bytes.Buffer{},
			stderr: &bytes.Buffer{},
			input: &controller.Input{
				Replace: "-/_",
			},
		},
	}
	logE := logrus.NewEntry(logrus.New())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := afero.NewMemMapFs()
			for path, content := range tt.files {
				if err := fs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := afero.WriteFile(fs, path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			ctrl := &controller.Controller{}
			ctrl.Init(fs, tt.stdout, tt.stderr)
			if err := ctrl.Run(logE, tt.input); err != nil {
				if tt.isErr {
					return
				}
				t.Fatal(err)
			}
			if tt.isErr {
				t.Fatal("error is expected")
			}
		})
	}
}
