// A heavily opinionated custom Fedora Atomic container image
// primarily meant for my own personal use
//
// Can be used as an example for building using the Fedora (../fedora/) dagger
// module
package main

import (
	"context"
)

// New initializes the Atomic Fedora module

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
	// Git repository root directory
	//
	// referenced by Atomic.fedoraAtomic to determine the path to local files
	source *Directory,
) *Atomic {
	opts := FedoraOpts{
		Registry: registry,
		Org:      org,
		Tag:      tag,
		Variant:  variant,
	}
	if suffix != nil {
		opts.Suffix = *suffix
	}

	return &Atomic{
		Fedora: dag.
			Fedora(opts),
	}
}

// TODO: docs
type Atomic struct {
	Fedora *Fedora
	Source *Directory
}

// TODO:
//   - [ ] Docs
//   - Cosign
//   - [X] Copy file
//   - [ ] Sign image
//   - [ ] Anything else in Containerfile?
//   - [ ] Anything else in build.sh?
//   - [ ] "container commit" ? (verify with template ublue)
//   - [ ] Re-work github action build.yaml into new one here...
//   - [ ] All other TODOs
//
// TODO: maybe...
//   - [ ] Copy --from image (akmods) => Not used today... only needed for Framework battery?
func (a *Atomic) Container() *Container {
	return a.fedoraAtomic().Container()
}

func (a *Atomic) Publish(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	registry string,
	// name of the image
	imageName string,
	// the tag to publish to
	tag string,
	// registry auth username
	// +optional
	username *string,
	// registry auth password/secret
	// +optional
	secret *Secret,
) (string, error) {
	var publishOpts FedoraPublishOpts
	if username != nil || secret != nil {
		publishOpts = FedoraPublishOpts{
			Username: *username,
			Secret:   secret,
		}
	}

	return a.
		fedoraAtomic().
		Publish(ctx, registry, imageName, tag, publishOpts)
}
