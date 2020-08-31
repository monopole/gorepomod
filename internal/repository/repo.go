package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

const dotDir = "."

type Repo struct {
	root    string
	modules []*Module
}

func NewRepo(root string) (*Repo, error) {
	return NewRepoWithExclusion(root, make(map[string]bool))
}

func NewRepoWithExclusion(
	root string, exclusionMap map[string]bool) (*Repo, error) {
	repo := &Repo{root: root}
	paths, err := getPathsToModules(exclusionMap)
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		m, err := ParseIntoRepo(repo, p)
		if err != nil {
			return nil, err
		}
		repo.modules = append(repo.modules, m)
	}
	return repo, nil
}

func (r *Repo) GetRoot() string {
	return r.root
}

func (r *Repo) FindModuleByRelPath(target string) *Module {
	for _, m := range r.modules {
		if m.InRepoPath() == target {
			return m
		}
	}
	return nil
}

type ModFunc func(*Module) error

func (r *Repo) Apply(f ModFunc) error {
	for _, m := range r.modules {
		err := f(m)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPathsToModules(
	exclusionMap map[string]bool) (result []string, err error) {
	err = filepath.Walk(
		dotDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("trouble at path %q: %v\n", path, err)
			}
			if info.IsDir() {
				if _, ok := exclusionMap[info.Name()]; ok {
					return filepath.SkipDir
				}
				return nil
			}
			if info.Name() == GoModFile {
				result = append(
					result, path[:len(path)-len(GoModFile)-1])
				return filepath.SkipDir
			}
			return nil
		})
	return
}
