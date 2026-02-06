package docker

import (
	"crypto/sha256"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

// ImageExists checks if a Docker image exists
func (p *DockerProvider) ImageExists(imageName string) bool {
	cmd := exec.Command("docker", "image", "inspect", imageName)
	return cmd.Run() == nil
}

// FindImageByLabel finds an image by a specific label value
func (p *DockerProvider) FindImageByLabel(label, value string) string {
	cmd := exec.Command("docker", "images",
		"--filter", fmt.Sprintf("label=%s=%s", label, value),
		"--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" && !strings.Contains(line, "<none>") {
			return line
		}
	}
	return ""
}

// GetImageLabel retrieves a specific label value from an image
func (p *DockerProvider) GetImageLabel(imageName, label string) string {
	cmd := exec.Command("docker", "inspect",
		"--format", fmt.Sprintf("{{index .Config.Labels %q}}", label),
		imageName)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// assetsHash returns a short hash of the embedded build assets (Dockerfile.base, entrypoint, firewall)
// Used in image tags so that changes to these files trigger a rebuild
func (p *DockerProvider) assetsHash() string {
	h := sha256.New()
	h.Write(p.embeddedDockerfileBase)
	h.Write(p.embeddedEntrypoint)
	h.Write(p.embeddedInitFirewall)
	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}

// GetBaseImageName returns the base image name for the current config
func (p *DockerProvider) GetBaseImageName() string {
	currentUser, err := user.Current()
	if err != nil {
		return "addt-base:latest"
	}
	return fmt.Sprintf("addt-base:v%s-node%s-go%s-uv%s-uid%s-%s",
		p.config.AddtVersion, p.config.NodeVersion, p.config.GoVersion, p.config.UvVersion, currentUser.Uid, p.assetsHash())
}
