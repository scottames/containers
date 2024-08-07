package main

import (
	"context"
	"dagger/atomic/internal/dagger"
	"fmt"
	"strings"
)

// publish builds and publishes the Fedora Atomic container image
func (a *Atomic) publish(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	imageRegistry string,
	// name of the image
	imageName string,
	// repository name, if different from imageName
	// +optional
	repository *string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *dagger.Secret,
	// additional tags to publish in addition to the default tags
	// default tags will be included unless skipDefaultTags is set:
	//  [majorVersion, majorVersion-date, date]
	// +optional
	additionalTags []string,
	// skip opinionated ublue-way of setting up signing config
	//   note: if basing off of ublue, this is already setup,
	//         but not for the source image
	// +optional
	// +default=false
	skipSigningConfig bool,
	// skip namespacing registry with username
	//   example:
	//     registry=ghcr.io username=foo
	//     default: => ghcr.io/foo/image
	//     skip:    => ghcr.io/image
	// +optional
	// +default=false
	skipRegistryNamespace bool,
	// skip adding default tags
	// +optional
	// +default=false
	skipDefaultTags bool,
) (*Atomic, error) {
	ctr, err := a.Container(ctx)
	if err != nil {
		return nil, err
	}

	if secret != nil {
		// NOTE: the auth step MUST be bare registry w/o username namespace
		ctr = ctr.WithRegistryAuth(imageRegistry, username, secret)
	}

	if !skipRegistryNamespace {
		imageRegistry = strings.ToLower(fmt.Sprintf("%s/%s", imageRegistry, username))
	}

	if repository == nil {
		repository = &imageName
	}

	if !skipSigningConfig {
		ctr = a.ctrSigningConfig(
			ctr,
			*repository,
			imageRegistry,
			imageName,
			a.ReleaseVersion,
		)
	}

	ctr = ctr.WithLabel("org.opencontainers.image.title", imageName).
		// NOTE: this must be the last thing to run prior to publishing
		WithExec([]string{"ostree", "container", "commit"})

	tags := additionalTags
	if !skipDefaultTags {
		tags = append(tags, a.Tags...)
	}

	// TODO: better to  do this similar to multi-stage passing multiple
	// containers to publish?
	for _, tag := range tags {
		digest, err := ctr.Publish(
			ctx,
			fmt.Sprintf("%s/%s:%s", imageRegistry, imageName, tag),
		)
		if err != nil {
			return nil, err
		}
		a.Digests = append(a.Digests, digest)
	}

	return a, nil
}

// Publish build and publish the Fedora atomic container image
func (a *Atomic) Publish(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	imageRegistry string,
	// name of the image
	imageName string,
	// repository name, if different from imageName
	// +optional
	repository *string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *dagger.Secret,
	// additional tags to publish in addition to the default tags
	// default tags will be included unless skipDefaultTags is set:
	//  [majorVersion, majorVersion-date, date]
	// +optional
	additionalTags []string,
	// skip opinionated ublue-way of setting up signing config
	//   note: if basing off of ublue, this is already setup,
	//         but not for the source image
	// +optional
	// +default=false
	skipSigningConfig bool,
	// skip namespacing registry with username
	//   example:
	//     registry=ghcr.io username=foo
	//     default: => ghcr.io/foo/image
	//     skip:    => ghcr.io/image
	// +optional
	// +default=false
	skipRegistryNamespace bool,
	// skip adding default tags
	// +optional
	// +default=false
	skipDefaultTags bool,
) ([]string, error) {
	_, err := a.publish(
		ctx,
		imageRegistry,
		imageName,
		repository,
		username,
		secret,
		additionalTags,
		skipSigningConfig,
		skipRegistryNamespace,
		skipDefaultTags,
	)
	if err != nil {
		return nil, err
	}

	return a.Digests, nil
}

// PublishAndSign build, publish, and sign (via cosign)
// the Fedora container image
func (a *Atomic) PublishAndSign(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	imageRegistry string,
	// name of the image
	imageName string,
	// repository name, if different from imageName
	// +optional
	repository *string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *dagger.Secret,
	// additional tags to publish in addition to the default tags
	// default tags will be included unless skipDefaultTags is set:
	//  [majorVersion, majorVersion-date, date]
	// +optional
	additionalTags []string,
	// skip opinionated ublue-way of setting up signing config
	//   note: if basing off of ublue, this is already setup,
	//         but not for the source image
	// +optional
	// +default=false
	skipSigningConfig bool,
	// skip namespacing registry with username
	//   example:
	//     registry=ghcr.io username=foo
	//     default: => ghcr.io/foo/image
	//     skip:    => ghcr.io/image
	// +optional
	// +default=false
	skipRegistryNamespace bool,
	// skip adding default tags
	// +optional
	// +default=false
	skipDefaultTags bool,
	// Cosign private key
	cosignPrivateKey dagger.Secret,
	// Cosign password
	cosignPassword dagger.Secret,
	// Docker config
	//+optional
	dockerConfig *dagger.File,
	// Cosign container image to be used to sign the digests
	// +optional
	// +default="chainguard/cosign:latest"
	cosignImage *string,
	// Cosign container image user
	// +optional
	// +default="nonroot"
	cosignUser *string,
) ([]string, error) {
	_, err := a.publish(
		ctx,
		imageRegistry,
		imageName,
		repository,
		username,
		secret,
		additionalTags,
		skipSigningConfig,
		skipRegistryNamespace,
		skipDefaultTags,
	)
	if err != nil {
		return nil, err
	}

	opts := dagger.CosignSignOpts{
		// Should never be nil due to Dagger setting default values
		CosignImage: *cosignImage,
		CosignUser:  *cosignUser,
	}
	if secret != nil {
		opts.RegistryUsername = username
		opts.RegistryPassword = secret
	}
	if dockerConfig != nil {
		opts.DockerConfig = dockerConfig
	}

	cosignStdout, err := dag.Cosign().Sign(
		ctx,
		&cosignPrivateKey,
		&cosignPassword,
		a.Digests,
		opts,
	)
	if err != nil {
		return nil, err
	}

	output := append([]string{"Published:"}, a.Digests...)
	output = append(output, "")
	output = append(output, cosignStdout...)

	return output, nil
}
