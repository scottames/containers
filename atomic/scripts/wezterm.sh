#!/usr/bin/env bash

set -ex

# renovate: depName=wez/wezterm datasource=github-releases
WEZTERM_RELEASE="20240203-110809-5046fc22"
WEZTERM_RELEASE_UNDERSCORE="${WEZTERM_RELEASE//-/_}"
# TODO: track fedora release (and arch?) [fedora 40 not available yet...]
WEZTERM_FILENAME="wezterm-${WEZTERM_RELEASE_UNDERSCORE}-1.fedora39.x86_64.rpm"

TMP_DIR=$(mktemp -d)
pushd "${TMP_DIR}"

curl -fsSL -O \
  "https://github.com/wez/wezterm/releases/download/${WEZTERM_RELEASE}/${WEZTERM_FILENAME}"
curl -fsSL -O \
  "https://github.com/wez/wezterm/releases/download/${WEZTERM_RELEASE}/${WEZTERM_FILENAME}.sha256"

sha256sum -c "${WEZTERM_FILENAME}.sha256"

rpm-ostree install "${WEZTERM_FILENAME}"

popd

rm -rf "${TMP_DIR}"
