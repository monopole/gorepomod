package edit

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/monopole/gorepomod/internal/repository"
)

// Editor runs `go mod` commands on an instance of Module.
// If doIt is false, the command is printed, but not run.
type Editor struct {
	module *repository.Module
	doIt   bool
}

func New(m *repository.Module, doIt bool) *Editor {
	return &Editor{
		doIt:   doIt,
		module: m,
	}
}

func (e *Editor) run(args ...string) error {
	c := exec.Command(
		"go",
		append([]string{"mod"}, args...)...)
	c.Dir = e.module.InRepoPath()
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
	target *repository.Module, oldV, newV *repository.SemanticVersion) error {
	return e.run(
		"edit",
		"-dropreplace="+target.FullPath()+"@"+oldV.String(),
		"-require="+target.FullPath()+"@"+newV.String(),
	)
}

func (e *Editor) UnPin(
	depth int, target *repository.Module, oldV *repository.SemanticVersion) error {
	var r strings.Builder
	r.WriteString(target.FullPath())
	r.WriteString("@")
	r.WriteString(oldV.String())
	r.WriteString("=")
	r.WriteString(upstairs(depth))
	r.WriteString(target.InRepoPath())
	return e.run(
		"edit",
		"-replace="+r.String(),
	)
}
