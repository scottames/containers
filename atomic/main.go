// # Fedora Atomic
//
// A custom Fedora Atomic container image dagger module heavily customized for
// my personal use.
//
// NOTE:
// Interested in making your own? Create an issue, we can work
// together to build a proper template. Or just fork and yolo!
//
// The heavy lifting behind Fedora Atomic images is done by the Fedora and
// Universal Blue communities.
//
// - Fedora Atomic: https://fedoraproject.org/atomic-desktops/)
// - Universal Blue: https://universal-blue.org)
package main

import (
	"context"
	"dagger/atomic/internal/dagger"
	"fmt"
	"slices"
	"strings"
)

func New(
	ctx context.Context,
	// Git repository root directory
	// referenced by Atomic.fedoraAtomic to determine the path to local files
	source *dagger.Directory,
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
	sourceFiles, err := source.Glob(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("unable to read source files: %w", err)
	}

	if !slices.Contains(sourceFiles, "atomic/") {
		return nil, fmt.Errorf(
			"Please run from the root of the git repository, 'atomic' not found in: %s",
			strings.Join(sourceFiles, ", "),
		)
	}

	a := &Atomic{
		Source:            source,
		Registry:          registry,
		Org:               org,
		Tag:               tag,
		Variant:           variant,
		Suffix:            suffix,
		Labels:            additionalLabels,
		SkipDefaultLabels: skipDefaultLabels,
	}

	return a, nil
}

// Atomic represents the Dagger module type
type Atomic struct {
	Source *dagger.Directory

	// Source container image
	Registry string
	Org      string
	Tag      string
	Variant  string
	Suffix   *string
	// BaseImage        string
	// BaseImageVersion string

	// Generated atomic container image
	Digests []string
	Labels  []string
	// MajorVersion *string
	// Date string
	Tags           []string
	ReleaseVersion string

	// Flags
	SkipDefaultLabels bool
}

// Container returns a Fedora Atomic container as a dagger.Container object
func (a *Atomic) Container(ctx context.Context) (*dagger.Container, error) {
	fedora, err := a.fedoraAtomic(ctx)
	if err != nil {
		return nil, err
	}

	return fedora.Container(), nil
}
