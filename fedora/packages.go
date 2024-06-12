package main

import (
	"context"
	"fmt"
	"path/filepath"
)

const etcYumReposD = "/etc/yum.repos.d/"

// Repo represents a yum repository object
type Repo struct {
	Url      string
	FileName string
	Keep     bool
}

// WithReposFromUrls will add the content at each given url and install them
// on the Container image prior to package installation via
// WithPackagesInstalled. Optionally removing the repository afterward, prior
// to exporting the atomic container.
func (f *Fedora) WithReposFromUrls(
	ctx context.Context,
	// urls of yum repository files to install (i.e. GitHub raw file)
	urls []string,
	// If true, the repository will not be removed on the generated atomic
	// Container image
	keep bool,
) *Fedora {
	for _, r := range urls {
		repo := Repo{
			Url:      r,
			Keep:     keep,
			FileName: filepath.Base(r),
		}
		f.Repos = append(f.Repos, &repo)
	}

	return f
}

// WithPackagesInstalled will install the given packages on the generated
// Container image.
func (f *Fedora) WithPackagesInstalled(
	ctx context.Context,
	// list of packages to be installed on the generated atomic Container
	// image
	packages []string,
) *Fedora {
	f.PackagesInstalled = packages

	return f
}

// WithPackagesRemoved will remove the given packages on the generated atomic
// Container image.
func (f *Fedora) WithPackagesRemoved(
	ctx context.Context,
	// list of packages to be removed on the generated atomic Container image
	packages []string,
) *Fedora {
	f.PackagesRemoved = packages

	return f
}

// ctrWithReposInstalled will get the f.Repos urls and install them in the
// returned Container image.
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

// ctrWithReposRemoved will remove f.Repos which are not marked to be kept
// in the generated atomic Container image.
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

// ctrWithPackagesInstalledAndRemoved will return the given Container with
// packages installed and removed as defined by the Fedora object
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

		return f.ctrWithExec(ctr, cmd)

	} else if removePackages {
		return f.ctrWithExec(ctr,
			[]string{"rpm-ostree", "override", "remove"},
			f.PackagesRemoved...,
		)
	} else if installPackages {
		return f.ctrWithExec(ctr,
			[]string{"rpm-ostree", "install"},
			f.PackagesInstalled...,
		)
	}

	return ctr
}
