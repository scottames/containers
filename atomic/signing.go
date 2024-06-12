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
		WithExec(
			[]string{
				// assumes cosign.pub already on image (see a.fedoraAtomic())
				"cp", "/usr/share/ublue-os/cosign.pub",
				fmt.Sprintf("/usr/etc/pki/containers/%s.pub", imageName),
			},
		).
		WithNewFile(
			"/usr/share/ublue-os/image-info.json",
			ContainerWithNewFileOpts{
				Contents: imageInfo, Permissions: 0644, Owner: "root",
			},
		).
		WithExec([]string{
			"cp", "/usr/etc/containers/registries.d/ublue-os.yaml",
			registriesD,
		}).
		WithExec(
			[]string{
				"sed", "-i",
				// assumes the source container
				fmt.Sprintf("s %s/%s %s g", a.Registry, a.Org, imageRegistry),
				registriesD,
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
