package cmd

import (
	"fmt"
	"os"

	"github.com/jedi4ever/nddt/provider/docker"
)

// PrintHelp displays usage information
func PrintHelp(version string) {
	PrintHelpWithFlags(version, "", "")
}

// PrintHelpWithFlags displays usage information with extension-specific flags
func PrintHelpWithFlags(version, imageName, command string) {
	fmt.Printf(`nddt - Run AI coding agents in containerized environments

Version: %s

Usage: nddt [options] [prompt]

Commands:
  shell                              Open bash shell in environment
  containers build [--build-arg ...] Build the container image
  containers [list|stop|rm|clean]    Manage persistent environments
  firewall [list|add|remove|reset]   Manage network firewall domains
  --nddt-update                      Check for and install updates
  --nddt-rebuild                     Rebuild the environment (Docker only)
  --nddt-version                     Show nddt version
  --nddt-list-extensions             List available extensions
  --nddt-help                        Show this help

`, version)

	// Try to get extension-specific flags
	if imageName != "" && command != "" {
		printExtensionFlags(imageName, command)
	} else {
		// Fallback to generic options
		fmt.Println(`Options:
  All options are passed to the agent. Generic flags transformed by extensions:
  --yolo                      Bypass permission checks (transformed by extension's args.sh)`)
	}

	fmt.Print(`
Environment Variables:
  NDDT_PROVIDER            Provider type: docker or daytona (default: docker)
  NDDT_NODE_VERSION        Node.js version (default: 22)
  NDDT_GO_VERSION          Go version (default: latest)
  NDDT_UV_VERSION          UV Python package manager version (default: latest)
  NDDT_ENV_VARS            Comma-separated env vars to pass (default: ANTHROPIC_API_KEY,GH_TOKEN)
  NDDT_GITHUB_DETECT       Auto-detect GitHub token from gh CLI (default: false)
  NDDT_PORTS               Comma-separated container ports to expose
  NDDT_PORT_RANGE_START    Starting port for auto allocation (default: 30000)
  NDDT_SSH_FORWARD         SSH forwarding mode: agent, keys, or empty
  NDDT_GPG_FORWARD         Enable GPG forwarding (true/false)
  NDDT_DIND_MODE           Docker-in-Docker mode: host, isolated (default: none)
  NDDT_ENV_FILE            Path to .env file (default: .env)
  NDDT_LOG                 Enable command logging (default: false)
  NDDT_LOG_FILE            Log file path
  NDDT_PERSISTENT          Enable persistent container mode (true/false)
  NDDT_WORKDIR             Override working directory (default: current directory)
  NDDT_WORKDIR_AUTOMOUNT   Auto-mount working directory to /workspace (default: true)
  NDDT_FIREWALL            Enable network firewall (default: false, requires --cap-add=NET_ADMIN)
  NDDT_FIREWALL_MODE       Firewall mode: strict, permissive, off (default: strict)
  NDDT_MODE                Execution mode: container or shell (default: container)
  NDDT_EXTENSIONS          Extensions to install at build time (e.g., claude,codex,gemini)
  NDDT_COMMAND             Command to run instead of claude (e.g., codex, gemini)

Per-Extension Configuration:
  NDDT_<EXT>_VERSION       Version for extension (e.g., NDDT_CLAUDE_VERSION=1.0.5)
                              Default versions defined in each extension's config.yaml
  NDDT_<EXT>_AUTOMOUNT     Auto-mount extension config (e.g., NDDT_CLAUDE_AUTOMOUNT=false)

Build Command:
  nddt containers build [--build-arg KEY=VALUE]...
                              Build the container image with optional build args
                              Example: nddt containers build --build-arg NDDT_EXTENSIONS=gastown

Examples:
  nddt --nddt-help
  nddt "Fix the bug in app.js"
  nddt --yolo "Refactor this entire codebase"
  nddt --help              # Shows agent's help
  nddt shell
  NDDT_COMMAND=codex nddt   # Run Codex instead of Claude
  NDDT_COMMAND=gemini nddt  # Run Gemini instead of Claude
`)
}

// printExtensionFlags queries and prints flags for the active extension
func printExtensionFlags(imageName, command string) {
	// Create a minimal docker provider to query extension flags
	p := &docker.DockerProvider{}
	flags := p.GetExtensionFlags(imageName, command)

	if len(flags) > 0 {
		fmt.Printf("Options (%s):\n", command)
		for _, flag := range flags {
			fmt.Printf("  %-25s %s\n", flag.Flag, flag.Description)
		}
		fmt.Println()
	} else {
		fmt.Println(`Options:
  All options are passed to the agent. Generic flags transformed by extensions:
  --yolo                      Bypass permission checks (transformed by extension's args.sh)`)
	}
}

// GetActiveCommand returns the active command from env or default
func GetActiveCommand() string {
	if cmd := os.Getenv("NDDT_COMMAND"); cmd != "" {
		return cmd
	}
	return "claude"
}
