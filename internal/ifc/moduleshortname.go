package ifc

import (
	"path/filepath"
	"strings"
)

// ModuleShortName is the in-repo path to the go.mod file, the unique
// in-repo name of the module.
// E.g. "" (empty), "kyaml", "cmd/config", "plugin/example/whatever".
// It's the name used to tag the repo at a particular module version.
type ModuleShortName string

// Never used in a tag.
const TopModule = ModuleShortName("{top}")

func (m ModuleShortName) Depth() int {
	if m == TopModule {
		return 0
	}
	return strings.Count(string(m), string(filepath.Separator)) + 1
}
