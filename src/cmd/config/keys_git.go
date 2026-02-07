package config

import (
	"fmt"

	cfgtypes "github.com/jedi4ever/addt/config"
)

// GetGitKeys returns all valid git config keys
func GetGitKeys() []KeyInfo {
	return []KeyInfo{
		{Key: "git.forward_config", Description: "Forward .gitconfig to container (default: true)", Type: "bool", EnvVar: "ADDT_GIT_FORWARD_CONFIG"},
		{Key: "git.config_path", Description: "Custom .gitconfig file path", Type: "string", EnvVar: "ADDT_GIT_CONFIG_PATH"},
	}
}

// GetGitValue retrieves a git config value
func GetGitValue(g *cfgtypes.GitSettings, key string) string {
	if g == nil {
		return ""
	}
	switch key {
	case "git.forward_config":
		if g.ForwardConfig != nil {
			return fmt.Sprintf("%v", *g.ForwardConfig)
		}
	case "git.config_path":
		return g.ConfigPath
	}
	return ""
}

// SetGitValue sets a git config value
func SetGitValue(g *cfgtypes.GitSettings, key, value string) {
	switch key {
	case "git.forward_config":
		b := value == "true"
		g.ForwardConfig = &b
	case "git.config_path":
		g.ConfigPath = value
	}
}

// UnsetGitValue clears a git config value
func UnsetGitValue(g *cfgtypes.GitSettings, key string) {
	switch key {
	case "git.forward_config":
		g.ForwardConfig = nil
	case "git.config_path":
		g.ConfigPath = ""
	}
}
