package repository

import (
	"fmt"
	"strconv"
	"strings"
)

// SemanticVersion is the semantic version per https://semver.org
type SemanticVersion struct {
	major int
	minor int
	patch int
}

func NewVersion(major, minor, patch int) *SemanticVersion {
	return &SemanticVersion{
		major: major,
		minor: minor,
		patch: patch,
	}
}

func ParseVersion(raw string) (*SemanticVersion, error) {
	if len(raw) < 6 {
		// e.g. minimal length is 6, e.g. "v1.2.3"
		return nil, fmt.Errorf("%q too short to be a version", raw)
	}
	if raw[0] != 'v' {
		return nil, fmt.Errorf("%q must start with v", raw)
	}
	fields := strings.SplitN(raw[1:], ".", -1)
	if len(fields) < 3 {
		return nil, fmt.Errorf("%q doesn't have the form v1.2.3", raw)
	}
	n := make([]int, 3)
	for i := 0; i < 3; i++ {
		var err error
		n[i], err = strconv.Atoi(fields[i])
		if err != nil {
			return nil, err
		}
	}
	return NewVersion(n[0], n[1], n[2]), nil
}

func (v *SemanticVersion) BumpMajor() *SemanticVersion {
	return NewVersion(v.major+1, 0, 0)
}

func (v *SemanticVersion) BumpMinor() *SemanticVersion {
	return NewVersion(v.major, v.minor+1, 0)
}

func (v *SemanticVersion) BumpPatch() *SemanticVersion {
	return NewVersion(v.major, v.minor, v.patch+1)
}

func (v *SemanticVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v *SemanticVersion) Equals(o *SemanticVersion) bool {
	return v.major == o.major && v.minor == o.minor && v.patch == o.patch
}

func (v *SemanticVersion) GreaterThan(o *SemanticVersion) bool {
	return v.major > o.major ||
		(v.major == o.major && v.minor > o.minor) ||
		(v.major == o.major && v.minor == o.minor && v.patch > o.patch)
}
