// A toolbox (distrobox) container intended to be a controlled sandbox for who
// knows what! (AI)

package main

import (
	"context"
)

type Sandbox struct {
	Registry string
	Org      *string
	Image    string
	Suffix   *string
	Tag      string

	Digests []string
}

func New(
	ctx context.Context,
	// Container registry
	// +optional
	// +default="docker.io"
	registry string,
	// Container registry organization
	// +optional
	org *string,
	// Container image name
	// +optional
	// +default="node"
	image string,
	// Variant suffix string
	// +optional
	suffix *string,
	// Tag or major release version
	// +optional
	// +default="20"
	tag string,
) *Sandbox {
	return &Sandbox{
		Registry: registry,
		Org:      org,
		Image:    image,
		Suffix:   suffix,
		Tag:      tag,
	}
}
