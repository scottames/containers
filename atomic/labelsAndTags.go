package main

import (
	"context"
	"fmt"
	"strings"
)

// TODO: docs
func (a *Atomic) fedoraWithLabelsFromCLI(ctx context.Context, fedora *Fedora) (*Fedora, error) {
	for _, l := range a.Labels {
		ll := strings.SplitN(l, "=", 2)
		if len(ll) < 2 {
			return nil, fmt.Errorf("invalid label: %s", ll)
		}

		fedora = fedora.WithLabel(FedoraWithLabelOpts{
			Name:  ll[0],
			Value: ll[1],
		})
	}

	return fedora, nil
}

// TODO: docs
func (a *Atomic) fedoraWithDefaultLabels(ctx context.Context, fedora *Fedora) *Fedora {
	version := a.Date
	if a.MajorVersion != nil {
		version = fmt.Sprintf("%s-%s", *a.MajorVersion, a.Date)
	}

	fedora = fedora.WithLabel(FedoraWithLabelOpts{
		Name:  "org.opencontainers.image.version",
		Value: version, // note: universal blue appends a build number, we do not
	})

	if a.Org == "ublue-os" {
		fedora = fedora.WithLabel(FedoraWithLabelOpts{
			Name:  "io.artifacthub.package.logo-url",
			Value: "https://avatars.githubusercontent.com/u/120078124?s=200&v=4",
		})
		fedora = fedora.WithLabel(FedoraWithLabelOpts{
			Name:  "org.opencontainers.image.base_image",
			Value: a.BaseImage,
		})
		fedora = fedora.WithLabel(FedoraWithLabelOpts{
			Name:  "org.opencontainers.image.base_image_version",
			Value: a.BaseImageVersion,
		})

	}

	for k, v := range labels {
		fedora = fedora.WithLabel(FedoraWithLabelOpts{
			Name:  k,
			Value: v,
		})
	}

	return fedora
}

func (a *Atomic) defaultTags() []string {
	tags := []string{}
	if a.MajorVersion != nil {
		tags = append(tags,
			*a.MajorVersion,
			fmt.Sprintf("%s-%s", *a.MajorVersion, a.Date),
		)

		if *a.MajorVersion == latestFedoraVersion {
			tags = append(tags, "latest")
		}
	}

	tags = append(tags, a.Date)

	return tags
}
