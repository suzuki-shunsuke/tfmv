package controller

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/lintnet/go-jsonnet-native-functions/pkg/net/url"
	"github.com/lintnet/go-jsonnet-native-functions/pkg/path"
	"github.com/lintnet/go-jsonnet-native-functions/pkg/path/filepath"
	"github.com/lintnet/go-jsonnet-native-functions/pkg/regexp"
	"github.com/lintnet/go-jsonnet-native-functions/pkg/strings"
)

func (c *Controller) evaluate(block *Block, ja ast.Node) (string, error) {
	b, err := json.Marshal(block)
	if err != nil {
		return "", fmt.Errorf("marshal a block: %w", err)
	}
	vm := NewVM(string(b))
	result, err := vm.Evaluate(ja)
	if err != nil {
		return "", fmt.Errorf("evaluate Jsonnet: %w", err)
	}
	var dest string
	if err := json.Unmarshal([]byte(result), &dest); err != nil {
		return "", fmt.Errorf("unmarshal as a JSON: %w", err)
	}
	return dest, nil
}

func SetNativeFunctions(vm *jsonnet.VM) {
	m := map[string]func(string) *jsonnet.NativeFunction{
		"filepath.Base":        filepath.Base,
		"path.Base":            path.Base,
		"path.Clean":           path.Clean,
		"path.Dir":             path.Dir,
		"path.Ext":             path.Ext,
		"path.IsAbs":           path.IsAbs,
		"path.Match":           path.Match,
		"path.Split":           path.Split,
		"regexp.MatchString":   regexp.MatchString,
		"strings.Contains":     strings.Contains,
		"strings.ContainsAny":  strings.ContainsAny,
		"strings.Count":        strings.Count,
		"strings.Cut":          strings.Cut,
		"strings.CutPrefix":    strings.CutPrefix,
		"strings.CutSuffix":    strings.CutSuffix,
		"strings.EqualFold":    strings.EqualFold,
		"strings.Fields":       strings.Fields,
		"strings.LastIndex":    strings.LastIndex,
		"strings.LastIndexAny": strings.LastIndexAny,
		"strings.Repeat":       strings.Repeat,
		"strings.Replace":      strings.Replace,
		"strings.TrimPrefix":   strings.TrimPrefix,
		"strings.TrimSpace":    strings.TrimSpace, //nolint:staticcheck
		"url.Parse":            url.Parse,
	}
	for k, v := range m {
		vm.NativeFunction(v(k))
	}
}

func NewVM(input string) *jsonnet.VM {
	vm := jsonnet.MakeVM()
	vm.ExtCode("input", input)
	SetNativeFunctions(vm)
	return vm
}
