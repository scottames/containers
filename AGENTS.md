# AGENTS.md

This file provides guidance to coding agents (Claude, Gemini, Opencode, Codex,
etc.) when working with code in this repository.

## Project Overview

Personal container image builder for custom

- Fedora Atomic (Silverblue, Niri) (`atomic/`)
- Toolbox/Distrobox (`toolbox/`)

Uses Dagger (Go SDK) for containerized build automation.

## Build System

**Dagger** is the primary build tool - all container building logic is in Go
modules:

- `atomic/` - Fedora Atomic desktop images (Silverblue, Niri variants)
- `toolbox/fedora/` - Fedora toolbox/distrobox container

**Just** is the task runner. All commands below are run from repo root.

## Common Commands

```bash
# List all recipes
just

# Setup Go workspace (required after clone)
just go-work

# Regenerate Dagger modules after dependency changes
just develop

# Build/evaluate containers locally (no publish)
just atomic-container
just fedora-toolbox-container

# Publish and sign (requires GITHUB_USERNAME, GITHUB_TOKEN, COSIGN_PASSWORD, COSIGN_PRIVATE_KEY)
just atomic-publish-and-sign
just fedora-toolbox-publish-and-sign

# Update scottames/daggerverse dependencies across all modules
just update-scottames-daggerverse <version>

# Initialize new Dagger module
just init <module>
```

## Architecture

Each Dagger module (`atomic/`, `toolbox/fedora/`) contains:

- `dagger.json` - Module configuration and dependencies
- `main.go` - Module initialization and Dagger function exports
- `*.go` - Implementation (packages, publishing, signing, etc.)
- `justfile` - Module-specific recipes (imported by root justfile)

Key patterns:

- Modules depend on `scottames/daggerverse` for `cosign` and `fedora`
  functionality
- Images are signed with Cosign via GitHub secrets
- Fedora version is pinned in root justfile (`tagFedoraLatestVersion`) and
  tracked by Renovate

## Tool Versions

Managed via Aqua (`.aqua/aqua.yaml`):

- `aqua install` to install pinned versions
- dagger, just, gobrew, ghcp

## CI/CD

GitHub Actions workflows in `.github/workflows/`:

- `atomic.yaml` - Daily builds of Atomic images
- `toolbox.yaml` / `reusable-toolbox.yaml` - Toolbox builds
- `dagger-update.yaml` - Auto-updates Dagger modules on renovate branches

## Fedora Atomic / ostree Filesystem Constraints

When adding install scripts to `atomic/scripts/`, be aware of ostree filesystem
behavior:

### Safe Locations (baked into image)

- `/usr/share/` - Standard location for application data
- `/usr/lib/` - Libraries and application binaries
- `/usr/bin/` - Executables (or symlinks to them)
- `/etc/` - Configuration files

### Problematic Locations

- `/opt` - Symlink to `/var/opt` on ostree systems. `/var` is a separate mutable
  filesystem that only exists on the live system, NOT during image build. Files
  written to `/opt` or `/var` during build will NOT persist in the final image.
- `/var/*` - Same issue as `/opt`

### Workaround Pattern (see `1Password.sh`)

If an RPM installs to `/opt`, you must:

1. Install the RPM (files land in `/var/opt/`)
2. Move files to `/usr/lib/<app>`
3. Create runtime symlinks via `/usr/lib/tmpfiles.d/<app>.conf`:
   ```
   L  /opt/<app>  -  -  -  -  /usr/lib/<app>
   ```

### Best Practice

When writing install scripts for tarballs/binaries, install directly to
`/usr/share/<app>` to avoid the `/opt` dance entirely.

References:

- [Fedora Silverblue filesystem structure](https://insujang.github.io/2020-07-15/fedora-silverblue/)
- [Universal Blue 1Password installer](https://github.com/ublue-os/bling/blob/main/modules/bling/installers/1password.sh)
