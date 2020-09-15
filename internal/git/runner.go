package git

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/monopole/gorepomod/internal/misc"
	"github.com/monopole/gorepomod/internal/semver"
)

const (
	refsTags       = "refs/tags/"
	pathSep        = "/"
	remoteOrigin   = misc.TrackedRepo("origin")
	remoteUpstream = misc.TrackedRepo("upstream")
	mainBranch   = "master"
)

var recognizedRemotes = []misc.TrackedRepo{remoteUpstream, remoteOrigin}

// Runner runs specific git tasks using the git CLI.
type Runner struct {
	// From which directory do we run the commands.
	workDir string
	// Run commands, or merely print commands.
	doIt    bool
}

func New(wd string, doIt bool) *Runner {
	return &Runner{workDir: wd, doIt: doIt}
}

func (gr *Runner) run(args ...string) (string, error) {
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

func (gr *Runner) runNoOut(args ...string) error {
	_, err := gr.run(args...)
	return err
}

// TODO: allow for other remote names.
func (gr *Runner) DetermineRemoteToUse() (misc.TrackedRepo, error) {
	out, err := gr.run("remote")
	if err != nil {
		return "", err
	}
	remotes := strings.Split(out, "\n")
	if len(remotes) < 1 {
		return "", fmt.Errorf("need at least one remote")
	}
	for _, n := range recognizedRemotes {
		if contains(remotes, n) {
			return n, nil
		}
	}
	return "", fmt.Errorf(
		"unable to find recognized remote %v", recognizedRemotes)
}

func contains(list []string, item misc.TrackedRepo) bool {
	for _, n := range list {
		if n == string(item) {
			return true
		}
	}
	return false
}

func (gr *Runner) LoadLocalTags() (result misc.VersionMap, err error) {
	var out string
	out, err = gr.run("tag", "-l")
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

func (gr *Runner) LoadRemoteTags(
	remote misc.TrackedRepo) (result misc.VersionMap, err error) {
	var out string
	out, err = gr.run("ls-remote", "--ref", string(remote))
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


func (gr *Runner) Debug(remote misc.TrackedRepo) error {
	return nil // gr.CheckoutMainBranch(remote)
}

func (gr *Runner) AssureCleanWorkspace() error {
	out, err := gr.run("status")
	if err != nil {
		return err
	}
	if gr.doIt && !strings.Contains(out, "nothing to commit, working tree clean") {
		return fmt.Errorf("the workspace isn't clean")
	}
	return nil
}

func (gr *Runner) AssureOnMainBranch() error {
	out, err := gr.run("status")
	if err != nil {
		return err
	}
	if gr.doIt && !strings.Contains(out, "On branch " + mainBranch) {
		return fmt.Errorf("expected to be on branch %q", mainBranch)
	}
	return nil
}

// CheckoutMainBranch does that.
func (gr *Runner) CheckoutMainBranch() error {
	return gr.runNoOut("checkout", mainBranch)
}

// FetchRemote does that.
func (gr *Runner) FetchRemote(remote misc.TrackedRepo) error {
	return gr.runNoOut("fetch", string(remote))
}

// MergeFromRemoteMain does a fast forward only merge with main branch.
func (gr *Runner) MergeFromRemoteMain(remote misc.TrackedRepo) error {
	return gr.runNoOut("merge", "--ff-only",
		strings.Join(
			[]string{string(remote), mainBranch}, pathSep))
}

// CheckoutReleaseBranch attempts to checkout or create a branch.
// If it's on the remote already, fail if we cannot check it out locally.
func (gr *Runner) CheckoutReleaseBranch(
		remote misc.TrackedRepo, branch string) error {
	yes, err := gr.doesRemoteBranchExist(remote, branch)
	if err != nil {
		return err
	}
	if yes {
		// Assume that if there's a remote, we also have a local.
		// Might be a bad assumption; if so, this returns an error.
		return gr.runNoOut("checkout", branch)
	}
	// Create the branch and check it out.
	out, err := gr.run("checkout", "-b", branch)
	if err != nil {
		return err
	}
	if gr.doIt && !strings.Contains(out, "Switched to new a branch") {
		return fmt.Errorf("unexpected branch creation output: %q", out)
	}
	return nil
}

func (gr *Runner) doesRemoteBranchExist(
		remote misc.TrackedRepo, branch string) (bool, error) {
	out, err := gr.run("branch", "-r")
	if err != nil {
		return false, err
	}
	lookFor := strings.Join([]string{string(remote), branch}, pathSep)
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == lookFor {
			return true, nil
		}
	}
	return false, nil
}

func (gr *Runner) PushBranchToRemote(
	remote misc.TrackedRepo, branch string) error {
	return gr.runNoOut("push", "-f", string(remote), branch)
}

func (gr *Runner) CreateLocalReleaseTag(tag, branch string) error {
	return gr.runNoOut(
		"tag", "-a",
		"-m", fmt.Sprintf("\"Release %s on branch %s\"", tag, branch),
		tag)
}

func (gr *Runner) DeleteLocalTag(tag string) error {
	return gr.runNoOut("tag", "--delete", tag)
}

func (gr *Runner) PushTagToRemote(
	remote misc.TrackedRepo, tag string) error {
	return gr.runNoOut("push", string(remote), tag)
}

func (gr *Runner) DeleteTagFromRemote(
	remote misc.TrackedRepo, tag string) error {
	return gr.runNoOut("push", string(remote), ":"+refsTags+tag)
}
