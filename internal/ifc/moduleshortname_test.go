package ifc_test

import (
	"testing"

	"github.com/monopole/gorepomod/internal/ifc"
)

func TestDepth(t *testing.T) {
	var testCases = map[string]struct {
		path          string
		expectedDepth int
	}{
		"zero": {
			path:       "{top}",
			expectedDepth: 0,
		},
		"one": {
			path:       "one",
			expectedDepth: 1,
		},
		"three": {
			path:       "one/two/three",
			expectedDepth: 3,
		},
	}
	for n, tc := range testCases {
		m := ifc.ModuleShortName(tc.path)
		d := m.Depth()
		if d != tc.expectedDepth {
			t.Fatalf(
				"%s: %s, expected %d, got %d",
				n, tc.path, tc.expectedDepth, d)
		}
	}
}
