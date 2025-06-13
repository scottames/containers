package main

import (
	"context"
	"dagger/sandbox/internal/dagger"
	"fmt"
	"strings"
)

// publish builds and publishes the Sandbox container image
func (s *Sandbox) publish(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	registry string,
	// name of the image
	imageName string,
	// additional tags to publish in addition to the default tags
	// +optional
	additionalTags []string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *dagger.Secret,
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
	// if true the "latest" tag will be published
	// +optional
	// +default=false
	latest bool,
	// timezone for the container
	// +optional
	// +default="UTC"
	tz string,
) (*Sandbox, error) {
	ctr := s.Container(ctx, tz)

	if secret != nil {
		// NOTE: the auth step MUST be bare registry w/o username namespace
		ctr = ctr.WithRegistryAuth(registry, username, secret)
	}

	if !skipRegistryNamespace {
		registry = strings.ToLower(fmt.Sprintf("%s/%s", registry, username))
	}

	ctr = ctr.WithLabel("org.opencontainers.image.title", imageName)

	tags := additionalTags
	if !skipDefaultTags {
		tags = append(tags, s.Tag)
	}
	if latest {
		tags = append(tags, "latest")
	}

	for _, tag := range tags {
		digest, err := ctr.Publish(
			ctx,
			fmt.Sprintf("%s/%s:%s", registry, imageName, tag),
		)
		if err != nil {
			return nil, err
		}
		s.Digests = append(s.Digests, digest)
	}

	return s, nil
}

// Publish build and publish the Sandbox container image
func (s *Sandbox) Publish(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	registry string,
	// name of the image
	imageName string,
	// additional tags to publish in addition to the default tags
	// +optional
	additionalTags []string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *dagger.Secret,
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
	// if true the "latest" tag will be published
	// +optional
	// +default=false
	latest bool,
	// timezone for the container
	// +optional
	// +default="UTC"
	tz string,
) ([]string, error) {
	_, err := s.publish(
		ctx,
		registry,
		imageName,
		additionalTags,
		username,
		secret,
		skipRegistryNamespace,
		skipDefaultTags,
		latest,
		tz,
	)
	if err != nil {
		return nil, err
	}

	return s.Digests, nil
}

// PublishAndSign build, publish, and sign (via cosign)
// the Sandbox container image
func (s *Sandbox) PublishAndSign(
	ctx context.Context,
	// registry url, e.g. ghcr.io
	registry string,
	// name of the image
	imageName string,
	// additional tags to publish in addition to the default tags
	// +optional
	additionalTags []string,
	// registry username
	// also used as the registry namespace
	username string,
	// registry auth password/secret
	// +optional
	secret *dagger.Secret,
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
	// if true the "latest" tag will be published
	// +optional
	// +default=false
	latest bool,
	// timezone for the container
	// +optional
	// +default="UTC"
	tz string,
) ([]string, error) {
	_, err := s.publish(
		ctx,
		registry,
		imageName,
		additionalTags,
		username,
		secret,
		skipRegistryNamespace,
		skipDefaultTags,
		latest,
		tz,
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
		s.Digests,
		opts,
	)
	if err != nil {
		return nil, err
	}

	output := append([]string{"Published:"}, s.Digests...)
	output = append(output, "")
	output = append(output, cosignStdout...)

	return output, nil
}
