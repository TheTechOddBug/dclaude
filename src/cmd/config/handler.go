package config

import (
	"fmt"
	"os"

	cfgtypes "github.com/jedi4ever/addt/config"
	"github.com/jedi4ever/addt/extensions"
)

// HandleCommand handles the config subcommand
func HandleCommand(args []string) {
	if len(args) == 0 {
		printHelp()
		return
	}

	switch args[0] {
	case "global":
		handleGlobal(args[1:])
	case "project":
		handleProject(args[1:])
	case "extension":
		handleExtension(args[1:])
	case "path":
		fmt.Printf("Global config:  %s\n", cfgtypes.GetGlobalConfigPath())
		fmt.Printf("Project config: %s\n", cfgtypes.GetProjectConfigPath())
	default:
		fmt.Printf("Unknown config command: %s\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

// handleGlobal handles global config subcommands
func handleGlobal(args []string) {
	if len(args) == 0 {
		printGlobalHelp()
		return
	}

	switch args[0] {
	case "list":
		listGlobal()
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: addt config global get <key>")
			os.Exit(1)
		}
		getGlobal(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: addt config global set <key> <value>")
			os.Exit(1)
		}
		setGlobal(args[1], args[2])
	case "unset":
		if len(args) < 2 {
			fmt.Println("Usage: addt config global unset <key>")
			os.Exit(1)
		}
		unsetGlobal(args[1])
	default:
		fmt.Printf("Unknown global config command: %s\n", args[0])
		printGlobalHelp()
		os.Exit(1)
	}
}

// handleProject handles project-level config subcommands
func handleProject(args []string) {
	if len(args) == 0 {
		printProjectHelp()
		return
	}

	switch args[0] {
	case "list":
		listProject()
	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: addt config project get <key>")
			os.Exit(1)
		}
		getProject(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: addt config project set <key> <value>")
			os.Exit(1)
		}
		setProject(args[1], args[2])
	case "unset":
		if len(args) < 2 {
			fmt.Println("Usage: addt config project unset <key>")
			os.Exit(1)
		}
		unsetProject(args[1])
	default:
		fmt.Printf("Unknown project config command: %s\n", args[0])
		printProjectHelp()
		os.Exit(1)
	}
}

// handleExtension handles extension-specific config subcommands
func handleExtension(args []string) {
	if len(args) == 0 {
		printExtensionHelp()
		return
	}

	// Check for --project flag anywhere in args
	useProject := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--project" {
			useProject = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	if len(args) == 0 {
		printExtensionHelp()
		return
	}

	extName := args[0]

	// Check if first arg is a subcommand (user forgot extension name)
	if extName == "list" || extName == "get" || extName == "set" || extName == "unset" {
		fmt.Println("Error: extension name required")
		fmt.Println()
		printExtensionHelp()
		os.Exit(1)
	}

	// Validate that the extension exists
	if !extensionExists(extName) {
		fmt.Printf("Error: extension '%s' does not exist\n", extName)
		fmt.Println("Run 'addt extensions list' to see available extensions")
		os.Exit(1)
	}

	if len(args) < 2 {
		// Default to list for extension
		listExtension(extName)
		return
	}

	switch args[1] {
	case "list":
		listExtension(extName)
	case "get":
		if len(args) < 3 {
			fmt.Println("Usage: addt config extension <name> get <key>")
			os.Exit(1)
		}
		getExtension(extName, args[2])
	case "set":
		if len(args) < 4 {
			fmt.Println("Usage: addt config extension <name> set <key> <value> [--project]")
			os.Exit(1)
		}
		setExtension(extName, args[2], args[3], useProject)
	case "unset":
		if len(args) < 3 {
			fmt.Println("Usage: addt config extension <name> unset <key> [--project]")
			os.Exit(1)
		}
		unsetExtension(extName, args[2], useProject)
	default:
		fmt.Printf("Unknown extension config command: %s\n", args[1])
		printExtensionHelp()
		os.Exit(1)
	}
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

func printHelp() {
	fmt.Println("Usage: addt config <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  global <subcommand>              Manage global configuration (~/.addt/config.yaml)")
	fmt.Println("  project <subcommand>             Manage project configuration (.addt.yaml)")
	fmt.Println("  extension <name> <subcommand>    Manage extension-specific configuration")
	fmt.Println("  path                             Show config file paths")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  addt config global list")
	fmt.Println("  addt config global set docker_cpus 2")
	fmt.Println("  addt config project set persistent true")
	fmt.Println("  addt config extension claude set version 1.0.5")
	fmt.Println()
	fmt.Println("Precedence (highest to lowest):")
	fmt.Println("  1. Environment variables (e.g., ADDT_DOCKER_CPUS)")
	fmt.Println("  2. Project config (.addt.yaml in current directory)")
	fmt.Println("  3. Global config (~/.addt/config.yaml)")
	fmt.Println("  4. Default values")
}

func printGlobalHelp() {
	fmt.Println("Usage: addt config global <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List all global configuration values")
	fmt.Println("  get <key>         Get a configuration value")
	fmt.Println("  set <key> <value> Set a configuration value")
	fmt.Println("  unset <key>       Remove a configuration value (use default)")
	fmt.Println()
	fmt.Println("Available keys:")
	keys := GetKeys()
	maxKeyLen := 0
	for _, k := range keys {
		if len(k.Key) > maxKeyLen {
			maxKeyLen = len(k.Key)
		}
	}
	for _, k := range keys {
		fmt.Printf("  %-*s  %s\n", maxKeyLen, k.Key, k.Description)
	}
}

func printProjectHelp() {
	fmt.Println("Usage: addt config project <command>")
	fmt.Println()
	fmt.Println("Manage project-level configuration stored in .addt.yaml")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List all project configuration values")
	fmt.Println("  get <key>         Get a configuration value")
	fmt.Println("  set <key> <value> Set a configuration value")
	fmt.Println("  unset <key>       Remove a configuration value")
	fmt.Println()
	fmt.Println("Project config overrides global config but is overridden by env vars.")
	fmt.Println()
	fmt.Println("Available keys:")
	keys := GetKeys()
	maxKeyLen := 0
	for _, k := range keys {
		if len(k.Key) > maxKeyLen {
			maxKeyLen = len(k.Key)
		}
	}
	for _, k := range keys {
		fmt.Printf("  %-*s  %s\n", maxKeyLen, k.Key, k.Description)
	}
}

func printExtensionHelp() {
	fmt.Println("Usage: addt config extension <name> <command> [--project]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list              List extension configuration")
	fmt.Println("  get <key>         Get a configuration value")
	fmt.Println("  set <key> <value> Set a configuration value")
	fmt.Println("  unset <key>       Remove a configuration value")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --project         Save to project config (.addt.yaml) instead of global")
	fmt.Println()
	fmt.Println("Available keys:")
	fmt.Println("  version     Extension version (e.g., \"1.0.5\", \"latest\", \"stable\")")
	fmt.Println("  automount   Auto-mount extension config directories (true/false)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  addt config extension claude list")
	fmt.Println("  addt config extension claude set version 1.0.5")
	fmt.Println("  addt config extension claude set automount false --project")
}
