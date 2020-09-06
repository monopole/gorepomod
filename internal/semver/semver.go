package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// SemVer is the semantic version per https://semver.org
type SemVer struct {
	major int
	minor int
	patch int
}

func New(major, minor, patch int) *SemVer {
	return &SemVer{
		major: major,
		minor: minor,
		patch: patch,
	}
}

// Versions implements sort.Interface based on the Age field.
type Versions []*SemVer

func (v Versions) Len() int           { return len(v) }
func (v Versions) Less(i, j int) bool { return v[j].LessThan(v[i]) }
func (v Versions) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

func Parse(raw string) (*SemVer, error) {
	if len(raw) < 6 {
		// e.g. minimal length is 6, e.g. "v1.2.3"
		return nil, fmt.Errorf("%q too short to be a version", raw)
	}
	if raw[0] != 'v' {
		return nil, fmt.Errorf("%q must start with v", raw)
	}
	fields := strings.Split(raw[1:], ".")
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
	return New(n[0], n[1], n[2]), nil
}

func (v *SemVer) BumpMajor() *SemVer {
	return New(v.major+1, 0, 0)
}

func (v *SemVer) BumpMinor() *SemVer {
	return New(v.major, v.minor+1, 0)
}

func (v *SemVer) BumpPatch() *SemVer {
	return New(v.major, v.minor, v.patch+1)
}

func (v *SemVer) String() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("v%d.%d.%d", v.major, v.minor, v.patch)
}

func (v *SemVer) Equals(o *SemVer) bool {
	return v.major == o.major && v.minor == o.minor && v.patch == o.patch
}

func (v *SemVer) LessThan(o *SemVer) bool {
	return v.major < o.major ||
			(v.major == o.major && v.minor < o.minor) ||
			(v.major == o.major && v.minor == o.minor && v.patch < o.patch)
}
