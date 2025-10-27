package main

func (a *Atomic) getPackageListFrom(packageMap map[string]map[string][]string) []string {
	packages := []string{}
	suffix := Main
	if a.Suffix != nil {
		suffix = *a.Suffix
	}

	for _, opts := range sliceStringProduct([]string{All, a.Variant, suffix}) {
		p, ok := packageMap[opts[0]][opts[1]]
		if ok {
			packages = append(packages, p...)
		}
	}

	return packages
}

func sliceStringProduct(ss []string) [][]string {
	r := [][]string{}
	for _, one := range ss {
		for _, two := range ss {
			r = append(r, []string{one, two})
		}
	}

	return r
}

const (
	All        = "all"
	Main       = "main"
	Niri       = "niri"
	Nvidia     = "nvidia"
	Silverblue = "silverblue"
)

var (
	reposForBuild = []string{ // will not be kept in final image
		"https://pkgs.tailscale.com/stable/fedora/tailscale.repo",
		"https://copr.fedorainfracloud.org/coprs/yalter/niri/repo/fedora-FEDORA_MAJOR_VERSION/yalter-niri-fedora-FEDORA_MAJOR_VERSION.repo",
		"https://copr.fedorainfracloud.org/coprs/scottames/ghostty/repo/fedora-FEDORA_MAJOR_VERSION/scottames-ghostty-fedora-FEDORA_MAJOR_VERSION.repo",
		"https://copr.fedorainfracloud.org/coprs/scottames/hypr/repo/fedora-FEDORA_MAJOR_VERSION/scottames-hypr-fedora-FEDORA_MAJOR_VERSION.repo",
		"https://copr.fedorainfracloud.org/coprs/scottames/swww/repo/fedora-FEDORA_MAJOR_VERSION/scottames-swww-fedora-FEDORA_MAJOR_VERSION.repo",
		"https://copr.fedorainfracloud.org/coprs/tofik/nwg-shell/repo/fedora-FEDORA_MAJOR_VERSION/tofik-nwg-shell-fedora-FEDORA_MAJOR_VERSION.repo",
	}
	// for layering, primarily because these packages do not play well with opt
	reposForImage = []string{
		"https://repo.vivaldi.com/stable/vivaldi-fedora.repo",
		"https://copr.fedorainfracloud.org/coprs/scottames/zen-browser/repo/fedora-FEDORA_MAJOR_VERSION/scottames-zen-browser-fedora-FEDORA_MAJOR_VERSION.repo",
	}
	packagesRemoved = map[string]map[string][]string{
		Silverblue: {
			Nvidia: {
				// https://github.com/ublue-os/hwe/blob/main/nvidia-install.sh#L29C19-L29C56
				//  not using any applicable hardware. Extension has root-only
				//  permission on metadata, causing errors with gext interaction
				"gnome-shell-extension-supergfxctl-gex",
				"supergfxctl",
			},
		},
		All: {
			All: {
				"opensc", // breaks Yubikey
			},
		},
	}
	packagesInstalled = map[string]map[string][]string{
		Niri: {
			All: {
				"gnome-keyring",
				"grim",
				"hypridle",
				"hyprlock",
				"hyprpaper",
				"hyprpicker",
				"mako",
				"niri",
				"nwg-look",
				"pavucontrol",
				"mate-polkit",
				"rofi-wayland",
				"rofimoji",
				"slurp",
				"swaybg",
				"swayidle",
				"swaylock",
				"swww",
				"waybar",
				"wlogout",
				"wtype",
				"xdg-desktop-portal-gnome",
				"xdg-desktop-portal-gtk",
			},
		},
		All: {
			All: {
				// Installed via script
				// "1password",
				// "1password-cli",

				"adobe-source-code-pro-fonts",
				"arm-image-installer",
				"cascadia-code-fonts",
				"dbus-x11",
				"firewall-config",
				"fish",
				"ghostty",
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
				"light",
				"lm_sensors", // required by freon gnome-ext
				"mozilla-fira-fonts-common",
				"mozilla-fira-mono-fonts",
				"mscore-fonts-all",
				"netcat",
				"NetworkManager-tui",
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
			},
		},
	}
)
