package main

import (
	"context"
	"fmt"
	"path/filepath"
)

const etcYumReposD = "/etc/yum.repos.d/"

// TODO: docs
type Repo struct {
	Url      string
	FileName string
	Keep     bool
}

func (f *Fedora) WithRepos(
	ctx context.Context,
	// Yum repositories to install
	repos []string,
	keep bool,
) *Fedora {
	foo := true
	if !keep {
		foo = false
	}
	for _, r := range repos {
		repo := Repo{
			Url:      r,
			Keep:     foo,
			FileName: filepath.Base(r),
		}
		f.Repos = append(f.Repos, &repo)
	}

	return f
}

func (f *Fedora) WithPackagesInstalled(ctx context.Context, packages []string) *Fedora {
	f.PackagesInstalled = packages

	return f
}

func (f *Fedora) WithPackagesRemoved(ctx context.Context, packages []string) *Fedora {
	f.PackagesRemoved = packages

	return f
}

func (f *Fedora) ctrWithReposInstalled(ctr *Container) (*Container, error) {
	if f.Repos == nil {
		return ctr, nil
	}

	for _, r := range f.Repos {
		contents, err := f.httpGet(r.Url)
		if err != nil {
			return nil, fmt.Errorf("error getting repo (%s) url: %w", r.FileName, err)
		}

		fileOpts := ContainerWithNewFileOpts{
			Contents:    string(contents),
			Permissions: 0644,
			Owner:       "root:root",
		}
		ctr = ctr.WithNewFile(fmt.Sprintf("%s/%s", etcYumReposD, r.FileName), fileOpts)
	}

	return ctr, nil
}

func (f *Fedora) ctrWithReposRemoved(ctr *Container) *Container {
	cmd := []string{"rm", "-f"}
	run := false
	for _, r := range f.Repos {
		if !r.Keep {
			run = true
			cmd = append(cmd, fmt.Sprintf("%s/%s", etcYumReposD, r.FileName))
		}
	}

	if run {
		return ctr.WithExec(cmd)
	}

	return ctr
}

func (f *Fedora) ctrWithPackagesInstalledAndRemoved(ctr *Container) *Container {
	removePackages := f.PackagesRemoved != nil && len(f.PackagesRemoved) > 0
	installPackages := f.PackagesInstalled != nil && len(f.PackagesInstalled) > 0

	// Doing both actions in one command allows for replacing required
	// packages with alternatives
	if installPackages && removePackages {
		cmd := []string{"rpm-ostree", "override", "remove"}
		cmd = append(cmd, f.PackagesRemoved...)
		for _, p := range f.PackagesInstalled {
			cmd = append(cmd, fmt.Sprintf("--install=%s", p))
		}
		return f.
			ctrWithExec(
				ctr,
				cmd,
			)
	} else if removePackages {
		return f.
			ctrWithExec(
				ctr,
				[]string{"rpm-ostree", "override", "remove"},
				f.PackagesRemoved...,
			)
	} else if installPackages {
		return f.
			ctrWithExec(
				ctr,
				[]string{"rpm-ostree", "install"},
				f.PackagesInstalled...,
			)
	}

	return ctr
}
