package cmd

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jedi4ever/nddt/extensions"
	"gopkg.in/yaml.v3"
)

// ExtensionConfig represents the config.yaml structure
type ExtensionConfig struct {
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	Entrypoint     string   `yaml:"entrypoint"`
	DefaultVersion string   `yaml:"default_version"`
	AutoMount      bool     `yaml:"auto_mount"`
	Dependencies   []string `yaml:"dependencies"`
}

// ListExtensions prints all available extensions
func ListExtensions() {
	exts, err := getExtensions()
	if err != nil {
		fmt.Printf("Error reading extensions: %v\n", err)
		return
	}

	// Find max lengths for alignment
	maxName := 4 // "Name"
	maxEntry := 10 // "Entrypoint"
	maxVer := 7 // "Version"
	for _, ext := range exts {
		if len(ext.Name) > maxName {
			maxName = len(ext.Name)
		}
		if len(ext.Entrypoint) > maxEntry {
			maxEntry = len(ext.Entrypoint)
		}
		ver := ext.DefaultVersion
		if ver == "" {
			ver = "latest"
		}
		if len(ver) > maxVer {
			maxVer = len(ver)
		}
	}

	// Print header
	fmt.Printf("  #  %-*s  %-*s  %-*s  %s\n", maxName, "Name", maxEntry, "Entrypoint", maxVer, "Version", "Description")
	fmt.Printf("  -  %-*s  %-*s  %-*s  %s\n", maxName, strings.Repeat("-", maxName), maxEntry, strings.Repeat("-", maxEntry), maxVer, strings.Repeat("-", maxVer), "-----------")

	// Print rows
	for i, ext := range exts {
		version := ext.DefaultVersion
		if version == "" {
			version = "latest"
		}
		fmt.Printf("%3d  %-*s  %-*s  %-*s  %s\n", i+1, maxName, ext.Name, maxEntry, ext.Entrypoint, maxVer, version, ext.Description)
	}

	fmt.Println()
	fmt.Println("Usage: ln -s /usr/local/bin/nddt ~/bin/<entrypoint>")
}

// getExtensions reads all extension configs from embedded filesystem
func getExtensions() ([]ExtensionConfig, error) {
	var configs []ExtensionConfig

	entries, err := fs.ReadDir(extensions.FS, ".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		configPath := entry.Name() + "/config.yaml"
		data, err := extensions.FS.ReadFile(configPath)
		if err != nil {
			continue // Skip directories without config.yaml
		}

		var cfg ExtensionConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			continue // Skip invalid configs
		}

		configs = append(configs, cfg)
	}

	// Sort by name
	sort.Slice(configs, func(i, j int) bool {
		return configs[i].Name < configs[j].Name
	})

	return configs, nil
}
