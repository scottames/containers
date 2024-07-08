package main

import "fmt"

// ctrSigningConfig updates the universal-blue-esk signing config
func (a *Atomic) ctrSigningConfig(
	ctr *Container,
	username string,
	repo string,
	imageRegistry string,
	imageName string,
	imageVersion string,
) *Container {
	imageInfo := fmt.Sprintf(`{
  "image-ref": "ostree-image-signed:docker://%s/%s",
  "image-tag": "%s"
		}`, imageRegistry, imageName, imageVersion)
	registriesD := fmt.Sprintf("/usr/etc/containers/registries.d/%s.yaml", imageName)

	yq := dag.Container().From("docker.io/mikefarah/yq")

	return ctr.
		WithFile("/usr/bin/yq", yq.File("/usr/bin/yq")).
		WithFile(fmt.Sprintf("/usr/etc/pki/containers/%s.pub", imageName), a.Source.File("cosign.pub")).
		WithNewFile(
			"/usr/share/ublue-os/image-info.json",
			ContainerWithNewFileOpts{
				Contents: imageInfo, Permissions: 0644, Owner: "root",
			},
		).
		WithNewFile(registriesD, ContainerWithNewFileOpts{
			Contents: fmt.Sprintf(`docker:
  %s/%s:
      use-sigstore-attachments: true
`,
				imageRegistry, username,
			),
			Permissions: 0644,
			Owner:       "root",
		}).
		WithExec([]string{
			"yq", "-i", "-o=j",
			fmt.Sprintf(`.transports.docker |=
{"%s/%s": [
		{
			"type": "sigstoreSigned",
			"keyPath": "/usr/etc/pki/containers/%s.pub",
			"signedIdentity": {
				"type": "matchRepository"
			}
		}
	]
}

+ .`, imageRegistry, repo, imageName), // TODO: git repo instead
			"/usr/etc/containers/policy.json",
		})
}
