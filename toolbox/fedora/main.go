// A heavily opinionated custom Fedora container image meant for use with toolbx
// or distrobox
package main

import (
	"context"
	"fmt"
)

var (
	distroboxHostExecSymlinks = []string{
		"docker",
		"flatpak",
		"hostnamectl",
		"loginctl",
		"nmcli",
		"op",
		"podman",
		"rpm-ostree",
		"systemctl",
		"transactional-update",
		"xclip",
		"xdg-open",
	}
	// renovate: datasource=github-releases depName=danielmiessler/fabric
	fabricVersion = "1.4.0"
	fabricGit     = "git+https://github.com/danielmiessler/fabric.git"
	labels        = map[string]string{
		"usage":   "This image is meant to be used with the toolbox or distrobox command",
		"summary": "A cloud-native terminal experience powered by Fedora",

		"com.github.containers.toolbox": "true",
	}
	packageUrlsWithReleaseVersion = []string{
		"https://download1.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-%s.noarch.rpm",
		"https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-%s.noarch.rpm",
	}
	packageGroups = []string{"Development Tools"}
	packages      = []string{
		"adw-gtk3-theme",
		"awscli",
		"bash-completion",
		"bc",
		"bzip2",
		"curl",
		"diffutils",
		"dnf-plugins-core",
		"findutils",
		"fish",
		"git",
		"glibc-all-langpacks",
		"glibc-locale-source",
		"gnupg2",
		"gnupg2-smime",
		"hostname",
		"iproute",
		"iputils",
		"keyutils",
		"krb5-libs",
		"less",
		"libxcrypt-compat",
		"lsof",
		"man-db",
		"man-pages",
		"mtr",
		"netcat",
		"ncurses",
		"nodejs",
		"nodejs-npm",
		"nss-mdns",
		"nvidia-vaapi-driver",
		"openssh-clients",
		"pam",
		"passwd",
		"pciutils",
		"pinentry",
		"pipx",
		"procps-ng",
		"ripgrep",
		"rsync",
		"shadow-utils",
		"sudo",
		"tcpdump",
		"time",
		"traceroute",
		"tree",
		"tzdata",
		"unzip",
		"util-linux",
		"vim",
		"wget",
		"which",
		"whois",
		"wl-clipboard",
		"words",
		"xz",
		"zenity",
		"zip",
		"zsh",

		// for fabric: https://github.com/danielmiessler/fabric
		"gcc-c++",
		"python3-devel",

		"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_64bit/session-manager-plugin.rpm",
	}
)

type FedoraToolbox struct {
	Registry       string
	Org            *string
	Image          string
	Suffix         *string
	Tag            string
	ReleaseVersion string

	Digests []string
}

func New(
	ctx context.Context,
	// Container registry
	// +optional
	// +default="registry.fedoraproject.org"
	registry string,
	// Container registry organization
	// +optional
	org *string,
	// Container image name
	// +optional
	// +default="fedora-toolbox"
	image string,
	// Variant suffix string
	// +optional
	suffix *string,
	// Tag or major release version
	// +optional
	// +default="40"
	tag string,
) *FedoraToolbox {
	return &FedoraToolbox{
		Registry: registry,
		Org:      org,
		Image:    image,
		Suffix:   suffix,
		Tag:      tag,
	}
}

// fedora returns the dagger.Fedora object with the current container context
// associated
func (ft *FedoraToolbox) fedora(ctx context.Context) *Fedora {
	opts := FedoraOpts{
		Registry: ft.Registry,
		Variant:  ft.Image,
		Tag:      ft.Tag,
	}
	if ft.Org != nil {
		opts.Org = *ft.Org
	}

	if ft.Suffix != nil {
		opts.Suffix = *ft.Suffix
	}

	fedora := dag.Fedora(opts)
	var err error

	ft.ReleaseVersion, err = fedora.ContainerReleaseVersionFromLabel(ctx)
	if err != nil || len(ft.ReleaseVersion) <= 0 {
		ft.ReleaseVersion = ft.Tag
	}

	return ft.distroboxHostExecSymlinks(fedora)
}

// Container returns the Fedora toolbx/distrobox dagger.Container
func (ft *FedoraToolbox) Container(ctx context.Context) (*Container, error) {
	fedora := ft.fedora(ctx)

	for n, v := range labels {
		fedora = fedora.WithLabel(n, v)
	}

	for _, s := range packageUrlsWithReleaseVersion {
		packages = append(packages, fmt.Sprintf(s, ft.ReleaseVersion))
	}

	db := dag.Distrobox()
	dbheFile := db.HostExecFile()
	hostSpawn := db.HostSpawnFile()

	ctr := fedora.
		WithPackagesInstalled(packages).
		WithPackageGroupsInstalled(packageGroups).
		WithPackagesSwapped("mesa-va-drivers", "mesa-va-drivers-freeworld").
		WithPackagesSwapped("mesa-vdpau-drivers", "mesa-vdpau-drivers-freeworld").
		Container(). // ✨ type becomes dagger.Container here!
		WithFile(
			"/usr/bin/distrobox-host-exec",
			dbheFile,
		).
		WithFile("/usr/bin/host-spawn", hostSpawn,
			ContainerWithFileOpts{Permissions: 0755, Owner: "root"},
		).
		WithExec([]string{
			"pipx", "install", "--global",
			fmt.Sprintf("%s@%s", fabricGit, fabricVersion),
		}).
		WithExec([]string{"dnf", "clean", "all"})

	return ctr, nil
}

func (ft *FedoraToolbox) distroboxHostExecSymlinks(fedora *Fedora) *Fedora {
	for _, link := range distroboxHostExecSymlinks {
		fedora = fedora.WithExec([]string{
			"ln", "-fs", "/usr/bin/distrobox-host-exec",
			fmt.Sprintf("/usr/local/bin/%s", link),
		}, false)
	}

	return fedora
}
