package repository

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
)

const GoModFile = "go.mod"

type Module struct {
	repo *Repo
	pm   *modfile.File
}

func NewModule(repo *Repo, f *modfile.File) *Module {
	return &Module{
		repo: repo,
		pm:   f,
	}
}

// Represents the trailing version label in a module name.
// See https://blog.golang.org/v2-go-modules
var trailingVersionPattern = regexp.MustCompile("/v\\d+$")

func (m *Module) InRepoPath() string {
	p := m.FullPath()[len(m.repo.root)+1:]
	return trailingVersionPattern.ReplaceAllString(p, "")
}

func (m *Module) Repo() *Repo {
	return m.repo
}

func (m *Module) FullPath() string {
	return m.pm.Module.Mod.Path
}

func (m *Module) Report() {
	fmt.Printf("FullPath: %s\n", m.FullPath())
	fmt.Printf(" RelPath: %s\n", m.InRepoPath())
	fmt.Printf("   Depth: %d\n", m.Depth())
	fmt.Println()
}

func (m *Module) Depth() int {
	return strings.Count(m.InRepoPath(), string(filepath.Separator)) + 1
}

func (m *Module) DependsOn(target *Module) (bool, SemanticVersion) {
	for _, r := range m.pm.Require {
		if r.Mod.Path == target.FullPath() {
			return true, SemanticVersion(r.Mod.Version)
		}
	}
	return false, ""
}

func readGoModFile(path string) (*modfile.File, error) {
	mPath := filepath.Join(path, GoModFile)
	content, err := ioutil.ReadFile(mPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %v\n", mPath, err)
	}
	return modfile.Parse(mPath, content, nil)
}

func ParseIntoRepo(repo *Repo, path string) (*Module, error) {
	parsed, err := readGoModFile(path)
	if err != nil {
		return nil, fmt.Errorf("error parsing content from %q: %v\n", path, err)
	}
	m := NewModule(repo, parsed)

	// Sanity check 1
	if !strings.HasPrefix(m.FullPath(), repo.GetRoot()) {
		return nil, fmt.Errorf(
			"module %s doesn't start with the repository name %q",
			m.FullPath(), repo.GetRoot())
	}

	// Sanity check 2
	if !strings.HasSuffix(path, m.InRepoPath()) {
		return nil, fmt.Errorf(
			"in file %q, the module name %q doesn't match the file's path",
			filepath.Join(path, GoModFile), m.InRepoPath())
	}

	// Other checks could be
	// - make sure all the versions specified in the file match semver pattern.
	// - in fix command, run "git tag -l" and verify that the tag exists.
	return m, nil
}
