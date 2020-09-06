package mod_test

import (
	"testing"

	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/mod"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var (
	goMod1 = modfile.File{
		Module: &modfile.Module{
			Mod: module.Version{
				Path:    "gh.com/hoser/garage",
				Version: "v2.3.4",
			},
		},
	}
	goMod2 = modfile.File{
		Module: &modfile.Module{
			Mod: module.Version{
				Path:    "gh.com/hoser/fruit/yellow/banana",
				Version: "v2.3.4",
			},
		},
	}
)

type mockRepo struct{}

func (mr mockRepo) FindModuleByRelPath(s string) ifc.LaModule {
	panic("implement me")
}

func (mr mockRepo) Apply(f ifc.ModFunc) error {
	panic("implement me")
}

func (mr mockRepo) ImportPath() string {
	return "gh.com/hoser"
}

var _ ifc.LaRepository = mockRepo{}

func TestComputeDepth(t *testing.T) {
	var repo mockRepo
	type testData struct {
		path          string
		goMod         *modfile.File
		expectedDepth int
	}
	var testCases = map[string]testData{
		"one": {
			goMod:         &goMod1,
			expectedDepth: 1,
		},
		"three": {
			goMod:         &goMod2,
			expectedDepth: 3,
		},
	}
	for n, tc := range testCases {
		m := mod.NewModule(repo, "hey", tc.goMod)
		d := m.Depth()
		if d != tc.expectedDepth {
			t.Fatalf(
				"%s: %s, expected %d, got %d",
				n, m.InRepoPath(), tc.expectedDepth, d)
		}
	}
}
