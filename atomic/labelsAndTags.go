package main

import (
	"context"
	"fmt"
	"strings"
)

// fedoraWithLabelsFromCLI returns the provided Fedora object with the labels
// from the CLI added
func (a *Atomic) fedoraWithLabelsFromCLI(
	fedora *Fedora,
) (*Fedora, error) {
	for _, l := range a.Labels {
		ll := strings.SplitN(l, "=", 2)
		if len(ll) < 2 {
			return nil, fmt.Errorf("invalid label: %s", ll)
		}

		fedora = fedora.WithLabel(ll[0], ll[1])
	}

	return fedora, nil
}

// fedoraWithDefaultLabels returns the provided Fedora object with pre-defined
// labels added:
//
//	org.opencontainers.image.version
//	org.opencontainers.image.base_image
//	org.opencontainers.image.base_image_version
//	io.artifacthub.package.logo-url (if org=ublue-os)
func (a *Atomic) fedoraWithDefaultLabels(
	ctx context.Context,
	fedora *Fedora,
) (*Fedora, error) {
	// note: universal blue appends a build number, we do not
	fedora = fedora.WithLabel(
		"org.opencontainers.image.version",
		a.ReleaseVersion,
	)

	if a.Org == "ublue-os" {
		fedora = fedora.WithLabel(
			"io.artifacthub.package.logo-url",
			"https://avatars.githubusercontent.com/u/120078124?s=200&v=4",
		)
	}
	baseImage, err := fedora.BaseImage(ctx)
	if err == nil {
		fedora = fedora.WithLabel("org.opencontainers.image.base_image", baseImage)
	}

	baseImageVersion, err := fedora.BaseImageVersion(ctx)
	if err == nil {
		fedora = fedora.WithLabel(
			"org.opencontainers.image.base_image_version",
			baseImageVersion,
		)
	}

	for k, v := range labels {
		fedora = fedora.WithLabel(k, v)
	}

	return fedora, nil
}
