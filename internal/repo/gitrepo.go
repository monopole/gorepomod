package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/mod"
	"github.com/monopole/gorepomod/internal/semver"
	"github.com/monopole/gorepomod/internal/utils"
)

const (
	dotGit           = ".git"
	srcPath          = "/src/"
	dotDir           = "."
	pathSep          = "/"
	rootedModulePath = "{top}"
)

type GitRepo struct {
	// srcRoot is the absolute path to the local Go src srcRoot,
	// the directory containing git repository clones.
	srcRoot string

	// importPath is the import path of repository,
	// e.g. github.com/kubernetes-sigs/kustomize
	// The directory {srcRoot}/{importPath} should contain a dotGit directory.
	// This directory might be a Go module, or contain directories
	// that are Go modules, or both.
	importPath string

	// modules is a list of Go modules found in the local repository.
	modules []ifc.LaModule

	// doIt, if true, allows modification to the repo
	doIt bool
}

func New(name string) (*GitRepo, error) {
	return &GitRepo{srcRoot: ".", importPath: name}, nil
}

func NewFromCwd() (*GitRepo, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if !utils.DirExists(filepath.Join(path, dotGit)) {
		return nil, fmt.Errorf("%q is not a git repo srcRoot", path)
	}
	index := strings.Index(path, srcPath)
	if index < 0 {
		return nil, fmt.Errorf("path %q doesn't contain %q", srcPath)
	}
	return &GitRepo{
		srcRoot:    path[:index+len(srcPath)-1],
		importPath: path[index+len(srcPath):],
	}, nil
}

func (r *GitRepo) Load(exclusions []string) error {
	r.modules = nil

	protoModules, err := loadProtoModules(exclusions)
	if err != nil {
		return err
	}

	// Some tags might exist for modules that have been renamed or deleted.
	pathToVersionMap := r.loadTags()

	for _, pm := range protoModules {

		// Sanity check 1
		if !strings.HasPrefix(pm.FullPath(), r.ImportPath()) {
			return fmt.Errorf(
				"module %s doesn't start with the repository name %q",
				pm.FullPath(), r.ImportPath())
		}

		inRepoPath := pm.InRepoPath(r.importPath)

		// Sanity check 2
		if !strings.HasSuffix(pm.RawPath(), inRepoPath) {
			return fmt.Errorf(
				"in file %q, the module name %q doesn't match the file's path",
				filepath.Join(pm.RawPath(), ifc.GoModFile), inRepoPath)
		}

		// Sanity check 3
		p1 := filepath.Join(r.ImportPath(), inRepoPath)
		p2 := pm.mf.Module.Mod.Path
		if !strings.HasPrefix(p2, p1) {
			return fmt.Errorf("path invariant broken; %q != %q", p1, p2)
		}

		// Find the latest version tag
		v := func () *semver.SemVer {
			versions := pathToVersionMap[inRepoPath]
			if versions == nil {
				return nil
			}
			return versions[0]
		}()

		r.modules = append(
			r.modules,
			mod.NewModule(r, inRepoPath, pm.mf, v))
	}
	return nil
}

func (r *GitRepo) AbsPath() string {
	return filepath.Join(r.srcRoot, r.ImportPath())
}

func (r *GitRepo) ImportPath() string {
	return r.importPath
}

func (r *GitRepo) Report() {
	fmt.Println(r.importPath)
	for _, m := range r.modules {
		fmt.Printf("%15s  %s\n", m.Version(), m.InRepoPath())
	}
}

func (r *GitRepo) FindModuleByRelPath(target string) ifc.LaModule {
	for _, m := range r.modules {
		if m.InRepoPath() == target {
			return m
		}
	}
	return nil
}

func (r *GitRepo) Apply(f ifc.ModFunc) error {
	for _, m := range r.modules {
		err := f(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *GitRepo) gitRun(args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = filepath.Join(r.srcRoot, r.importPath)
	if r.doIt {
		out, err := c.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s out=%q", err.Error(), out)
		}
		return string(out), nil
	} else {
		fmt.Printf("in %-60s; %s\n", c.Dir, c.String())
		return "", nil
	}
}

func (r *GitRepo) loadTags() (result map[string]semver.Versions) {
	r.doIt = true
	out, err := r.gitRun("tag", "-l")
	if err != nil {
		panic(err)
	}
	result = make(map[string]semver.Versions)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		fields := strings.Split(l, pathSep)
		v, err := semver.Parse(fields[len(fields)-1])
		if err != nil {
			// Silently ignore versions we don't understand.
			continue
		}
		p := rootedModulePath
		if len(fields) > 1 {
			p = strings.Join(fields[:len(fields)-1], pathSep)
		}
		result[p] = append(result[p], v)
	}
	for _, versions := range result {
		sort.Sort(versions)
	}
	return
}
