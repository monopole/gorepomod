package repo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/monopole/gorepomod/internal/edit"
	"github.com/monopole/gorepomod/internal/misc"
	"github.com/monopole/gorepomod/internal/semver"
)

// Manager manages a git repo.
// All data already loaded and validated, it's ready to go.
type Manager struct {
	// Underlying file system facts.
	dg *DotGitData

	// The remote used for fetching tags, pushing tags,
	// and pushing release branches.
	remoteName misc.TrackedRepo

	// The list of known Go modules in the repo.
	modules misc.LesModules
}

func (mgr *Manager) AbsPath() string {
	return mgr.dg.AbsPath()
}

func (mgr *Manager) RepoPath() string {
	return mgr.dg.RepoPath()
}

func (mgr *Manager) FindModule(
	target misc.ModuleShortName) misc.LaModule {
	return mgr.modules.Find(target)
}

func (mgr *Manager) Tidy(doIt bool) error {
	return mgr.modules.Apply(func(m misc.LaModule) error {
		return edit.New(m, doIt).Tidy()
	})
}

func (mgr *Manager) Pin(
	doIt bool, target misc.LaModule, newV semver.SemVer) error {
	return mgr.modules.Apply(func(m misc.LaModule) error {
		if yes, oldVersion := m.DependsOn(target); yes {
			return edit.New(m, doIt).Pin(target, oldVersion, newV)
		}
		return nil
	})
}

func (mgr *Manager) UnPin(doIt bool, target misc.LaModule) error {
	return mgr.modules.Apply(func(m misc.LaModule) error {
		if yes, oldVersion := m.DependsOn(target); yes {
			return edit.New(m, doIt).UnPin(target, oldVersion)
		}
		return nil
	})
}

func (mgr *Manager) List() error {
	fmt.Printf("   src path: %s\n", mgr.dg.SrcPath())
	fmt.Printf("  repo path: %s\n", mgr.RepoPath())
	fmt.Printf("     remote: %s\n", mgr.remoteName)
	format := "%-" + strconv.Itoa(mgr.modules.LenLongestName()+2) + "s%-11s%-11s%s\n"
	fmt.Printf(
		format, "NAME", "LOCAL", "REMOTE", "INTRA-REPO-DEPENDENCIES")
	return mgr.modules.Apply(func(m misc.LaModule) error {
		fmt.Printf(
			format, m.ShortName(),
			m.VersionLocal(), m.VersionRemote(), mgr.modules.InternalDeps(m))
		return nil
	})
}

func determineBranchAndTag(
	m misc.LaModule, v semver.SemVer) (string, string) {
	if m.ShortName() == misc.ModuleAtTop {
		return fmt.Sprintf("release-%s", v.BranchLabel()), v.String()
	}
	return fmt.Sprintf(
			"release-%s-%s", m.ShortName(), v.BranchLabel()),
		string(m.ShortName()) + "/" + v.String()
}

func doesBranchExist(gr *gitRunner, name string) (bool, error) {
	out, err := gr.run("branch", "-a")
	if err != nil {
		return false, err
	}
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if strings.Trim(l, " ") == name {
			return true, nil
		}
	}
	return false, nil
}

func assureBranchExists(gr *gitRunner, name string) error {
	yes, err := doesBranchExist(gr, name)
	if err != nil {
		return err
	}
	if yes {
		return nil
	}
	out, err := gr.run("checkout", "-b", name)
	if err != nil {
		return err
	}
	if !strings.Contains(out, "switched to new branch") {
		return fmt.Errorf("unexpected branch creation output: %q", out)
	}
	return nil
}

func (mgr *Manager) pushBranchToRemote(gr *gitRunner, name string) error {
	_, err := gr.run("push", "-f", string(mgr.remoteName), name)
	return err
}

func (mgr *Manager) Release(
	target misc.LaModule, bump semver.SvBump, doIt bool) error {

	newVersion := target.VersionLocal().Bump(bump)

	fmt.Printf(
		"Releasing %s, stepping from %s to %s\n",
		target.ShortName(), target.VersionLocal(), newVersion)

	if newVersion.Equals(target.VersionRemote()) {
		return fmt.Errorf(
			"version %s already exists on remote - delete it first", newVersion)
	}
	if newVersion.LessThan(target.VersionRemote()) {
		fmt.Printf(
			"version %s is less than the most recent remote version (%s)",
			newVersion, target.VersionRemote())
	}

	gr := newGitRunner(mgr.AbsPath(), doIt)

	branch, tag := determineBranchAndTag(target, newVersion)
	if err := assureBranchExists(gr, branch); err != nil {
		return err
	}

	// Create the tag locally.
	if _, err := gr.run(
		"tag", "-a", tag, "-m",
		fmt.Sprintf("Release %s on branch %s", tag, branch)); err != nil {
		return err
	}
	// Push the branch to the remote.
	if _, err := gr.run(
		"push", "-f", string(mgr.remoteName), branch); err != nil {
		return err
	}
  // Push the tag to the remote.
	if _, err := gr.run(
		"push", string(mgr.remoteName), tag); err != nil {
		return err
	}
	fmt.Println("release complete")
	return nil
}

// TODO: implement
func (mgr *Manager) UnRelease(target misc.LaModule, doIt bool) error {
	runner := newGitRunner(mgr.AbsPath(), doIt)
	_, err := runner.run("status")
	return err
}
