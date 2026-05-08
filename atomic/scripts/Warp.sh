#!/usr/bin/env bash

set -ouex pipefail

echo "=> Installing Warp Terminal"

# On libostree systems, /opt is a symlink to /var/opt. RPM payloads
# installed there need to be moved into immutable image-owned storage.
mkdir -p /var/opt

rpm --import https://releases.warp.dev/linux/keys/warp.asc

cat >/etc/yum.repos.d/warpdotdev.repo <<'EOF'
[warpdotdev]
name=warpdotdev
baseurl=https://releases.warp.dev/linux/rpm/stable
enabled=1
gpgcheck=1
gpgkey=https://releases.warp.dev/linux/keys/warp.asc
EOF

if command -v rpm-ostree >/dev/null 2>&1; then
  rpm-ostree install warp-terminal
elif command -v dnf5 >/dev/null 2>&1; then
  dnf5 install -y warp-terminal
else
  dnf install -y warp-terminal
fi

# Clean up the yum repo as updates are baked in based on this script.
rm -f /etc/yum.repos.d/warpdotdev.repo

source_count=0
if [ -d /var/opt/warpdotdev ]; then
  source_count=$((source_count + 1))
fi
if [ -d /usr/lib/opt/warpdotdev ]; then
  source_count=$((source_count + 1))
fi
if [ ! -L /opt ] && [ -d /opt/warpdotdev ]; then
  source_count=$((source_count + 1))
fi

if [ "${source_count}" -gt 1 ]; then
  echo "Unexpected Warp layout: multiple warpdotdev payload directories found" >&2
  rpm -ql warp-terminal | grep -E 'warpdotdev|warp-terminal|/opt|/usr/lib/opt' || true
  exit 1
fi

if [ -d /var/opt/warpdotdev ]; then
  rm -rf /usr/lib/warpdotdev
  mv /var/opt/warpdotdev /usr/lib/warpdotdev
elif [ -d /usr/lib/opt/warpdotdev ]; then
  rm -rf /usr/lib/warpdotdev
  mv /usr/lib/opt/warpdotdev /usr/lib/warpdotdev
elif [ ! -L /opt ] && [ -d /opt/warpdotdev ]; then
  rm -rf /usr/lib/warpdotdev
  mv /opt/warpdotdev /usr/lib/warpdotdev
fi

if [ -d /usr/lib/warpdotdev ]; then
  if [ ! -x /usr/lib/warpdotdev/warp-terminal/warp ]; then
    echo "Warp executable was not found after payload relocation" >&2
    rpm -ql warp-terminal | grep -E 'warpdotdev|warp-terminal|/opt|/usr/lib/opt' || true
    exit 1
  fi

  mkdir -p /usr/lib/tmpfiles.d

  cat >/usr/lib/tmpfiles.d/warpdotdev.conf <<'EOF'
L  /opt/warpdotdev  -  -  -  -  /usr/lib/warpdotdev
EOF

  rm -f /usr/bin/warp-terminal
  ln -s /opt/warpdotdev/warp-terminal/warp /usr/bin/warp-terminal
else
  echo "Warp payload directory was not found after package installation" >&2
  rpm -ql warp-terminal | grep -E 'warpdotdev|warp-terminal|/opt|/usr/lib/opt' || true
  exit 1
fi

rpm -ql warp-terminal | grep -E 'warpdotdev|warp-terminal|/opt|/usr/lib/opt' || true
