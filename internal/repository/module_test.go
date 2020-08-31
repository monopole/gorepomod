package repository_test

import (
	"testing"

	"github.com/monopole/gorepomod/internal/repository"
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

func TestComputeDepth(t *testing.T) {
	repo, _ := repository.NewRepo("gh.com/hoser")
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
		m := repository.NewModule(repo, tc.goMod)
		d := m.Depth()
		if d != tc.expectedDepth {
			t.Fatalf(
				"%s: %s, expected %d, got %d",
				n, m.InRepoPath(), tc.expectedDepth, d)
		}
	}
}
