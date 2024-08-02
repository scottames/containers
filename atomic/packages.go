package main

var (
	reposForBuild = []string{ // will not be kept in final image
		// TODO: match fedora version (when available, 40 returns 404)
		"https://pkgs.tailscale.com/stable/fedora/39/tailscale.repo",
	}
	reposForImage = []string{
		"https://repo.vivaldi.com/stable/vivaldi-fedora.repo", // Layering for now...
	}
	packagesRemoved = []string{
		"opensc", // breaks Yubikey
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
		"kitty",
		"kitty-doc",
		"kitty-shell-integration",
		"kitty-terminfo",
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
)
