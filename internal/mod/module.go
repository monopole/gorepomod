package mod

import (
	"path/filepath"

	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/semver"
	"golang.org/x/mod/modfile"
)

// Module is an immutable representation of a Go module.
type Module struct {
	repo      ifc.LaRepository
	shortName ifc.ModuleShortName
	mf        *modfile.File
	depth     int
	v         semver.SemVer
}

func NewModule(
	repo ifc.LaRepository,
	shortName ifc.ModuleShortName,
	mf *modfile.File,
	v semver.SemVer) *Module {
	return &Module{
		repo:      repo,
		shortName: shortName,
		mf:        mf,
		v:         v,
	}
}

func (m *Module) GitRepo() ifc.LaRepository {
	return m.repo
}

func (m *Module) Version() semver.SemVer {
	return m.v
}

// AbsPath is the absolute path to the module's go.mod file.
func (m *Module) AbsPath() string {
	return filepath.Join(m.repo.AbsPath(), string(m.ShortName()))
}

// SrcRelativePath is the relative path below the Go src root.
func (m *Module) SrcRelativePath() string {
	return filepath.Join(m.repo.ImportPath(), string(m.ShortName()))
}

func (m *Module) ShortName() ifc.ModuleShortName {
	return m.shortName
}

func (m *Module) DependsOn(target ifc.LaModule) (bool, semver.SemVer) {
	for _, r := range m.mf.Require {
		if r.Mod.Path == target.SrcRelativePath() {
			v, err := semver.Parse(r.Mod.Version)
			if err != nil {
				panic(err)
			}
			return true, v
		}
	}
	return false, semver.Zero()
}
