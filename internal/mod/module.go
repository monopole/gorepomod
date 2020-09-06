package mod

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/semver"
	"golang.org/x/mod/modfile"
)

type Module struct {
	repo       ifc.LaRepository
	inRepoPath string
	pm         *modfile.File
	v          *semver.SemVer
}

func NewModule(
	repo ifc.LaRepository,
	inRepoPath string,
	f *modfile.File,
	v *semver.SemVer) *Module {
	return &Module{
		repo:       repo,
		inRepoPath: inRepoPath,
		pm:         f,
		v:          v,
	}
}

func (m *Module) GitRepo() ifc.LaRepository {
	return m.repo
}

func (m *Module) Version() *semver.SemVer {
	return m.v
}

// AbsPath is the absolute path to the module's go.mod file.
func (m *Module) AbsPath() string {
	return filepath.Join(m.repo.AbsPath(), m.InRepoPath())
}

// SrcRelativePath is the relative path below the Go src root.
func (m *Module) SrcRelativePath() string {
	return filepath.Join(m.repo.ImportPath(), m.InRepoPath())
}

func (m *Module) InRepoPath() string {
	return m.inRepoPath
}

func (m *Module) Report() {
	fmt.Printf("           AbsPath: %s\n", m.AbsPath())
	fmt.Printf("   SrcRelativePath: %s\n", m.SrcRelativePath())
	fmt.Printf("        InRepoPath: %s\n", m.InRepoPath())
	fmt.Printf("           Version: %s\n", m.v.String())
	fmt.Println()
}

func (m *Module) Depth() int {
	return strings.Count(m.InRepoPath(), string(filepath.Separator)) + 1
}

func (m *Module) DependsOn(target ifc.LaModule) (bool, *semver.SemVer) {
	for _, r := range m.pm.Require {
		if r.Mod.Path == target.SrcRelativePath() {
			v, err := semver.Parse(r.Mod.Version)
			if err != nil {
				panic(err)
			}
			return true, v
		}
	}
	return false, nil
}
