package main

import (
	"context"
	"dagger/sandbox/internal/dagger"
	"fmt"
	"strings"
)

// renovate: datasource=github-releases depName=dandavison/delta
const gitDeltaVersion = "0.18.2"

func (s *Sandbox) Container(ctx context.Context,
	// Timezone for the container
	// +optional
	// +default="UTC"
	tz string,
) *dagger.Container {
	baseImage := s.Image + ":" + s.Tag
	if s.Org != nil {
		baseImage = *s.Org + "/" + baseImage
	}
	if s.Registry != "docker.io" {
		baseImage = s.Registry + "/" + baseImage
	}

	gitDeltaFile := fmt.Sprintf(
		"git-delta_%s_${ARCH}.deb",
		gitDeltaVersion,
	)

	nodeUserPath := []string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
		"/usr/local/share/npm-global/bin",
	}

	return dag.Container().
		From(baseImage).
		WithEnvVariable("TZ", tz).
		// Install basic development tools and iptables/ipset
		WithExec([]string{"apt", "update"}).
		WithExec([]string{
			"apt", "install", "-y",
			"less",
			"git",
			"procps",
			"sudo",
			"fzf",
			"fish",
			"man-db",
			"unzip",
			"gnupg2",
			"gh",
			"iptables",
			"ipset",
			"iproute2",
			"dnsutils",
			"aggregate",
			"jq",
		}).
		// Ensure default node user has access to /usr/local/share
		WithExec([]string{"mkdir", "-p", "/usr/local/share/npm-global"}).
		WithExec([]string{"chown", "-R", "node:node", "/usr/local/share"}).
		// Create workspace and config directories and set permissions
		WithExec([]string{
			"mkdir", "-p", "/workspace", "/home/node/.claude",
		}).
		WithExec([]string{
			"chown", "-R", "node:node", "/workspace", "/home/node/.claude",
		}).
		WithWorkdir("/workspace").
		// Install git-delta
		WithExec([]string{
			"sh", "-c",
			fmt.Sprintf(`ARCH=$(dpkg --print-architecture) &&
			 wget "https://github.com/dandavison/delta/releases/download/%s/%s" &&
			 dpkg -i "%s" &&
			 rm "%s"`,
				gitDeltaVersion, gitDeltaFile, gitDeltaFile, gitDeltaFile,
			),
		}).
		// Switch to node user
		WithUser("node").
		// Set up npm config for non-root user
		WithEnvVariable(
			"NPM_CONFIG_PREFIX",
			"/usr/local/share/npm-global",
		).
		WithEnvVariable(
			"PATH",
			strings.Join(nodeUserPath, ":"),
		).

		// Set fish as default shell
		WithEnvVariable("SHELL", "/usr/bin/fish").
		// Install Claude CLI using full path
		WithExec([]string{
			"/usr/local/bin/npm", "install", "-g", "@anthropic-ai/claude-code",
		}).
		// Copy and set up firewall script
		WithFile("/usr/local/bin/init-firewall.sh",
			dag.CurrentModule().Source().File("init-firewall.sh"),
			dagger.ContainerWithFileOpts{Permissions: 0755},
		).
		WithUser("root").
		WithExec([]string{
			"sh", "-c",
			`echo "node ALL=(root) NOPASSWD: /usr/local/bin/init-firewall.sh" > /etc/sudoers.d/node-firewall &&
			 chmod 0440 /etc/sudoers.d/node-firewall`,
		}).
		WithUser("node")
}
