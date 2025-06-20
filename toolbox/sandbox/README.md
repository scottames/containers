# Sandbox Toolbox Container

A toolbox (distrobox) container intended to be a controlled sandbox for who
knows what! (AI)

- primarily intended for use with [Claude code](https://www.anthropic.com/engineering/claude-code-best-practices)
  and taken from their
  [devcontainer example](https://github.com/anthropics/claude-code/tree/main/.devcontainer)

## Ways to use it

### Dynamic Distrobox with Wrapper Script

Create a wrapper script for project-specific containers:

```bash
#!/bin/bash
# sandbox-claude
PROJECT_PATH="${1:-$(pwd)}"
CONTAINER_NAME="sandbox-$(basename "$PROJECT_PATH")"

distrobox enter "$CONTAINER_NAME" \
  --additional-flags "--volume $PROJECT_PATH:/workspace:Z" \
  -- claude "${@:2}" 2>/dev/null ||
distrobox create --name "$CONTAINER_NAME" --image <registry>/sandbox:latest \
  --unshare-all --no-home --volume "$PROJECT_PATH:/workspace:Z" &&
distrobox enter "$CONTAINER_NAME" -- claude "${@:2}"
```

Usage: `sandbox-claude /path/to/project` or `sandbox-claude` (uses current directory)

### Podman Wrapper

For ephemeral containers without persistence:

```bash
#!/bin/bash
# sandbox-claude
PROJECT_PATH="${1:-$(pwd)}"
podman run -it --rm \
  --volume "$PROJECT_PATH:/workspace:Z" \
  --cap-add NET_ADMIN \
  --name "sandbox-$(basename "$PROJECT_PATH")" \
  <registry>/sandbox:latest \
  bash -c "sudo /usr/local/bin/init-firewall.sh && claude ${*:2}"
```

Usage: `sandbox-claude /path/to/project [claude-args]`

### Persistent Distrobox with Home

For shared configuration across projects:

```bash
distrobox create --name sandbox --image <registry>/sandbox:latest \
  --unshare-all --home ~/.local/share/sandbox-home

# Then navigate to projects or mount dynamically
distrobox enter sandbox --additional-flags "--volume /path/to/project:/workspace:Z"
```
