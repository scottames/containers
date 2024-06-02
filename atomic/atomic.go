package main

import "fmt"

var (
	description   = "scottames' custom Fedora Silverblue native container image powered by Universal Blue."
	reposForBuild = []string{ // will not be kept in final image
		// TODO: move this
		"https://raw.githubusercontent.com/scottames/ublue/live/config/repos/tailscale.repo",
		"https://repository.mullvad.net/rpm/stable/mullvad.repo",
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
		// "mullvad-vpn",
		"netcat",
		"open-sans-fonts",
		"pam-u2f",
		"pamu2fcfg",
		"pipx",
		"podman-compose",
		"podman-tui",
		"podmansh",
		"powerline-fonts",
		"powertop",
		"pulseaudio-utils",
		"python3-pip", // needed by Yafti
		"tailscale",
		"udica",
		"wl-clipboard",
		"xclip",
		"yubico-piv-tool-devel",
		"yubikey-manager",
		"yubikey-manager-qt",

		// Virt-manager packages from bluefin-dx
		// "edk2-ovmf",
		// "genisoimage",
		// "libvirt",
		// "qemu",
		// "qemu-char-spice",
		// "qemu-device-display-virtio-gpu",
		// "qemu-device-display-virtio-vga",
		// "qemu-device-usb-redirect",
		// "qemu-img",
		// "qemu-system-x86-core",
		// "qemu-user-binfmt",
		// "qemu-user-static",
		// "virt-manager",
		// "virt-viewer",
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
		// TODO: variables in signing script won't work as-is
		// "signing.sh",
		// "vivaldi.sh", // layering for now...
		"wezterm.sh",
	}
)

// fedoraAtomic defines the custom Fedora Atomic container image
//
// the container and publish functions both refer to this as their source
func (a *Atomic) fedoraAtomic() *Fedora {
	scriptsPost := []*File{}
	for _, script := range scriptsPostPackageInstall {
		scriptsPost = append(scriptsPost, a.Source.File(
			fmt.Sprintf("atomic/scripts/%s", script),
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

	// Fedora is derived from the installed dagger module dependency
	return dag.Fedora(opts).
		WithDescription(description).
		WithLabel(FedoraWithLabelOpts{
			Name: "io.artifacthub.package.readme-url",
			// TODO: update
			Value: "https://raw.githubusercontent.com/scottames/ublue/live/README.md",
		}).
		WithDirectory("/usr", a.Source.Directory("atomic/files/usr")).
		WithFile("/usr/share/ublue-os/cosign.pub", a.Source.File("cosign.pub")).
		WithRepos(reposForImage, true).  // true => keep repo in final image
		WithRepos(reposForBuild, false). // false => delete repo file in final image
		WithPackagesInstalled(packagesInstalled).
		WithPackagesRemoved(packagesRemoved).
		WithExecScripts(scriptsPost, false).          // false => post package install
		WithExec([]string{"update-ca-trust"}, false). // false => post package install
		WithExec([]string{"ostree", "container", "commit"}, false)
}
