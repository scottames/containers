// A heavily opinionated custom Fedora Atomic container image
// primarily meant for my own personal use
//
// Can be used as an example for building using the Fedora (../fedora/) dagger
// module
package main

import (
	"context"
	"strings"
	"time"
)

// New initializes the Atomic Fedora module
func New(
	ctx context.Context,
	// Git repository root directory TODO: <- support running from anywhere
	//
	// referenced by Atomic.fedoraAtomic to determine the path to local files
	source *Directory,
	// Container registry
	// +optional
	// +default="quay.io"
	registry string,
	// Container registry organization
	// +optional
	// +default="fedora-ostree-desktops"
	org string,
	// Atomic variant
	// +optional
	// +default="silverblue"
	variant string,
	// Variant suffix string
	// e.g. main (as related to ublue-os images)
	// +optional
	suffix *string,
	// Tag or major release version
	// +optional
	// +default="40"
	tag string,
	// Labels to be applied to the generated container image in addition
	// to the default labels
	// +optional
	additionalLabels []string,
	// Optionally skip default labels
	// +optional
	// +default=false
	skipDefaultLabels bool,
) (*Atomic, error) {
	now := time.Now()
	a := &Atomic{
		Source:            source,
		Registry:          registry,
		Org:               org,
		Tag:               tag,
		Variant:           variant,
		Suffix:            suffix,
		Labels:            additionalLabels,
		Date:              strings.Replace(now.Format(time.DateOnly), "-", "", -1), // 20241031
		SkipDefaultLabels: skipDefaultLabels,
	}

	opts := FedoraContainerAddressOpts{}
	if suffix != nil {
		opts = FedoraContainerAddressOpts{Suffix: *suffix}
	}
	baseImage, err := dag.Fedora().
		ContainerAddress(ctx, registry, org, variant, tag, opts)
	if err != nil {
		return nil, err
	}
	a.BaseImage = baseImage
	baseImageCtr := dag.Container().From(a.BaseImage)
	a.BaseImageVersion, _ = containerImageVersionFromLabel(ctx, baseImageCtr)

	majorVersion, err := calculateContainerMajorVersion(ctx, baseImageCtr)
	// NOTE: major version skipped if not found! (not treated as error)
	if err == nil {
		a.MajorVersion = &majorVersion
	}

	a.Tags = a.defaultTags()

	return a, nil
}

// Atomic represents the Dagger module type
type Atomic struct {
	Source *Directory

	// Source container image
	Registry         string
	Org              string
	Tag              string
	Variant          string
	Suffix           *string
	BaseImage        string
	BaseImageVersion string

	// Generated atomic container image
	Digests      []string
	Labels       []string
	MajorVersion *string
	Date         string
	Tags         []string

	// Flags
	SkipDefaultLabels bool
}

// Container returns a Fedora Atomic container as a dagger.Container object
func (a *Atomic) Container(ctx context.Context) (*Container, error) {
	fedora, err := a.fedoraAtomic(ctx)
	if err != nil {
		return nil, err
	}

	return fedora.Container(), nil
}
