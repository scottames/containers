#!/usr/bin/env bash

set -ex

# renovate: depName=wez/wezterm datasource=github-releases
WEZTERM_RELEASE="20240203-110809-5046fc22"
WEZTERM_RELEASE_UNDERSCORE="${WEZTERM_RELEASE//-/_}"
FEDORA_RELEASE="$(rpm -E %fedora)"
WEZTERM_FILENAME="wezterm-${WEZTERM_RELEASE_UNDERSCORE}-1.fedora${FEDORA_RELEASE}.x86_64.rpm"

TMP_DIR=$(mktemp -d)
pushd "${TMP_DIR}"

if [[ "${FEDORA_RELEASE}" -ge "40" ]]; then

  WEZTERM_FILENAME="wezterm-nightly-fedora40.rpm"
  curl -fsSL -O \
    "https://github.com/wez/wezterm/releases/download/nightly/${WEZTERM_FILENAME}"

  curl -fsSL -O \
    "https://github.com/wez/wezterm/releases/download/nightly/${WEZTERM_FILENAME}.sha256"

else

  curl -fsSL -O \
    "https://github.com/wez/wezterm/releases/download/${WEZTERM_RELEASE}/${WEZTERM_FILENAME}"

  curl -fsSL -O \
    "https://github.com/wez/wezterm/releases/download/${WEZTERM_RELEASE}/${WEZTERM_FILENAME}.sha256"

fi

sha256sum -c "${WEZTERM_FILENAME}.sha256"

rpm-ostree install "${WEZTERM_FILENAME}"

popd

rm -rf "${TMP_DIR}"
