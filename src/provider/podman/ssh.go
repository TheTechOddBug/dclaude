package podman

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedi4ever/addt/config/security"
	"github.com/jedi4ever/addt/util"
)

// HandleSSHForwarding configures SSH forwarding based on config.
// Modes:
//   - "proxy": Forward filtered SSH agent (only allowed keys)
//   - "agent" or "true": Forward SSH agent socket
//   - "keys": Mount ~/.ssh directory read-only
//   - "" or other: No SSH forwarding
//
// If allowedKeys is set, proxy mode is automatically enabled for agent forwarding
func (p *PodmanProvider) HandleSSHForwarding(sshForward, homeDir, username string, allowedKeys []string) []string {
	var args []string

	// If allowed keys are specified, use proxy mode regardless of sshForward setting
	if len(allowedKeys) > 0 && (sshForward == "agent" || sshForward == "true" || sshForward == "proxy") {
		return p.handleSSHProxyForwarding(homeDir, username, allowedKeys)
	}

	if sshForward == "proxy" {
		// Proxy mode without filters - just forward all keys through proxy
		return p.handleSSHProxyForwarding(homeDir, username, nil)
	} else if sshForward == "agent" || sshForward == "true" {
		args = p.handleSSHAgentForwarding(homeDir, username)
	} else if sshForward == "keys" {
		args = p.handleSSHKeysForwarding(homeDir, username)
	}

	return args
}

// handleSSHProxyForwarding creates a filtered SSH agent proxy
func (p *PodmanProvider) handleSSHProxyForwarding(homeDir, username string, allowedKeys []string) []string {
	var args []string

	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		fmt.Println("Warning: SSH_AUTH_SOCK not set, cannot create SSH proxy")
		return args
	}

	// Create the proxy agent
	proxy, err := security.NewSSHProxyAgent(sshAuthSock, allowedKeys)
	if err != nil {
		fmt.Printf("Warning: failed to create SSH proxy: %v\n", err)
		return args
	}

	// Start the proxy
	if err := proxy.Start(); err != nil {
		fmt.Printf("Warning: failed to start SSH proxy: %v\n", err)
		return args
	}

	// Store proxy for cleanup
	p.sshProxy = proxy

	// Mount the proxy socket
	proxySocket := proxy.SocketPath()
	args = append(args, "-v", fmt.Sprintf("%s:/ssh-agent", proxySocket))
	args = append(args, "-e", "SSH_AUTH_SOCK=/ssh-agent")

	// Mount safe SSH files only (config, known_hosts, public keys)
	args = append(args, p.mountSafeSSHFiles(homeDir, username)...)

	if len(allowedKeys) > 0 {
		fmt.Printf("SSH proxy active: only keys matching %v are accessible\n", allowedKeys)
	} else {
		fmt.Println("SSH proxy active: all keys accessible")
	}

	return args
}

// handleSSHAgentForwarding forwards the SSH agent socket into the container
func (p *PodmanProvider) handleSSHAgentForwarding(homeDir, username string) []string {
	var args []string

	sshAuthSock := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSock == "" {
		return args
	}

	// Check if socket exists and is accessible
	if _, err := os.Stat(sshAuthSock); err != nil {
		return args
	}

	// Check for macOS launchd sockets (may have issues with container runtimes)
	if strings.Contains(sshAuthSock, "com.apple.launchd") || strings.Contains(sshAuthSock, "/var/folders/") {
		fmt.Println("Warning: SSH agent forwarding may not work on macOS (use ADDT_SSH_FORWARD=keys)")
		return args
	}

	// Mount the SSH agent socket
	args = append(args, "-v", fmt.Sprintf("%s:/ssh-agent", sshAuthSock))
	args = append(args, "-e", "SSH_AUTH_SOCK=/ssh-agent")

	// Mount safe SSH files only (config, known_hosts, public keys)
	args = append(args, p.mountSafeSSHFiles(homeDir, username)...)

	return args
}

// handleSSHKeysForwarding mounts the entire ~/.ssh directory read-only
func (p *PodmanProvider) handleSSHKeysForwarding(homeDir, username string) []string {
	var args []string

	sshDir := filepath.Join(homeDir, ".ssh")
	if _, err := os.Stat(sshDir); err == nil {
		args = append(args, "-v", fmt.Sprintf("%s:/home/%s/.ssh:ro", sshDir, username))
	}

	return args
}

// mountSafeSSHFiles creates a temp directory with only safe SSH files
// (config, known_hosts, public keys) and returns mount arguments
func (p *PodmanProvider) mountSafeSSHFiles(homeDir, username string) []string {
	var args []string

	sshDir := filepath.Join(homeDir, ".ssh")
	if _, err := os.Stat(sshDir); err != nil {
		return args
	}

	tmpDir, err := os.MkdirTemp("", "ssh-safe-*")
	if err != nil {
		return args
	}

	p.tempDirs = append(p.tempDirs, tmpDir)

	// Copy safe files only
	util.SafeCopyFile(filepath.Join(sshDir, "config"), filepath.Join(tmpDir, "config"))
	util.SafeCopyFile(filepath.Join(sshDir, "known_hosts"), filepath.Join(tmpDir, "known_hosts"))

	// Copy public keys
	files, _ := filepath.Glob(filepath.Join(sshDir, "*.pub"))
	for _, f := range files {
		util.SafeCopyFile(f, filepath.Join(tmpDir, filepath.Base(f)))
	}

	args = append(args, "-v", fmt.Sprintf("%s:/home/%s/.ssh:ro", tmpDir, username))

	return args
}
