package main

import "fmt"

// ctrSigningConfig updates the universal-blue-esk signing config
func (a *Atomic) ctrSigningConfig(
	ctr *Container,
	imageName string,
	imageRegistry string,
	version string,
) *Container {
	imageInfo := fmt.Sprintf(`{
  "image-ref": "ostree-image-signed:docker://%s/%s",
  "image-tag": "%s"
		}`, imageRegistry, imageName, version)
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
			Contents: fmt.Sprintf(
				"docker:\\n  %s:\\n    use-sigstore-attachments: true\\n",
				imageRegistry,
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
+ .`, imageRegistry, imageName, imageName),
			"/usr/etc/containers/policy.json",
		})
}
