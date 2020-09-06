package arguments

import (
	"fmt"
	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/semver"
	"github.com/monopole/gorepomod/internal/utils"
	"os"
)

const (
	doItFlag = "--doIt"
	cmdPin   = "pin"
	cmdUnPin = "unpin"
	cmdTidy  = "tidy"
	cmdList  = "list"
)

var (
	commands = []string{cmdPin, cmdUnPin, cmdTidy, cmdList}

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
	List
)

type Args struct {
	cmd        Command
	dependency ifc.ModuleShortName
	version    semver.SemVer
	doIt       bool
}

func (a *Args) GetCommand() Command {
	return a.cmd
}

func (a *Args) Version() semver.SemVer {
	return a.version
}

func (a *Args) Dependency() ifc.ModuleShortName {
	return a.dependency
}

func (a *Args) Exclusions() (result []string) {
	// Make sure the list has no repeats.
	for k := range utils.SliceToSet(excSlice) {
		result = append(result, k)
	}
	return
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
		result.dependency = ifc.ModuleShortName(os.Args[2])
		if argCount() < 3 {
			return nil, fmt.Errorf("pin needs a version argument, e.g. v1.2.3")
		}
		result.version, err = semver.Parse(os.Args[3])
		if err != nil {
			return nil, err
		}
	case cmdUnPin:
		result.cmd = UnPin
		if argCount() < 2 {
			return nil, fmt.Errorf("unpin needs a dependency to unpin")
		}
		result.dependency = ifc.ModuleShortName(os.Args[2])
	case cmdTidy:
		result.cmd = Tidy
	case cmdList:
		result.cmd = List
	default:
		return nil, fmt.Errorf("command must be one of %v", commands)
	}
	result.doIt = os.Args[argCount()] == doItFlag
	return
}
