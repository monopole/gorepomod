package repository

import (
	"fmt"
	"regexp"
)

// SemanticVersion is the semantic version per https://semver.org
type SemanticVersion string

var semVerPattern = regexp.MustCompile("^v\\d+\\.\\d+\\.\\d+$")

func NewSemanticVersion(raw string) (SemanticVersion, error) {
	if semVerPattern.MatchString(raw) {
		return SemanticVersion(raw), nil
	}
	return SemanticVersion(raw), fmt.Errorf(
		"%q isn't a valid sematic version in the form v1.2.3", raw)
}

func (v SemanticVersion) String() string {
	return string(v)
}
