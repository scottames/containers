#!/usr/bin/env bash

# Obsidian - A powerful knowledge base on top of local Markdown files
# https://obsidian.md
#
# Downloads and installs the official tarball since Obsidian's license
# does not permit redistribution via package repositories like COPR.

set -ouex pipefail

# renovate: datasource=github-releases depName=obsidianmd/obsidian-releases
OBSIDIAN_VERSION="v1.11.5"

echo "=> Installing Obsidian ${OBSIDIAN_VERSION}"

# Download and extract
DOWNLOAD_URL="https://github.com/obsidianmd/obsidian-releases/releases/download/${OBSIDIAN_VERSION}/obsidian-${OBSIDIAN_VERSION#v}.tar.gz"
curl -fsSL "${DOWNLOAD_URL}" | tar -xz -C /tmp

# Install to /usr/share/obsidian (safe for ostree - baked into image)
mkdir -p /usr/share/obsidian
cp -a /tmp/obsidian-"${OBSIDIAN_VERSION#v}"/* /usr/share/obsidian/
rm -rf /tmp/obsidian-"${OBSIDIAN_VERSION#v}"

# Create symlink for binary
ln -s /usr/share/obsidian/obsidian /usr/bin/obsidian

# Create .desktop file (Obsidian no longer ships one)
cat > /usr/share/applications/md.obsidian.Obsidian.desktop <<'DESKTOP'
[Desktop Entry]
Name=Obsidian
Comment=A powerful knowledge base on top of local Markdown files
Exec=obsidian %U
Terminal=false
Type=Application
Icon=obsidian
StartupWMClass=obsidian
Categories=Office;Utility;TextEditor;
MimeType=text/markdown;x-scheme-handler/obsidian;
DESKTOP

# Install icon
mkdir -p /usr/share/icons/hicolor/512x512/apps
cp /usr/share/obsidian/resources/icon.png /usr/share/icons/hicolor/512x512/apps/obsidian.png

echo "=> Obsidian ${OBSIDIAN_VERSION} installed successfully"
