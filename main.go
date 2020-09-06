package main

import (
	"fmt"
	"github.com/monopole/gorepomod/internal/arguments"
	"github.com/monopole/gorepomod/internal/edit"
	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/repo"
	"log"
	"os"
	"strconv"
)

const (
	usageMsg = `# gorepomod

A tool for managing Go modules in a git repository
with more than one Go module, where there are
dependencies between the modules.

This is a fancy version of

  find ./ -name "go.mod" | xargs go mod {some operation}

Run it from a local git repository root.

It walks the repository's tree looking for Go modules
(i.e. go.mod files), loads and examines them all,
and does the following on each module _m_:

 - list

   Lists the modules and inter-repo dependencies.

 - tidy

   Tidy _m_'s go.mod file.

 - unpin {module}

   If _m_ depends on a _{repository}/{module}_,
   then _m_'s dependency on it will be replaced by
   a relative path to the in-repo module.

 - pin {module} {version}

   The opposite of 'unpin'.  Replacements are removed,
   and _m_'s dependency is pinned to a specific, previously
   tagged and released version of _{module}_.
   _{version}_ should be in semver form, e.g. v1.2.3.

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
	gr, err := repo.NewFromCwd()
	if err != nil {
		usage(err)
	}
	err = gr.Load(args.Exclusions())
	if err != nil {
		usage(err)
	}
	switch args.GetCommand() {
	case arguments.List:
		err = gr.Apply(func(m ifc.LaModule) error {
			fmt.Printf(
				"%10s  %-" + strconv.Itoa(gr.LenLongestName()+2) + "s%v\n",
				m.Version(), m.ShortName(),  gr.InternalDeps(m))
			return nil
		})
	case arguments.Tidy:
		err = gr.Apply(func(m ifc.LaModule) error {
			return edit.New(m, args.DoIt()).Tidy()
		})
	case arguments.Pin:
		fallthrough
	case arguments.UnPin:
		targetDep := gr.FindModule(args.Dependency())
		if targetDep == nil {
			usage(fmt.Errorf(
				"cannot find dependency target module %q in repo %s",
				args.Dependency(), gr.ImportPath()))
		}
		err = gr.Apply(func(m ifc.LaModule) error {
			editor := edit.New(m, args.DoIt())
			if yes, oldVersion := m.DependsOn(targetDep); yes {
				if args.GetCommand() == arguments.Pin {
					return editor.Pin(targetDep, oldVersion, args.Version())
				}
				return editor.UnPin(m.ShortName().Depth(), targetDep, oldVersion)
			}
			return nil
		})
	}
	if err != nil {
		log.Fatal(err)
	}
}
