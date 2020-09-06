package ifc

import "github.com/monopole/gorepomod/internal/semver"

type ModFunc func(LaModule) error

const GoModFile = "go.mod"

type LaRepository interface {
	ImportPath() string
	AbsPath() string
	Apply(f ModFunc) error
	FindModuleByRelPath(string) LaModule
}

type LaModule interface {
	InRepoPath() string
	SrcRelativePath() string
	AbsPath() string
	Depth() int
	Report()
	Version() *semver.SemVer
	DependsOn(LaModule) (bool, *semver.SemVer)
}
