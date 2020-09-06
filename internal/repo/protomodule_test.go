package repo

import (
	"testing"

	"github.com/monopole/gorepomod/internal/ifc"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

const irrelevantName = ifc.ModuleShortName("whatever")

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
						Path:    "gh.com/hoser/garage",
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
						Path:    "gh.com/hoser/fruit/yellow/banana",
						Version: "v2.3.4",
					},
				},
			},
		},
	}
	for n, tc := range testCases {
		m := protoModule{pathToGoMod: "irrelevant", mf: tc.modFile}
		actual := m.ShortName("gh.com/hoser")
		if actual != tc.name {
			t.Errorf(
				"%s: expected %s, got %s", n, tc.name, actual)
		}
	}
}
