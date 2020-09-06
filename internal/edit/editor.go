package edit

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/monopole/gorepomod/internal/ifc"
	"github.com/monopole/gorepomod/internal/semver"
)

// Editor runs `go mod` commands on an instance of Module.
// If doIt is false, the command is printed, but not run.
type Editor struct {
	module ifc.LaModule
	doIt   bool
}

func New(m ifc.LaModule, doIt bool) *Editor {
	return &Editor{
		doIt:   doIt,
		module: m,
	}
}

func (e *Editor) run(args ...string) error {
	c := exec.Command(
		"go",
		append([]string{"mod"}, args...)...)
	c.Dir = string(e.module.ShortName())
	if e.doIt {
		out, err := c.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s out=%q", err.Error(), out)
		}
	} else {
		fmt.Printf("in %-60s; %s\n", c.Dir, c.String())
	}
	return nil
}

func upstairs(depth int) string {
	var b strings.Builder
	for i := 0; i < depth; i++ {
		b.WriteString("../")
	}
	return b.String()
}

func (e *Editor) Tidy() error {
	return e.run("tidy")
}

func (e *Editor) Pin(
	target ifc.LaModule, oldV, newV semver.SemVer) error {
	return e.run(
		"edit",
		"-dropreplace="+target.SrcRelativePath()+"@"+oldV.String(),
		"-require="+target.SrcRelativePath()+"@"+newV.String(),
	)
}

func (e *Editor) UnPin(
	depth int, target ifc.LaModule, oldV semver.SemVer) error {
	var r strings.Builder
	r.WriteString(target.SrcRelativePath())
	r.WriteString("@")
	r.WriteString(oldV.String())
	r.WriteString("=")
	r.WriteString(upstairs(depth))
	r.WriteString(string(target.ShortName()))
	return e.run(
		"edit",
		"-replace="+r.String(),
	)
}
