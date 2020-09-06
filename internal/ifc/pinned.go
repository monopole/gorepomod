package ifc

import "github.com/monopole/gorepomod/internal/semver"

type Pinned struct {
	M LaModule
	V semver.SemVer
}

func (p Pinned) String() string {
	return string(p.M.ShortName()) + ":" + p.V.String()
}
