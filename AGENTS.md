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
