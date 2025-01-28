package rename

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
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/suzuki-shunsuke/tfmv/pkg/types"
)

type JsonnetRenamer struct {
	node ast.Node
}

func NewJsonnetRenamer(logE *logrus.Entry, fs afero.Fs, file string) (*JsonnetRenamer, error) {
	// read Jsonnet
	logE.Debug("reading a jsonnet file")
	b, err := afero.ReadFile(fs, file)
	if err != nil {
		return nil, fmt.Errorf("read a jsonnet file: %w", err)
	}
	// parse Jsonnet
	logE.Debug("parsing a jsonnet file")
	node, err := jsonnet.SnippetToAST(file, string(b))
	if err != nil {
		return nil, fmt.Errorf("parse a jsonnet file: %w", err)
	}
	return &JsonnetRenamer{node: node}, nil
}

func (j *JsonnetRenamer) Rename(block *types.Block) (string, error) {
	b, err := json.Marshal(block)
	if err != nil {
		return "", fmt.Errorf("marshal a block: %w", err)
	}
	vm := NewVM(string(b))
	result, err := vm.Evaluate(j.node)
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
