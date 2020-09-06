package ifc

import (
	"path/filepath"
	"strings"
)

type ModuleShortName string

const TopModule = ModuleShortName("{top}")

func (m ModuleShortName) Depth() int {
	if m == TopModule {
		return 0
	}
	return strings.Count(string(m), string(filepath.Separator)) + 1
}
