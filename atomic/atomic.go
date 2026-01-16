package main

import (
	"context"
	"dagger/atomic/internal/dagger"
	"fmt"
)

const (
	// renovate: datasource=docker depName=quay.io/fedora-ostree-desktops/silverblue
	latestFedoraVersion = "40"

	description = "scottames' custom Fedora Silverblue native container image powered by Universal Blue."
)

var (
	labels = map[string]string{
		"io.artifacthub.package.readme-url": "https://raw.githubusercontent.com/scottames/containers/main/atomic/README.md",
		"org.opencontainers.image.url":      "https://github.com/scottames/containers/tree/main/atomic",
	}

	scriptsPostPackageInstall = []string{
		"1Password.sh",
	}
)

// fedoraAtomic defines the custom Fedora Atomic container image
//
// the container and publish functions both refer to this as their source
func (a *Atomic) fedoraAtomic(ctx context.Context) (*dagger.Fedora, error) {
	scriptsPost := []*dagger.File{}
	for _, script := range scriptsPostPackageInstall {
		scriptsPost = append(scriptsPost, a.Source.File(
			fmt.Sprintf(
				"atomic/scripts/%s",
				script,
			),
		))
	}

	opts := dagger.FedoraOpts{
		Registry: a.Registry,
		Org:      a.Org,
		Tag:      a.Tag,
		Variant:  a.Variant,
	}

	// Niri is Silverblue-based - it should be labeled Niri,
	//  but pulled from Silverblue
	if opts.Variant == Niri {
		opts.Variant = Silverblue
	}

	if a.Suffix != nil {
		opts.Suffix = *a.Suffix
	}

	fedora := dag.Fedora(opts)

	version, err := fedora.ReleaseVersion(ctx)
	if err != nil {
		version, err = fedora.Date(ctx)
		if err != nil {
			return nil, err
		}
	}

	a.ReleaseVersion = version

	a.Tags, err = fedora.DefaultTags(ctx,
		dagger.FedoraDefaultTagsOpts{Latest: latestFedoraVersion == version},
	)
	if err != nil {
		return nil, err
	}

	fedora, err = a.fedoraWithLabelsFromCLI(fedora)
	if err != nil {
		return nil, err
	}

	if !a.SkipDefaultLabels {
		fedora, err = a.fedoraWithDefaultLabels(ctx, fedora)
		if err != nil {
			return nil, err
		}

		fedora = fedora.WithDescription(description)
	}

	finalReposForBuild := replaceStringInSlice(
		reposForBuild,
		"FEDORA_MAJOR_VERSION",
		version,
	)

	finalReposForImage := replaceStringInSlice(
		reposForImage,
		"FEDORA_MAJOR_VERSION",
		version,
	)

	// Fedora is derived from the installed dagger module dependency
	return fedora.
			WithDescription(description).
			WithDirectory(
				"/usr",
				a.Source.Directory("atomic/files/usr"),
			).
			// true => keep repo in final image
			WithReposFromUrls(finalReposForImage, true).
			// false => delete repo file in final image
			WithReposFromUrls(finalReposForBuild, false).
			WithPackagesInstalled(
				a.getPackageListFrom(packagesInstalled),
			).
			WithPackagesRemoved(
				a.getPackageListFrom(packagesRemoved),
			).
			WithExecScripts(
				scriptsPost,
				false, // false => post package install
			).
			WithExec(
				[]string{"update-ca-trust"},
				false, // false => post package install
			),
		nil
}
