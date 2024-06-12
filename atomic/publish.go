package main

import (
	"context"
	"fmt"
	"strings"
)

// publish builds and publishes the Fedora Atomic container image
func (a *Atomic) publish(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	registry string,
	// name of the image
	imageName string,
	// additional tags to publish in addition to the default tags
	// default tags will be included unless skipDefaultTags is set:
	//  [majorVersion, majorVersion-date, date]
	// +optional
	additionalTags []string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *Secret,
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
		ctr = ctr.WithRegistryAuth(registry, username, secret)
	}

	if !skipRegistryNamespace {
		registry = strings.ToLower(fmt.Sprintf("%s/%s", registry, username))
	}

	if !skipSigningConfig {
		version := "latest"
		if a.MajorVersion != nil {
			version = *a.MajorVersion
		} else if len(additionalTags) > 0 {
			version = additionalTags[0]
		}
		ctr = a.ctrSigningConfig(ctr, imageName, registry, version)
	}

	ctr = ctr.WithLabel("org.opencontainers.image.title", imageName).
		// NOTE: this must be the last thing to run prior to publishing
		WithExec([]string{"ostree", "container", "commit"})

	tags := additionalTags
	if !skipDefaultTags {
		tags = append(tags, a.Tags...)
	}

	for _, tag := range tags {
		digest, err := ctr.Publish(
			ctx,
			fmt.Sprintf("%s/%s:%s", registry, imageName, tag),
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
	registry string,
	// name of the image
	imageName string,
	// additional tags to publish in addition to the default tags
	// default tags will be included unless skipDefaultTags is set:
	//  [majorVersion, majorVersion-date, date]
	// +optional
	additionalTags []string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *Secret,
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
		registry,
		imageName,
		additionalTags,
		username,
		secret,
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
	registry string,
	// name of the image
	imageName string,
	// additional tags to publish in addition to the default tags
	// default tags will be included unless skipDefaultTags is set:
	//  [majorVersion, majorVersion-date, date]
	// +optional
	additionalTags []string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *Secret,
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
	cosignPrivateKey Secret,
	// Cosign password
	cosignPassword Secret,
	// Docker config
	//+optional
	dockerConfig *File,
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
		registry,
		imageName,
		additionalTags,
		username,
		secret,
		skipSigningConfig,
		skipRegistryNamespace,
		skipDefaultTags,
	)
	if err != nil {
		return nil, err
	}

	opts := CosignSignOpts{
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
