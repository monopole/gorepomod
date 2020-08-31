package main

import (
	"fmt"
	"log"
	"os"

	"github.com/monopole/gorepomod/internal/arguments"
	"github.com/monopole/gorepomod/internal/edit"
	"github.com/monopole/gorepomod/internal/repository"
)

const (
	usageMsg = `usage:

  gorepomod unpin {dependency}
  gorepomod pin {dependency} {version}
  gorepomod tidy

e.g.

  gorepomod unpin kyaml --doIt
  gorepomod pin kyaml v0.7.0 --doIt

This program must be run from a local git repository root.
The program walks the repository's tree looking for Go
modules (i.e. looking for 'go.mod' files), and performs
one of the following operations on each module {m}:

  tidy

    Tidy {m}'s go.mod file.

  unpin

    If {m} depends on a {repository}/{dependency},
    then {m}'s dependency on it will be replaced by
    a relative path to the in-repo module.

  pin {version}

    The opposite of 'unpin'.  Replacements are removed,
    and {m}'s dependency is pinned to a specific,
    previously tagged and released version of {dependency}.
    {version} should be in semver form, e.g. v1.2.3.

`
)

func usage(err error) {
	fmt.Print(usageMsg)
	if err != nil {
		fmt.Printf("argument error: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func main() {
	args, err := arguments.Parse()
	if err != nil {
		usage(err)
	}
	repo, err := repository.NewRepoWithExclusion(
		args.RepoName(), args.Exclusions())
	if err != nil {
		usage(err)
	}
	if args.GetCommand() == arguments.Tidy {
		err = repo.Apply(func(m *repository.Module) error {
			return edit.New(m, args.DoIt()).Tidy()
		})
	} else {
		targetDep := repo.FindModuleByRelPath(args.Dependency())
		if targetDep == nil {
			usage(fmt.Errorf(
				"cannot find dependency target module %q in repo %s",
				args.Dependency(), args.RepoName()))
		}
		err = repo.Apply(func(m *repository.Module) error {
			editor := edit.New(m, args.DoIt())
			if args.GetCommand() == arguments.Tidy {
				return editor.Tidy()
			}
			if yes, oldVersion := m.DependsOn(targetDep); yes {
				switch args.GetCommand() {
				case arguments.Pin:
					return editor.Pin(targetDep, oldVersion, args.Version())
				case arguments.UnPin:
					return editor.UnPin(m.Depth(), targetDep, oldVersion)
				}
			}
			return nil
		})
	}
	if err != nil {
		log.Fatal(err)
	}
}
