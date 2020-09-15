package repo

import (
	"fmt"
	"strconv"

	"github.com/monopole/gorepomod/internal/edit"
	"github.com/monopole/gorepomod/internal/git"
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
			m.VersionLocal().Pretty(),
			m.VersionRemote().Pretty(),
			mgr.modules.InternalDeps(m))
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

func (mgr *Manager) Debug(target misc.LaModule, doIt bool) error {
	gr := git.New(mgr.AbsPath(), doIt)
	return gr.Debug(mgr.remoteName)
}

// Release supports a gitlab flow style release process.
//
// * All development happens in the branch named "master".
// * Each minor release gets its own branch.
// *
func (mgr *Manager) Release(
	target misc.LaModule, bump semver.SvBump, doIt bool) error {

	newVersion := target.VersionLocal().Bump(bump)

	if newVersion.Equals(target.VersionRemote()) {
		return fmt.Errorf(
			"version %s already exists on remote - delete it first", newVersion)
	}
	if newVersion.LessThan(target.VersionRemote()) {
		fmt.Printf(
			"version %s is less than the most recent remote version (%s)",
			newVersion, target.VersionRemote())
	}

	fmt.Printf(
		"Releasing %s, stepping from %s to %s\n",
		target.ShortName(), target.VersionLocal(), newVersion)

	gr := git.New(mgr.AbsPath(), doIt)

	relBranch, relTag := determineBranchAndTag(target, newVersion)

	if err := gr.AssureCleanWorkspace(); err != nil {
		return err
	}
	if err := gr.FetchRemote(mgr.remoteName); err != nil {
		return err
	}
	if err := gr.CheckoutMainBranch(); err != nil {
		return err
	}
	if err := gr.MergeFromRemoteMain(mgr.remoteName); err != nil {
		return err
	}
	if err := gr.AssureCleanWorkspace(); err != nil {
		return err
	}
	if err := gr.CheckoutReleaseBranch(mgr.remoteName, relBranch); err != nil {
		return err
	}
	if err := gr.MergeFromRemoteMain(mgr.remoteName); err != nil {
		return err
	}
	if err := gr.PushBranchToRemote(mgr.remoteName, relBranch); err != nil {
		return err
	}
	if err := gr.CreateLocalReleaseTag(relTag, relBranch); err != nil {
		return err
	}
	if err := gr.PushTagToRemote(mgr.remoteName, relTag); err != nil {
		return err
	}
	if err := gr.CheckoutMainBranch(); err != nil {
		return err
	}
	return nil
}

func (mgr *Manager) UnRelease(target misc.LaModule, doIt bool) error {
	fmt.Printf(
		"Unreleasing %s/%s\n",
		target.ShortName(), target.VersionRemote())

	_, tag := determineBranchAndTag(target, target.VersionRemote())

	gr := git.New(mgr.AbsPath(), doIt)

	if err := gr.DeleteTagFromRemote(mgr.remoteName, tag); err != nil {
		return err
	}
	if err := gr.DeleteLocalTag(tag); err != nil {
		return err
	}
	return nil
}
