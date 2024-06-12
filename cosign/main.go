// Cosign container image signing
package main

import (
	"context"
)

// Cosign represents the cosign Dagger module type
type Cosign struct{}

// Sign will run cosign from the image, as defined by the cosignImage
// parameter, to sign the given Container image digests
//
// Note: keyless signing not supported as-is
//
// See https://edu.chainguard.dev/open-source/sigstore/cosign/an-introduction-to-cosign/
func (f *Cosign) Sign(
	ctx context.Context,
	// Cosign private key
	privateKey Secret,
	// Cosign password
	password Secret,
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
	// Container image digests to sign
	digests ...string,
) ([]string, error) {
	stdouts := []string{}
	for _, d := range digests {
		cmd := []string{"sign", d, "--key", "env://COSIGN_PRIVATE_KEY"}
		if registryUsername != nil && registryPassword != nil {
			pwd, err := registryPassword.Plaintext(ctx)
			if err != nil {
				return nil, err
			}

			cmd = append(
				cmd,
				"--registry-username",
				*registryUsername,
				"--registry-password",
				pwd,
			)
		}
		cosign := dag.
			Container().
			From(*cosignImage).
			WithUser(*cosignUser).
			WithEnvVariable("COSIGN_YES", "true").
			WithSecretVariable("COSIGN_PASSWORD", &password).
			WithSecretVariable("COSIGN_PRIVATE_KEY", &privateKey).
			WithExec(cmd)

		if dockerConfig != nil {
			cosign = cosign.WithMountedFile(
				"/home/nonroot/.docker/config.json",
				dockerConfig,
				ContainerWithMountedFileOpts{Owner: *cosignUser})
		}

		stdout, err := cosign.Stdout(ctx)
		if err != nil {
			return nil, err
		}

		stdouts = append(stdouts, stdout)
	}

	return stdouts, nil
}
