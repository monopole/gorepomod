package ifc

import (
	"github.com/monopole/gorepomod/internal/semver"
)

type ModFunc func(LaModule) error

const GoModFile = "go.mod"

type LaRepository interface {
	ImportPath() string
	AbsPath() string
	Apply(f ModFunc) error
	FindModule(ModuleShortName) LaModule
}

type LaModule interface {
	ShortName() ModuleShortName
	SrcRelativePath() string
	AbsPath() string
	Version() semver.SemVer
	DependsOn(LaModule) (bool, semver.SemVer)
}
