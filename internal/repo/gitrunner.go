package repo

import (
	"fmt"
	"os/exec"
	"strings"
)

type gitRunner struct {
	workDir string
	doIt    bool
}

func newGitRunner(wd string, doIt bool) *gitRunner {
	return &gitRunner{workDir: wd, doIt: doIt}
}

func (gr *gitRunner) run(args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = gr.workDir
	if gr.doIt {
		out, err := c.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf(
				"%s out=%q", err.Error(), strings.TrimSpace(string(out)))
		}
		return string(out), nil
	}
	fmt.Printf("in %-60s; %s\n", c.Dir, c.String())
	return "", nil
}

func (gr *gitRunner) runNoOut(args ...string) error {
	_, err := gr.run(args...)
	return err
}
