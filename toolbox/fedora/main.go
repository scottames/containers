// A heavily opinionated custom Fedora container image meant for use with toolbx
// or distrobox
package main

import (
	"context"
	"dagger/toolbox-fedora/internal/dagger"
	"fmt"
)

var (
	labels = map[string]string{
		"usage":   "This image is meant to be used with the toolbox or distrobox command",
		"summary": "A cloud-native terminal experience powered by Fedora",

		"com.github.containers.toolbox": "true",
	}
	reposForBuild = []string{
		"https://copr.fedorainfracloud.org/coprs/scottames/mise/repo/fedora-FEDORA_MAJOR_VERSION/scottames-mise-fedora-FEDORA_MAJOR_VERSION.repo",
	}
	packageUrlsWithReleaseVersion = []string{
		"https://download1.rpmfusion.org/nonfree/fedora/rpmfusion-nonfree-release-%s.noarch.rpm",
		"https://download1.rpmfusion.org/free/fedora/rpmfusion-free-release-%s.noarch.rpm",
	}
	packageGroups = []string{"development-tools"}
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
		"mise",
		"mtr",
		"netcat",
		"ncurses",
		"nodejs",
		"nodejs-npm",
		"nss-mdns",
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

		// to build xwayland-satellite (Niri)
		"clang-libs",
		"xcb-util-cursor-devel",

		// for fabric: https://github.com/danielmiessler/fabric
		"gcc-c++",
		"python3-devel",

		// for guarddog (security tool - needs pygit2)
		"libgit2-devel",

		"https://s3.amazonaws.com/session-manager-downloads/plugin/latest/linux_64bit/session-manager-plugin.rpm",
	}

	// Security vetting tools installed via pipx --global
	pipxSecurityTools = []string{
		"semgrep",
		"bandit",
		"pip-audit",
		"guarddog",
	}

	// Security vetting tools installed via cargo (built in separate container)
	cargoSecurityTools = []string{
		"cargo-audit",
		"cargo-deny",
		"cargo-geiger",
	}

	// Security vetting tools installed via npm -g
	npmSecurityTools = []string{
		"socket",
	}
)

type FedoraToolbox struct {
	Source         *dagger.Directory
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
	// Source directory containing scripts and config files
	// +optional
	source *dagger.Directory,
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
	tag string,
) *FedoraToolbox {
	return &FedoraToolbox{
		Source:   source,
		Registry: registry,
		Org:      org,
		Image:    image,
		Suffix:   suffix,
		Tag:      tag,
	}
}

// fedora returns the dagger.Fedora object with the current container context
// associated
func (ft *FedoraToolbox) fedora(ctx context.Context) *dagger.Fedora {
	opts := dagger.FedoraOpts{
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

	return fedora
}

// Container returns the Fedora toolbx/distrobox dagger.Container
func (ft *FedoraToolbox) Container(ctx context.Context) (*dagger.Container, error) {
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

	finalReposForBuild := replaceStringInSlice(
		reposForBuild,
		"FEDORA_MAJOR_VERSION",
		ft.ReleaseVersion,
	)

	// Build Rust security tools in a separate container
	cargoInstallArgs := append([]string{"cargo", "install"}, cargoSecurityTools...)
	rustBuilder := dag.Container().
		From("rust:latest").
		WithMountedCache("/root/.cargo/registry", dag.CacheVolume("cargo-registry")).
		WithExec(cargoInstallArgs)

	ctr := fedora.
		WithPackagesInstalled(packages).
		WithPackageGroupsInstalled(packageGroups).
		WithReposFromUrls(finalReposForBuild, false). // false => delete repo file in final image
		WithPackagesSwapped("mesa-va-drivers", "mesa-va-drivers-freeworld").
		WithPackagesSwapped("mesa-vdpau-drivers", "mesa-vdpau-drivers-freeworld").
		Container(). // ✨ type becomes dagger.Container here!
		WithFile(
			"/usr/bin/distrobox-host-exec",
			dbheFile,
		).
		WithFile("/usr/bin/host-spawn", hostSpawn,
			dagger.ContainerWithFileOpts{Permissions: 0755, Owner: "root"},
		)

	// Copy Rust security tool binaries from builder
	for _, tool := range cargoSecurityTools {
		ctr = ctr.WithFile(
			fmt.Sprintf("/usr/local/bin/%s", tool),
			rustBuilder.File(fmt.Sprintf("/usr/local/cargo/bin/%s", tool)),
			dagger.ContainerWithFileOpts{Permissions: 0755},
		)
	}

	// Install pipx security tools globally
	for _, tool := range pipxSecurityTools {
		ctr = ctr.WithExec([]string{"pipx", "install", "--global", tool})
	}

	// Install npm security tools globally
	for _, tool := range npmSecurityTools {
		ctr = ctr.WithExec([]string{"npm", "install", "-g", tool})
	}

	// Copy scripts and config if source is provided
	if ft.Source != nil {
		scriptsDir := ft.Source.Directory("scripts")
		ctr = ctr.
			WithFile("/usr/local/bin/vet-tool.sh", scriptsDir.File("vet-tool.sh"),
				dagger.ContainerWithFileOpts{Permissions: 0755}).
			WithFile("/usr/local/bin/vet-deps.sh", scriptsDir.File("vet-deps.sh"),
				dagger.ContainerWithFileOpts{Permissions: 0755}).
			WithDirectory("/etc/security-tools", dag.Directory().
				WithFile("mise-security-tools.toml", scriptsDir.File("mise-security-tools.toml")))
	}

	ctr = ctr.WithExec([]string{"dnf", "clean", "all"})

	return ctr, nil
}
