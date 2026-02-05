package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DetectContainerRuntime automatically detects which container runtime to use
// Priority: explicit ADDT_PROVIDER > Podman (if available) > Docker (if running) > Podman (default)
func DetectContainerRuntime() string {
	// If explicitly set, use that
	if provider := os.Getenv("ADDT_PROVIDER"); provider != "" {
		return provider
	}

	// Check if Podman is available (preferred - no daemon required)
	if isPodmanAvailable() {
		return "podman"
	}

	// Check if Docker is available and running
	if isDockerRunning() {
		return "docker"
	}

	// Default to podman (will offer to install if not available)
	return "podman"
}

// EnsureContainerRuntime ensures a container runtime is available
// Downloads Podman automatically if needed (unless Docker is explicitly selected)
func EnsureContainerRuntime() (string, error) {
	// If Docker is explicitly selected, use it without auto-download
	if provider := os.Getenv("ADDT_PROVIDER"); provider == "docker" {
		if !isDockerRunning() {
			return "", fmt.Errorf("Docker is explicitly selected but not running")
		}
		return "docker", nil
	}

	// Check if Podman is already available
	if isPodmanAvailable() {
		return "podman", nil
	}

	// Check if Docker is available as fallback
	if isDockerRunning() {
		return "docker", nil
	}

	// Neither available - auto-download Podman
	fmt.Println("No container runtime found. Downloading Podman...")
	if err := DownloadPodman(); err != nil {
		return "", fmt.Errorf("failed to download Podman: %w", err)
	}

	// Verify it works now
	if isPodmanAvailable() {
		return "podman", nil
	}

	return "", fmt.Errorf("Podman downloaded but not working")
}

// isDockerRunning checks if Docker daemon is running
func isDockerRunning() bool {
	// First check if docker command exists
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return false
	}

	// Check if daemon is responsive
	cmd := exec.Command(dockerPath, "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// isPodmanAvailable checks if Podman is available (no daemon needed)
// Checks both system Podman and bundled Podman
func isPodmanAvailable() bool {
	podmanPath := GetPodmanPath()
	if podmanPath == "" {
		return false
	}

	// Podman doesn't need a daemon, just check version works
	cmd := exec.Command(podmanPath, "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// GetPodmanPath returns the path to Podman binary (system or bundled)
func GetPodmanPath() string {
	// First check system Podman
	if path, err := exec.LookPath("podman"); err == nil {
		return path
	}

	// Check bundled Podman
	bundledPath := GetBundledPodmanPath()
	if bundledPath != "" {
		if _, err := os.Stat(bundledPath); err == nil {
			return bundledPath
		}
	}

	return ""
}

// GetRuntimeInfo returns information about the detected runtime
func GetRuntimeInfo() (runtime string, version string, extras []string) {
	runtime = DetectContainerRuntime()

	switch runtime {
	case "docker":
		version = getDockerVersion()
	case "podman":
		version = getPodmanVersion()
		if hasPasta() {
			extras = append(extras, "pasta")
		}
	}

	return runtime, version, extras
}

func getDockerVersion() string {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func getPodmanVersion() string {
	podmanPath := GetPodmanPath()
	if podmanPath == "" {
		return "unknown"
	}
	// Use --version flag which works without daemon connection
	cmd := exec.Command(podmanPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	// Parse "podman version X.Y.Z" -> "X.Y.Z"
	version := strings.TrimSpace(string(output))
	return strings.TrimPrefix(version, "podman version ")
}

func hasPasta() bool {
	_, err := exec.LookPath("pasta")
	return err == nil
}
