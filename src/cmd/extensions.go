package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedi4ever/addt/extensions"
)

// PrintVersion prints addt version and loaded extension version
func PrintVersion(version, defaultNodeVersion, defaultGoVersion, defaultUvVersion string) {
	fmt.Printf("addt %s\n", version)
	fmt.Println()

	// Default tool versions
	fmt.Println("Tools:")
	fmt.Printf("  Node.js:  %s\n", defaultNodeVersion)
	fmt.Printf("  Go:       %s\n", defaultGoVersion)
	fmt.Printf("  UV:       %s\n", defaultUvVersion)
	fmt.Println()

	// Get loaded extension (from env or binary name)
	extName := os.Getenv("ADDT_EXTENSIONS")
	if extName == "" {
		extName = os.Getenv("ADDT_COMMAND")
	}
	// No default - extension must be explicitly set via symlink or env
	// Take first extension if comma-separated
	if idx := strings.Index(extName, ","); idx != -1 {
		extName = extName[:idx]
	}

	// Get version for this extension
	extVersion := os.Getenv("ADDT_" + strings.ToUpper(extName) + "_VERSION")
	if extVersion == "" {
		// Look up default version from config
		exts, err := extensions.GetExtensions()
		if err == nil {
			for _, ext := range exts {
				if ext.Name == extName {
					extVersion = ext.DefaultVersion
					break
				}
			}
		}
	}
	if extVersion == "" {
		extVersion = "latest"
	}

	fmt.Println("Extension:")
	fmt.Printf("  %-16s %s\n", extName, extVersion)
}

// ListExtensions prints all available extensions
func ListExtensions() {
	exts, err := extensions.GetExtensions()
	if err != nil {
		fmt.Printf("Error reading extensions: %v\n", err)
		return
	}

	// Find max lengths for alignment
	maxName := 4   // "Name"
	maxEntry := 10 // "Entrypoint"
	maxVer := 7    // "Version"
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
	fmt.Printf("  #  %-*s  %-*s  %-*s  %-6s  %s\n", maxName, "Name", maxEntry, "Entrypoint", maxVer, "Version", "Source", "Description")
	fmt.Printf("  -  %-*s  %-*s  %-*s  %-6s  %s\n", maxName, strings.Repeat("-", maxName), maxEntry, strings.Repeat("-", maxEntry), maxVer, strings.Repeat("-", maxVer), "------", "-----------")

	// Print rows
	for i, ext := range exts {
		version := ext.DefaultVersion
		if version == "" {
			version = "latest"
		}
		source := "built-in"
		if ext.IsLocal {
			source = "local"
		}
		fmt.Printf("%3d  %-*s  %-*s  %-*s  %-6s  %s\n", i+1, maxName, ext.Name, maxEntry, ext.Entrypoint, maxVer, version, source, ext.Description)
	}

	// Show local extensions directory info
	localDir := extensions.GetLocalExtensionsDir()
	if localDir != "" {
		fmt.Printf("\nLocal extensions directory: %s\n", localDir)
	}
}

// ShowExtensionInfo displays detailed info about a specific extension
func ShowExtensionInfo(name string) {
	exts, err := extensions.GetExtensions()
	if err != nil {
		fmt.Printf("Error reading extensions: %v\n", err)
		os.Exit(1)
	}

	for _, ext := range exts {
		if ext.Name == name {
			version := ext.DefaultVersion
			if version == "" {
				version = "latest"
			}

			fmt.Printf("%s\n", ext.Name)
			fmt.Printf("%s\n\n", strings.Repeat("=", len(ext.Name)))

			fmt.Printf("  %s\n\n", ext.Description)

			source := "built-in"
			if ext.IsLocal {
				source = "local (~/.addt/extensions/" + ext.Name + ")"
			}

			fmt.Println("Configuration:")
			fmt.Printf("  Entrypoint:  %s\n", ext.Entrypoint)
			fmt.Printf("  Version:     %s\n", version)
			fmt.Printf("  Auto-mount:  %v\n", ext.AutoMount)
			fmt.Printf("  Source:      %s\n", source)

			if len(ext.Dependencies) > 0 {
				fmt.Printf("  Depends on:  %s\n", strings.Join(ext.Dependencies, ", "))
			}

			if len(ext.EnvVars) > 0 {
				fmt.Println("\nEnvironment Variables:")
				for _, env := range ext.EnvVars {
					fmt.Printf("  - %s\n", env)
				}
			}

			if len(ext.Mounts) > 0 {
				fmt.Println("\nMounts:")
				for _, m := range ext.Mounts {
					fmt.Printf("  - %s -> %s\n", m.Source, m.Target)
				}
			}

			if len(ext.Flags) > 0 {
				fmt.Println("\nFlags:")
				for _, f := range ext.Flags {
					fmt.Printf("  %-12s %s\n", f.Flag, f.Description)
				}
			}

			fmt.Println("\nUsage:")
			fmt.Printf("  addt run %s [args...]\n", ext.Name)
			return
		}
	}

	fmt.Printf("Extension not found: %s\n", name)
	fmt.Println("Run 'addt extensions list' to see available extensions")
	os.Exit(1)
}

// GetEntrypointForExtension returns the entrypoint command for a given extension name
// If extension not found, returns the extension name itself as fallback
func GetEntrypointForExtension(extName string) string {
	exts, err := extensions.GetExtensions()
	if err != nil {
		return extName
	}

	for _, ext := range exts {
		if ext.Name == extName {
			if ext.Entrypoint != "" {
				return ext.Entrypoint
			}
			return extName
		}
	}

	return extName
}

// extensionExists checks if an extension with the given name exists
func extensionExists(name string) bool {
	exts, err := extensions.GetExtensions()
	if err != nil {
		return false
	}
	for _, ext := range exts {
		if ext.Name == name {
			return true
		}
	}
	return false
}

// CreateExtension creates a new local extension with template files
func CreateExtension(name string) {
	// Validate name
	if name == "" {
		fmt.Println("Error: extension name cannot be empty")
		os.Exit(1)
	}

	// Check if name contains only valid characters
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			fmt.Printf("Error: extension name can only contain lowercase letters, numbers, hyphens, and underscores\n")
			os.Exit(1)
		}
	}

	// Get local extensions directory
	localDir := extensions.GetLocalExtensionsDir()
	if localDir == "" {
		fmt.Println("Error: could not determine local extensions directory")
		os.Exit(1)
	}

	extDir := filepath.Join(localDir, name)

	// Check if extension already exists
	if _, err := os.Stat(extDir); err == nil {
		fmt.Printf("Error: extension '%s' already exists at %s\n", name, extDir)
		os.Exit(1)
	}

	// Create extension directory
	if err := os.MkdirAll(extDir, 0755); err != nil {
		fmt.Printf("Error: failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Create config.yaml
	configContent := fmt.Sprintf(`name: %s
description: Description of your extension
entrypoint: %s
default_version: latest
auto_mount: false
dependencies: []
env_vars: []
mounts: []
flags: []
`, name, name)

	if err := os.WriteFile(filepath.Join(extDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		fmt.Printf("Error: failed to create config.yaml: %v\n", err)
		os.Exit(1)
	}

	// Create install.sh
	installContent := fmt.Sprintf(`#!/bin/bash
# %s - Installation script
# This script runs during 'addt build %s'

set -e

echo "Extension [%s]: Installing..."

# Get version from environment or default to latest
VERSION="${%s_VERSION:-latest}"

# TODO: Add your installation commands here
# Examples:
#   sudo npm install -g your-package
#   pip install your-package
#   go install github.com/your/package@latest

echo "Extension [%s]: Done."
`, name, name, name, strings.ToUpper(strings.ReplaceAll(name, "-", "_")), name)

	if err := os.WriteFile(filepath.Join(extDir, "install.sh"), []byte(installContent), 0755); err != nil {
		fmt.Printf("Error: failed to create install.sh: %v\n", err)
		os.Exit(1)
	}

	// Create setup.sh
	setupContent := fmt.Sprintf(`#!/bin/bash
# %s - Setup script
# This script runs at container startup before the entrypoint

echo "Setup [%s]: Initializing environment"

# TODO: Add any runtime setup commands here
# Examples:
#   export MY_VAR="value"
#   source ~/.my-config
`, name, name)

	if err := os.WriteFile(filepath.Join(extDir, "setup.sh"), []byte(setupContent), 0755); err != nil {
		fmt.Printf("Error: failed to create setup.sh: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created extension '%s' at:\n", name)
	fmt.Printf("  %s\n", extDir)
	fmt.Println()
	fmt.Println("Files created:")
	fmt.Println("  config.yaml  - Extension configuration")
	fmt.Println("  install.sh   - Installation script (runs during build)")
	fmt.Println("  setup.sh     - Setup script (runs at container start)")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit %s/config.yaml to configure your extension\n", extDir)
	fmt.Printf("  2. Edit %s/install.sh to add installation commands\n", extDir)
	fmt.Printf("  3. Build with: addt build %s\n", name)
	fmt.Printf("  4. Run with:   addt run %s\n", name)
}
