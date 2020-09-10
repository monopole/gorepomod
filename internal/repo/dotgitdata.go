package repo

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monopole/gorepomod/internal/misc"
	"github.com/monopole/gorepomod/internal/semver"
	"github.com/monopole/gorepomod/internal/utils"
)

const (
	dotGitFileName = ".git"
	srcHint        = "/src/"
	pathSep        = "/"
)

// DotGitData holds basic information about a local .git file
type DotGitData struct {
	// srcPath is the absolute path to the local Go src directory.
	// This used to be $GOPATH/src.
	// It's the directory containing git repository clones.
	srcPath string
	// The path below srcPath to a particular repository
	// directory, a directory containing a .git directory.
	// Typically {repoOrg}/{repoUserName}, e.g. sigs.k8s.io/cli-utils
	repoPath string
}

func (dg *DotGitData) SrcPath() string {
	return dg.srcPath
}

func (dg *DotGitData) RepoPath() string {
	return dg.repoPath
}

func (dg *DotGitData) AbsPath() string {
	return filepath.Join(dg.srcPath, dg.repoPath)
}

func (dg *DotGitData) Report() {
	fmt.Printf("   src path: %s\n", dg.SrcPath())
	fmt.Printf("  repo path: %s\n", dg.RepoPath())
	fmt.Printf("   abs path: %s\n", dg.AbsPath())
}

// NewDotGitDataFromPath wants the incoming path to hold dotGit
// E.g.
//   ~/gopath/src/sigs.k8s.io/kustomize
//   ~/gopath/src/github.com/monopole/gorepomod
func NewDotGitDataFromPath(path string) (*DotGitData, error) {
	if !utils.DirExists(filepath.Join(path, dotGitFileName)) {
		return nil, fmt.Errorf(
			"%q doesn't have a %q file", path, dotGitFileName)
	}
	// This is an attempt to figure out where the user has cloned
	// their repos.  In the old days, it was an import path under
	// $GOPATH/src.  If we cannot guess it, we may need to ask for it,
	// or maybe proceed without knowing it.
	index := strings.Index(path, srcHint)
	if index < 0 {
		return nil, fmt.Errorf(
			"path %q doesn't contain %q", path, srcHint)
	}
	return &DotGitData{
		srcPath:  path[:index+len(srcHint)-1],
		repoPath: path[index+len(srcHint):],
	}, nil
}

// It's a factory factory.
func (dg *DotGitData) NewRepoFactory(
	exclusions []string) (*ManagerFactory, error) {
	modules, err := loadProtoModules(dg.AbsPath(), exclusions)
	if err != nil {
		return nil, err
	}
	err = dg.checkModules(modules)
	if err != nil {
		return nil, err
	}

	runner := newGitRunner(dg.AbsPath(), true)
	remoteName, err := determineRemoteToUse(runner)
	if err != nil {
		return nil, err
	}

	// Some tags might exist for modules that
	// have been renamed or deleted; ignore those.
	// There might be newer tags locally than remote,
	// so report both.
	localTags, err := loadLocalTags(runner)
	if err != nil {
		return nil, err
	}
	remoteTags, err := loadRemoteTags(runner, remoteName)
	if err != nil {
		return nil, err
	}

	return &ManagerFactory{
		dg:               dg,
		modules:          modules,
		remoteName:       remoteName,
		versionMapLocal:  localTags,
		versionMapRemote: remoteTags,
	}, nil
}

func (dg *DotGitData) checkModules(modules []*protoModule) error {
	for _, pm := range modules {

		file := filepath.Join(pm.PathToGoMod(), goModFile)

		// Do the paths make sense?
		if !strings.HasPrefix(pm.FullPath(), dg.RepoPath()) {
			return fmt.Errorf(
				"module %q doesn't start with the repository name %q",
				pm.FullPath(), dg.RepoPath())
		}

		shortName := pm.ShortName(dg.RepoPath())
		if shortName == misc.ModuleAtTop {
			if pm.PathToGoMod() != dg.AbsPath() {
				return fmt.Errorf("in %q, problem with top module", file)
			}
		} else {
			// Do the relative path and short name make sense?
			if !strings.HasSuffix(pm.PathToGoMod(), string(shortName)) {
				return fmt.Errorf(
					"in %q, the module name %q doesn't match the file's pathToGoMod %q",
					file, shortName, pm.PathToGoMod())
			}
		}
	}
	return nil
}

// TODO: allow for other remote names.
func determineRemoteToUse(runner *gitRunner) (misc.TrackedRepo, error) {
	out, err := runner.run("remote")
	if err != nil {
		return "", err
	}
	remotes := strings.Split(out, "\n")
	if len(remotes) < 1 {
		return "", fmt.Errorf("need at least one remote")
	}
	for _, n := range misc.RecognizedRemotes {
		if contains(remotes, n) {
			return n, nil
		}
	}
	return "", fmt.Errorf(
		"unable to find recognized remote %v", misc.RecognizedRemotes)
}

func contains(list []string, item misc.TrackedRepo) bool {
	for _, n := range list {
		if n == string(item) {
			return true
		}
	}
	return false
}

func loadLocalTags(
	runner *gitRunner) (result misc.VersionMap, err error) {
	var out string
	out, err = runner.run("tag", "-l")
	if err != nil {
		return nil, err
	}
	result = make(misc.VersionMap)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		n, v, err := parseModuleSpec(l)
		if err != nil {
			// ignore it
			continue
		}
		result[n] = append(result[n], v)
	}
	for _, versions := range result {
		sort.Sort(versions)
	}
	return
}

func loadRemoteTags(
	runner *gitRunner,
	remote misc.TrackedRepo) (result misc.VersionMap, err error) {
	var out string
	out, err = runner.run("ls-remote", "--ref", string(remote))
	if err != nil {
		return nil, err
	}
	result = make(misc.VersionMap)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		fields := strings.Fields(l)
		if len(fields) < 2 {
			// ignore it
			continue
		}
		if !strings.HasPrefix(fields[1], refsTags) {
			// ignore it
			continue
		}
		path := fields[1][len(refsTags):]
		n, v, err := parseModuleSpec(path)
		if err != nil {
			// ignore it
			continue
		}
		result[n] = append(result[n], v)
	}
	for _, versions := range result {
		sort.Sort(versions)
	}
	return
}

func parseModuleSpec(
	path string) (n misc.ModuleShortName, v semver.SemVer, err error) {
	fields := strings.Split(path, pathSep)
	v, err = semver.Parse(fields[len(fields)-1])
	if err != nil {
		// Silently ignore versions we don't understand.
		return "", v, err
	}
	n = misc.ModuleAtTop
	if len(fields) > 1 {
		n = misc.ModuleShortName(
			strings.Join(fields[:len(fields)-1], pathSep))
	}
	return
}
