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
	return &Atomic{
		Source:   source,
		Registry: registry,
		Org:      org,
		Tag:      tag,
		Variant:  variant,
		Suffix:   suffix,
	}
}

// TODO: docs
type Atomic struct {
	Source   *Directory
	Registry string
	Org      string
	Tag      string
	Variant  string
	Suffix   *string
	Digests  []string
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
	// the tag(s) to publish to
	tags []string,
	// registry auth username
	// +optional
	username *string,
	// registry auth password/secret
	// +optional
	secret *Secret,
) (*Atomic, error) {
	var publishOpts FedoraPublishOpts
	if username != nil || secret != nil {
		publishOpts = FedoraPublishOpts{
			Username: *username,
			Secret:   secret,
		}
	}

	digests, err := a.
		fedoraAtomic().
		Publish(ctx, registry, imageName, tags, publishOpts)
	if err != nil {
		return nil, err
	}

	a.Digests = append(a.Digests, digests...)

	return a, nil
}

// TODO: docs
func (a *Atomic) Sign(
	ctx context.Context,
	// Cosign private key
	privateKey Secret,
	// Cosign password
	password Secret,
	// Docker image digest
	//+optional
	digest *string,
	// registry username
	//+optional
	registryUsername *string,
	// name of the image
	//+optional
	registryPassword *Secret,
	// Docker config
	//+optional
	dockerConfig *File,
	// Cosign container image
	//+optional
	//+default="chainguard/cosign:latest"
	cosignImage *string,
	// Cosign container image user
	//+optional
	//+default="nonroot"
	cosignUser *string,
) ([]string, error) {
	ds := a.Digests
	if digest != nil {
		ds = []string{*digest}
	}

	opts := FedoraSignOpts{
		CosignImage: *cosignImage,
		CosignUser:  *cosignUser,
	}
	if registryUsername != nil && registryPassword != nil {
		opts.RegistryUsername = *registryUsername
		opts.RegistryPassword = registryPassword
	}
	if dockerConfig != nil {
		opts.DockerConfig = dockerConfig
	}

	stdouts := []string{}
	for _, d := range ds {
		opts.Digest = d
		var err error
		stdouts, err = dag.Fedora().Sign(
			ctx,
			&privateKey,
			&password,
			opts,
		)
		if err != nil {
			return nil, err
		}
	}

	return stdouts, nil
}
