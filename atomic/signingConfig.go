package main

import (
	"dagger/atomic/internal/dagger"
	"fmt"
)

// ctrSigningConfig updates the universal-blue-esk signing config
func (a *Atomic) ctrSigningConfig(
	ctr *dagger.Container,
	repo string,
	imageRegistry string,
	imageName string,
	imageVersion string,
) *dagger.Container {
	imageInfo := fmt.Sprintf(`{
  "image-ref": "ostree-image-signed:docker://%s/%s",
  "image-tag": "%s"
}`, imageRegistry, imageName, imageVersion)
	registriesD := fmt.Sprintf("/usr/etc/containers/registries.d/%s.yaml", imageName)

	yq := dag.Container().From("docker.io/mikefarah/yq")
	cosignPubKeyPath := fmt.Sprintf("/usr/etc/pki/containers/%s.pub", imageName)

	return ctr.
		WithFile("/usr/bin/yq", yq.File("/usr/bin/yq")).
		WithFile(cosignPubKeyPath, a.Source.File("cosign.pub")).
		WithNewFile(
			"/usr/share/ublue-os/image-info.json",
			imageInfo,
			dagger.ContainerWithNewFileOpts{
				Permissions: 0644, Owner: "root",
			},
		).
		WithNewFile(
			registriesD,
			fmt.Sprintf(`docker:
  %s:
      use-sigstore-attachments: true
`, imageRegistry),
			dagger.ContainerWithNewFileOpts{
				Permissions: 0644,
				Owner:       "root",
			}).
		WithExec([]string{
			"yq", "-i", "-o=j",
			fmt.Sprintf(`.transports.docker |=
{"%s/%s": [
		{
			"type": "sigstoreSigned",
			"keyPath": "%s",
			"signedIdentity": {
				"type": "matchRepository"
			}
		}
	]
}

+ .`, imageRegistry, repo, cosignPubKeyPath), // TODO: git repo instead
			"/usr/etc/containers/policy.json",
		})
}
