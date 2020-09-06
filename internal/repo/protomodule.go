package repo

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/utils"
	"golang.org/x/mod/modfile"
)

type protoModule struct {
	path string
	mf   *modfile.File
}

func (pm *protoModule) FullPath() string {
	return pm.mf.Module.Mod.Path
}

func (pm *protoModule) RawPath() string {
	return pm.path
}

// Represents the trailing version label in a module name.
// See https://blog.golang.org/v2-go-modules
var trailingVersionPattern = regexp.MustCompile("/v\\d+$")

// InRepoPath is in repo path to the go.mod file.
// It's usually a short path, e.g. "" (empty), "kyaml", "cmd/config".
// The same string used to tag the repo at a particular module version.
func (pm *protoModule) InRepoPath(repoImportPath string) string {
	p := pm.FullPath()[len(repoImportPath)+1:]
	return trailingVersionPattern.ReplaceAllString(p, "")
}

func loadProtoModules(
	exclusions []string) (result []*protoModule, err error) {
	var paths []string
	paths, err = getPathsToModules(exclusions)
	if err != nil {
		return
	}
	for _, p := range paths {
		var pm *protoModule
		pm, err = loadProtoModule(p)
		if err != nil {
			return
		}
		result = append(result, pm)
	}
	return
}

func loadProtoModule(path string) (*protoModule, error) {
	mPath := filepath.Join(path, ifc.GoModFile)
	content, err := ioutil.ReadFile(mPath)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %v\n", mPath, err)
	}
	f, err := modfile.Parse(mPath, content, nil)
	if err != nil {
		return nil, err
	}
	return &protoModule{path: path, mf: f}, nil
}

func getPathsToModules(exclusions []string) (result []string, err error) {
	exclusionMap := utils.SliceToSet(exclusions)
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
			if info.Name() == ifc.GoModFile {
				result = append(
					result, path[:len(path)-len(ifc.GoModFile)-1])
				return filepath.SkipDir
			}
			return nil
		})
	return
}
