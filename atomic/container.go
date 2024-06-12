package main

import (
	"context"
	"fmt"
	"strings"
)

// containerImageVersionFromLabel returns the label value for the image version, defined as
// org.opencontainers.image.version and a possible error
func containerImageVersionFromLabel(ctx context.Context, ctr *Container) (string, error) {
	versionLabel := "org.opencontainers.image.version"
	return ctr.Label(ctx, versionLabel)
}

func calculateContainerMajorVersion(ctx context.Context, ctr *Container) (string, error) {
	version, err := containerImageVersionFromLabel(ctx, ctr)
	if err != nil {
		return "", err
	}
	majorVersion := strings.Split(version, ".")
	if len(majorVersion) <= 0 {
		return "", fmt.Errorf(
			"unable to determine major version from base image version: %s",
			version,
		)
	}

	return majorVersion[0], nil
}
