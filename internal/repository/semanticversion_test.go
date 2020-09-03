package repository

import (
	"testing"
)

func TestParse(t *testing.T) {
	var testCases = map[string]struct {
		raw    string
		v      *SemanticVersion
		errMsg string
	}{
		"one": {
			raw:    "v1.2.3",
			v:      &SemanticVersion{major: 1, minor: 2, patch: 3},
			errMsg: "",
		},
		"two": {
			raw:    "v2.0.9999",
			v:      &SemanticVersion{major: 2, minor: 0, patch: 9999},
			errMsg: "",
		},
		"three": {
			raw:    "heyho",
			v:      nil,
			errMsg: "\"heyho\" too short to be a version",
		},
	}
	for n, tc := range testCases {
		v, err := ParseVersion(tc.raw)
		if err == nil {
			if tc.errMsg != "" {
				t.Errorf(
					"%s: no error, but expected err %q", n, tc.errMsg)
			}
			if !v.Equals(tc.v) {
				t.Errorf(
					"%s: expected %v, got %v", n, tc.v, v)
			}
		} else {
			if tc.errMsg == "" {
				t.Errorf(
					"%s: unexpected error %v", n, err)
			} else {
				if tc.errMsg != err.Error() {
					t.Errorf(
						"%s: expected err msg %q, but got %q",
						n, tc.errMsg, err.Error())
				}
			}
		}
	}
}

func TestGreaterThan(t *testing.T) {
	var testCases = map[string]struct {
		v1       *SemanticVersion
		v2       *SemanticVersion
		expected bool
	}{
		"one": {
			v1:       &SemanticVersion{major: 2, minor: 2, patch: 3},
			v2:       &SemanticVersion{major: 1, minor: 2, patch: 3},
			expected: true,
		},
		"two": {
			v1:       &SemanticVersion{major: 1, minor: 3, patch: 3},
			v2:       &SemanticVersion{major: 1, minor: 2, patch: 3},
			expected: true,
		},
		"three": {
			v1:       &SemanticVersion{major: 1, minor: 2, patch: 4},
			v2:       &SemanticVersion{major: 1, minor: 2, patch: 3},
			expected: true,
		},
		"eq": {
			v1:       &SemanticVersion{major: 2, minor: 2, patch: 3},
			v2:       &SemanticVersion{major: 2, minor: 2, patch: 3},
			expected: false,
		},
	}
	for n, tc := range testCases {
		actual := tc.v1.GreaterThan(tc.v2)
		if actual != tc.expected {
			t.Errorf(
				"%s: expected %v, got %v for %s GreaterThan %s",
				n, tc.expected, actual, tc.v1.String(), tc.v2.String())
		}
	}
}
