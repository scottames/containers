package main

import (
	"context"
	"fmt"
)

// Publishes this container as a new image to the specified address.
//
// Publish returns a fully qualified ref.
//
// It can also publish platform variants.
//
// Examples usage:
//
//	dagger call -m fedora \
//		--tag 40 \
//		--variant silverblue \
//	with-packages-installed \
//		--packages=curl \
//	publish \
//		--registry=ghcr.io \
//		--image-name=my-image \
//		--username=$REGISTRY_USERNAME \
//		--secret=env:REGISTRY_SECRET \
//		--tag=latest
func (f *Fedora) Publish(
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
	suffix := ""
	if f.Suffix != nil {
		suffix = fmt.Sprintf("-%s", *f.Suffix)
	}

	const sourceStr = "%s/%s/%s%s:%s" // registry/org/variant+suffix:tag
	ctr, err := f.
		ContainerFrom(
			ctx,
			fmt.Sprintf(sourceStr,
				f.Registry,
				f.Org,
				f.Variant,
				suffix,
				f.Tag,
			),
		)
	if err != nil {
		return "", err
	}

	if username != nil && secret != nil {
		ctr.WithRegistryAuth(registry, *username, secret)
	}

	ctr = ctr.WithLabel("org.opencontainers.image.title", imageName)

	// TODO: verify these labels carry over!
	// org.opencontainers.image.version=${{ steps.labels.outputs.VERSION }}
	// io.artifacthub.package.logo-url=https://avatars.githubusercontent.com/u/120078124?s=200&v=4
	//
	// TODO: do this with https://github.com/docker/metadata-action?tab=readme-ov-file#outputs or here?
	// "org.opencontainers.image.created": "2024-06-01T15:07:50.585Z",
	// "org.opencontainers.image.description": "A base Universal Blue silverblue image with batteries included",
	// "org.opencontainers.image.licenses": "Apache-2.0",
	// "org.opencontainers.image.revision": "c8d9b00faefec18b2476b10de1be46f496524023",
	// "org.opencontainers.image.source": "https://github.com/ublue-os/main",
	// "org.opencontainers.image.title": "silverblue-main",
	// "org.opencontainers.image.url": "https://github.com/ublue-os/main",
	// "org.opencontainers.image.version": "40.20240601.0",

	image, err := ctr.Publish(
		ctx,
		fmt.Sprintf("%s/%s:%s", registry, imageName, tag),
	)
	if err != nil {
		return "", err
	}

	return image, nil
}
