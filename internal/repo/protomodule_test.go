package repo

import (
	"testing"

	"github.com/monopole/gorepomod/internal/ifc"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func TestShortName(t *testing.T) {
	var testCases = map[string]struct {
		name    ifc.ModuleShortName
		modFile *modfile.File
	}{
		"one": {
			name:    ifc.ModuleShortName("garage"),
			modFile: &modfile.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path:    "gh.com/micheal/garage",
						Version: "v2.3.4",
					},
				},
			},
		},
		"three": {
			name:    ifc.ModuleShortName("fruit/yellow/banana"),
			modFile: &modfile.File{
				Module: &modfile.Module{
					Mod: module.Version{
						Path:    "gh.com/micheal/fruit/yellow/banana",
						Version: "v2.3.4",
					},
				},
			},
		},
	}
	for n, tc := range testCases {
		m := protoModule{pathToGoMod: "irrelevant", mf: tc.modFile}
		actual := m.ShortName("gh.com/micheal")
		if actual != tc.name {
			t.Errorf(
				"%s: expected %s, got %s", n, tc.name, actual)
		}
	}
}
