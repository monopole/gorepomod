package arguments

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/monopole/gorepomod/internal/repository"
	"github.com/monopole/gorepomod/internal/utils"
)

const (
	doItFlag = "--doIt"
	dotGit   = ".git"
	srcPath  = "/src/"
	cmdPin   = "pin"
	cmdUnPin = "unpin"
	cmdTidy  = "tidy"
)

var (
	commands = []string{cmdPin, cmdUnPin, cmdTidy}

	// TODO: make this a PATH-like flag
	// e.g.: --excludes ".git:.idea:site:docs"
	excSlice = []string{
		".git",
		".github",
		".idea",
		"docs",
		"examples",
		"hack",
		"site",
		"releasing",
	}
)

type Command int

const (
	Tidy Command = iota
	UnPin
	Pin
)

type Args struct {
	cmd        Command
	dependency string
	version    *repository.SemanticVersion
	repo       string
	doIt       bool
}

func (a *Args) Report() {
	fmt.Printf("     cmd: %s\n", a.cmd)
	fmt.Printf("     dep: %s\n", a.dependency)
	fmt.Printf(" version: %s\n", a.version)
	fmt.Printf("    repo: %s\n", a.repo)
	fmt.Printf("    doIt: %version\n", a.doIt)
}

func (a *Args) GetCommand() Command {
	return a.cmd
}

func (a *Args) RepoName() string {
	return a.repo
}

func (a *Args) Version() *repository.SemanticVersion {
	return a.version
}

func (a *Args) Dependency() string {
	return a.dependency
}

func (a *Args) Exclusions() map[string]bool {
	result := make(map[string]bool)
	for _, x := range excSlice {
		if _, ok := result[x]; ok {
			log.Fatalf("programmer error - repeated exclusion: %s", x)
		} else {
			result[x] = true
		}
	}
	return result
}

func (a *Args) DoIt() bool {
	return a.doIt
}

func argCount() int {
	return len(os.Args) - 1
}

func Parse() (result *Args, err error) {
	result = &Args{}
	if argCount() < 1 {
		return nil, fmt.Errorf("command needs at least one arg")
	}
	switch os.Args[1] {
	case cmdPin:
		result.cmd = Pin
		if argCount() < 2 {
			return nil, fmt.Errorf("pin needs a dependency to pin")
		}
		result.dependency = os.Args[2]
		if argCount() < 3 {
			return nil, fmt.Errorf("pin needs a version argument, e.g. v1.2.3")
		}
		result.version, err = repository.ParseVersion(os.Args[3])
		if err != nil {
			return nil, err
		}
	case cmdUnPin:
		result.cmd = UnPin
		if argCount() < 2 {
			return nil, fmt.Errorf("unpin needs a dependency to unpin")
		}
		result.dependency = os.Args[2]
	case cmdTidy:
		result.cmd = Tidy
	default:
		return nil, fmt.Errorf("command must be one of %v", commands)
	}
	var dir string
	dir, err = os.Getwd()
	if err != nil {
		return nil, err
	}
	if !utils.DirExists(dotGit) {
		return nil, fmt.Errorf("your pwd %s is not a git repo root", dir)
	}
	index := strings.Index(dir, srcPath)
	if index < 0 {
		return nil, fmt.Errorf("cwd path doesn't contain %q", srcPath)
	}
	result.repo = dir[index+len(srcPath):]
	result.doIt = os.Args[argCount()] == doItFlag
	return
}
