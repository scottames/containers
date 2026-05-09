#!/usr/bin/env bash

# Zed - A high-performance, multiplayer code editor
# https://zed.dev
#
# Zed does not provide an official Fedora repository. Install the official
# tarball directly into /usr/share to keep the payload in the immutable image.

set -ouex pipefail

ZED_CHANNEL="${ZED_CHANNEL:-stable}"

case "$(uname -m)" in
  x86_64)
    ZED_ARCH="x86_64"
    ;;
  aarch64 | arm64)
    ZED_ARCH="aarch64"
    ;;
  *)
    echo "Unsupported Zed architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

echo "=> Installing Zed ${ZED_CHANNEL} (${ZED_ARCH})"

DOWNLOAD_URL="https://cloud.zed.dev/releases/${ZED_CHANNEL}/latest/download?asset=zed&arch=${ZED_ARCH}&os=linux&source=containers"
TMP_DIR="$(mktemp -d)"
ARCHIVE="${TMP_DIR}/zed-linux-${ZED_ARCH}.tar.gz"
trap 'rm -rf "${TMP_DIR}"' EXIT

curl \
  --fail \
  --show-error \
  --location \
  --retry 5 \
  --retry-all-errors \
  --connect-timeout 15 \
  --max-time 300 \
  "${DOWNLOAD_URL}" \
  -o "${ARCHIVE}"

tar -tzf "${ARCHIVE}" | while IFS= read -r path; do
  case "${path}" in
    zed.app | zed.app/*) ;;
    *)
      echo "Unexpected Zed archive path: ${path}" >&2
      exit 1
      ;;
  esac
done

tar \
  --extract \
  --gzip \
  --file "${ARCHIVE}" \
  --directory "${TMP_DIR}" \
  --no-same-owner \
  --no-same-permissions \
  --delay-directory-restore

rm -rf /usr/share/zed.app
mv "${TMP_DIR}/zed.app" /usr/share/zed.app

if [ ! -x /usr/share/zed.app/bin/zed ]; then
  echo "Zed executable was not found after install" >&2
  exit 1
fi

ln -sf /usr/share/zed.app/bin/zed /usr/bin/zed

install -Dm0644 \
  /usr/share/zed.app/share/applications/dev.zed.Zed.desktop \
  /usr/share/applications/dev.zed.Zed.desktop
install -Dm0644 \
  /usr/share/zed.app/share/icons/hicolor/512x512/apps/zed.png \
  /usr/share/icons/hicolor/512x512/apps/zed.png

if ! grep -q '^Exec=zed' /usr/share/applications/dev.zed.Zed.desktop; then
  echo "Zed desktop file does not use the installed zed command" >&2
  exit 1
fi

if ! grep -q '^Icon=zed' /usr/share/applications/dev.zed.Zed.desktop; then
  echo "Zed desktop file does not use the installed zed icon" >&2
  exit 1
fi

echo "=> Zed ${ZED_CHANNEL} installed successfully"
