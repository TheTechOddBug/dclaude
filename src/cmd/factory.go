package cmd

import (
	"fmt"

	"github.com/jedi4ever/nddt/assets"
	"github.com/jedi4ever/nddt/extensions"
	"github.com/jedi4ever/nddt/provider"
	"github.com/jedi4ever/nddt/provider/daytona"
	"github.com/jedi4ever/nddt/provider/docker"
)

// NewProvider creates a new provider based on the specified type
func NewProvider(providerType string, cfg *provider.Config) (provider.Provider, error) {
	switch providerType {
	case "docker", "":
		return docker.NewDockerProvider(cfg, assets.DockerDockerfile, assets.DockerEntrypoint, assets.DockerInitFirewall, assets.DockerInstallSh, extensions.FS)
	case "daytona":
		return daytona.NewDaytonaProvider(cfg, assets.DaytonaDockerfile, assets.DaytonaEntrypoint)
	default:
		return nil, fmt.Errorf("unknown provider type: %s (supported: docker, daytona)", providerType)
	}
}
