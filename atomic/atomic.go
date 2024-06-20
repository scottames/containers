package main

import (
	"context"
	"fmt"
)

const (
	// renovate: datasource=docker depName=quay.io/fedora-ostree-desktops/silverblue
	latestFedoraVersion = "40"

	description = "scottames' custom Fedora Silverblue native container image powered by Universal Blue."
)

var (
	labels = map[string]string{
		"io.artifacthub.package.readme-url": "https://raw.githubusercontent.com/scottames/containers/main/atomic/README.md",
		"org.opencontainers.image.url":      "https://github.com/scottames/containers/tree/main/atomic",
	}
	reposForBuild = []string{ // will not be kept in final image
		// TODO: match fedora version (when available, 40 returns 404)
		"https://pkgs.tailscale.com/stable/fedora/39/tailscale.repo",
	}
	reposForImage = []string{
		"https://repo.vivaldi.com/stable/vivaldi-fedora.repo", // Layering for now...
	}
	packagesInstalled = []string{
		// Installed via script
		// "1password",
		// "1password-cli",

		"adobe-source-code-pro-fonts",
		"arm-image-installer",
		"cascadia-code-fonts",
		"dbus-x11",
		"firewall-config",
		"fish",
		"google-droid-sans-fonts",
		"google-droid-sans-mono-fonts",
		"google-go-mono-fonts",
		"google-noto-color-emoji-fonts",
		"google-noto-emoji-fonts",
		"google-noto-fonts-common",
		"google-roboto-fonts",
		"ibm-plex-mono-fonts",
		"iotop",
		"jetbrains-mono-fonts-all",
		"langpacks-en",
		"libadwaita",
		"lm_sensors", // required by freon gnome-ext
		"mozilla-fira-fonts-common",
		"mozilla-fira-mono-fonts",
		"mscore-fonts-all",
		"netcat",
		"open-sans-fonts",
		"pam-u2f",
		"pamu2fcfg",
		"pipx",
		"podman-compose",
		"podman-tui",
		"powerline-fonts",
		"powertop",
		"pulseaudio-utils",
		"tailscale",
		"udica",
		"wl-clipboard",
		"xclip",
		"yubico-piv-tool-devel",
		"yubikey-manager",
		"yubikey-manager-qt",

		// Qemu / Virt-manager
		"edk2-ovmf",
		"genisoimage",
		"libvirt",
		"qemu",
		"qemu-char-spice",
		"qemu-device-display-virtio-gpu",
		"qemu-device-display-virtio-vga",
		"qemu-device-usb-redirect",
		"qemu-img",
		"qemu-system-x86-core",
		"qemu-user-binfmt",
		"qemu-user-static",
		"virt-manager",
		"virt-viewer",
		// Required for ZSA voyager
		"gtk3",
		"libusb",
		"webkit2gtk3",
		"webkit2gtk4.0",
		// Required for https://github.com/oae/gnome-shell-pano
		"libgda",
		"libgda-sqlite",
	}

	packagesRemoved = []string{
		"opensc", // breaks Yubikey
	}
	scriptsPostPackageInstall = []string{
		"1Password.sh",
		"wezterm.sh",
	}
)

// fedoraAtomic defines the custom Fedora Atomic container image
//
// the container and publish functions both refer to this as their source
func (a *Atomic) fedoraAtomic(ctx context.Context) (*Fedora, error) {
	scriptsPost := []*File{}
	for _, script := range scriptsPostPackageInstall {
		scriptsPost = append(scriptsPost, a.Source.File(
			fmt.Sprintf(
				"atomic/scripts/%s",
				script,
			),
		))
	}

	opts := FedoraOpts{
		Registry: a.Registry,
		Org:      a.Org,
		Tag:      a.Tag,
		Variant:  a.Variant,
	}
	if a.Suffix != nil {
		opts.Suffix = *a.Suffix
	}

	fedora := dag.Fedora(opts)

	version, err := fedora.ReleaseVersion(ctx)
	if err != nil {
		version, err = fedora.Date(ctx)
		if err != nil {
			return nil, err
		}
	}

	a.ReleaseVersion = version

	a.Tags, err = fedora.DefaultTags(ctx,
		FedoraDefaultTagsOpts{Latest: latestFedoraVersion == version},
	)
	if err != nil {
		return nil, err
	}

	fedora, err = a.fedoraWithLabelsFromCLI(fedora)
	if err != nil {
		return nil, err
	}

	if !a.SkipDefaultLabels {
		fedora, err = a.fedoraWithDefaultLabels(ctx, fedora)
		if err != nil {
			return nil, err
		}

		fedora = fedora.WithDescription(description)
	}

	// Fedora is derived from the installed dagger module dependency
	return fedora.
			WithDescription(description).
			WithDirectory("/usr", a.Source.Directory("atomic/files/usr")).
			WithReposFromUrls(reposForImage, true).
			// true => keep repo in final image
			WithReposFromUrls(reposForBuild, false).
			// false => delete repo file in final image
			WithPackagesInstalled(packagesInstalled).
			WithPackagesRemoved(packagesRemoved).
			WithExecScripts(scriptsPost, false).          // false => post package install
			WithExec([]string{"update-ca-trust"}, false), // false => post package install
		nil
}
