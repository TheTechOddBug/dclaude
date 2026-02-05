package podman

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jedi4ever/addt/config/security"
)

// tmuxProxy holds the proxy listener and connections
type tmuxProxy struct {
	listener       net.Listener
	upstreamSocket string
	proxyDir       string
	wg             sync.WaitGroup
	running        bool
	mu             sync.Mutex
}

// HandleTmuxForwarding handles tmux socket forwarding to the container.
// When enabled, it detects the active tmux socket and creates a proxy in ~/.addt/sockets/
// so that Podman machine can access it.
func (p *PodmanProvider) HandleTmuxForwarding(enabled bool) []string {
	if !enabled {
		return nil
	}

	// Check if we're inside a tmux session
	tmuxEnv := os.Getenv("TMUX")
	if tmuxEnv == "" {
		// Not in a tmux session, nothing to forward
		return nil
	}

	// TMUX env format: /tmp/tmux-1000/default,12345,0
	// First part is the socket path
	parts := strings.Split(tmuxEnv, ",")
	if len(parts) < 1 {
		return nil
	}

	socketPath := parts[0]
	if socketPath == "" {
		return nil
	}

	// Verify socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil
	}

	// Create a proxy socket in ~/.addt/sockets/ so Podman machine can access it
	proxyDir, proxySocket, err := p.createTmuxProxy(socketPath)
	if err != nil {
		fmt.Printf("Warning: failed to create tmux proxy: %v\n", err)
		return nil
	}

	// Build new TMUX env with proxy socket path
	newTmuxEnv := proxySocket
	if len(parts) > 1 {
		newTmuxEnv = proxySocket + "," + strings.Join(parts[1:], ",")
	}

	var args []string

	// Mount the proxy socket directory
	args = append(args, "-v", fmt.Sprintf("%s:%s", proxyDir, proxyDir))

	// Pass the modified TMUX environment variable
	args = append(args, "-e", fmt.Sprintf("TMUX=%s", newTmuxEnv))

	// Also pass TMUX_PANE if set
	if tmuxPane := os.Getenv("TMUX_PANE"); tmuxPane != "" {
		args = append(args, "-e", fmt.Sprintf("TMUX_PANE=%s", tmuxPane))
	}

	return args
}

// createTmuxProxy creates a Unix socket proxy from ~/.addt/sockets/ to the real tmux socket
func (p *PodmanProvider) createTmuxProxy(upstreamSocket string) (string, string, error) {
	// Create socket directory in ~/.addt/sockets/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get home dir: %w", err)
	}

	socketsDir := filepath.Join(homeDir, ".addt", "sockets")
	if err := os.MkdirAll(socketsDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create sockets dir: %w", err)
	}

	// Create unique subdirectory for this proxy instance
	proxyDir, err := os.MkdirTemp(socketsDir, "tmux-proxy-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Set restrictive permissions
	if err := os.Chmod(proxyDir, 0700); err != nil {
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to set permissions: %w", err)
	}

	// Write PID file for cleanup
	if err := security.WritePIDFile(proxyDir); err != nil {
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to write PID file: %w", err)
	}

	// Use same socket name as original for compatibility
	socketName := filepath.Base(upstreamSocket)
	proxySocket := filepath.Join(proxyDir, socketName)

	// Create the proxy listener
	listener, err := net.Listen("unix", proxySocket)
	if err != nil {
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to create proxy socket: %w", err)
	}

	if err := os.Chmod(proxySocket, 0600); err != nil {
		listener.Close()
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to set socket permissions: %w", err)
	}

	proxy := &tmuxProxy{
		listener:       listener,
		upstreamSocket: upstreamSocket,
		proxyDir:       proxyDir,
		running:        true,
	}

	// Store for cleanup
	p.tmuxProxy = proxy

	// Start accepting connections
	go proxy.acceptLoop()

	return proxyDir, proxySocket, nil
}

func (tp *tmuxProxy) acceptLoop() {
	for {
		conn, err := tp.listener.Accept()
		if err != nil {
			tp.mu.Lock()
			running := tp.running
			tp.mu.Unlock()
			if !running {
				return
			}
			continue
		}

		tp.wg.Add(1)
		go tp.handleConnection(conn)
	}
}

func (tp *tmuxProxy) handleConnection(client net.Conn) {
	defer tp.wg.Done()
	defer client.Close()

	// Connect to upstream tmux socket
	upstream, err := net.Dial("unix", tp.upstreamSocket)
	if err != nil {
		return
	}
	defer upstream.Close()

	// Bidirectional proxy
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(upstream, client)
	}()

	go func() {
		defer wg.Done()
		io.Copy(client, upstream)
	}()

	wg.Wait()
}

func (tp *tmuxProxy) Stop() {
	tp.mu.Lock()
	tp.running = false
	tp.mu.Unlock()

	if tp.listener != nil {
		tp.listener.Close()
	}
	tp.wg.Wait()
	os.RemoveAll(tp.proxyDir)
}
