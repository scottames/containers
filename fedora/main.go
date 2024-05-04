// A generated module for Fedora functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
)

// TODO:
//  - [X] default to quay/fedora
//  - [X] add ublue option
//  - [ ] with registry
//    - [ ] option to remove upon return (publish, etc.) after packages installed
//  - [ ] with exec (script) (necessary?)
//  - [ ] with packages
//  - [ ] with overlay (i.e. Wolfi module)

// Fedora derived atomic image:
//
// dagger call -m fedora \
// --tag 40 \
// --variant silverblue \
// container terminal
//
// Universal Blue derived:
//
// dagger call -m fedora \
// --registry ghcr.io \
// --org ublue-os \
// --tag 40 \
// --variant silverblue \
// --suffix main \
// container terminal

func New(
	ctx context.Context,
	// Container registry
	// +optional
	// +default="quay.io"
	registry string,
	// Container registry organization
	// +optional
	// +default="fedora-ostree-desktops"
	org string,
	variant string,
	// +optional
	suffix *string,
	tag string,
) *Fedora {
	return &Fedora{
		Tag:      tag,
		Org:      org,
		Registry: registry,
		Suffix:   suffix,
		Variant:  variant,
	}
}

type Fedora struct {
	Tag      string
	Org      string
	Registry string
	Suffix   *string
	Variant  string
}

func (f Fedora) Container() *Container {
	const sourceStr = "%s/%s/%s%s:%s"

	suffix := ""
	if f.Suffix != nil {
		suffix = fmt.Sprintf("-%s", *f.Suffix)
	}

	return f.ContainerFrom(fmt.Sprintf(
		sourceStr,
		f.Registry,
		f.Org,
		f.Variant,
		suffix,
		f.Tag,
	))
}

func (f Fedora) ContainerFrom(from string) *Container {
	return dag.
		Container().
		From(from)
}
