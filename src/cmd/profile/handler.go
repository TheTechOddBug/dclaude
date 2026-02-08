package profile

import (
	"fmt"
	"os"
	"sort"
	"strings"

	cfgcmd "github.com/jedi4ever/addt/cmd/config"
	cfgtypes "github.com/jedi4ever/addt/config"
)

// HandleCommand handles the profile subcommand
func HandleCommand(args []string) {
	if len(args) == 0 {
		printHelp()
		return
	}

	switch args[0] {
	case "list":
		listProfiles()
	case "show":
		if len(args) < 2 {
			fmt.Println("Usage: addt profile show <name>")
			fmt.Println()
			printAvailableProfiles()
			os.Exit(1)
		}
		showProfile(args[1])
	case "apply":
		if len(args) < 2 {
			fmt.Println("Usage: addt profile apply <name> [-g]")
			fmt.Println()
			printAvailableProfiles()
			os.Exit(1)
		}
		remaining := args[1:]
		useGlobal := false
		var name string
		for _, arg := range remaining {
			if arg == "-g" || arg == "--global" {
				useGlobal = true
			} else {
				name = arg
			}
		}
		if name == "" {
			fmt.Println("Usage: addt profile apply <name> [-g]")
			os.Exit(1)
		}
		applyProfile(name, useGlobal)
	case "-h", "--help", "help":
		printHelp()
	default:
		fmt.Printf("Unknown profile command: %s\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func listProfiles() {
	profiles, err := GetProfiles()
	if err != nil {
		fmt.Printf("Error loading profiles: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available profiles:")
	fmt.Println()
	for _, p := range profiles {
		fmt.Printf("  %-12s %s\n", p.Name, p.Description)
	}
	fmt.Println()
	fmt.Println("Use 'addt profile show <name>' to see what a profile sets.")
	fmt.Println("Use 'addt profile apply <name>' to apply a profile.")
}

func showProfile(name string) {
	p, err := GetProfile(name)
	if err != nil {
		fmt.Printf("Error loading profile: %v\n", err)
		os.Exit(1)
	}
	if p == nil {
		fmt.Printf("Unknown profile: %s\n", name)
		fmt.Println()
		printAvailableProfiles()
		os.Exit(1)
	}

	fmt.Printf("Profile: %s\n", p.Name)
	fmt.Printf("Description: %s\n", p.Description)
	fmt.Println()

	// Sort keys for consistent output
	keys := make([]string, 0, len(p.Settings))
	for k := range p.Settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Find max key length for alignment
	maxLen := 0
	for _, k := range keys {
		if len(k) > maxLen {
			maxLen = len(k)
		}
	}

	fmt.Printf("  %-*s  %-15s  %s\n", maxLen, "Key", "Profile", "Default")
	fmt.Printf("  %s  %s  %s\n", strings.Repeat("-", maxLen), strings.Repeat("-", 15), strings.Repeat("-", 15))

	for _, k := range keys {
		val := p.Settings[k]
		def := cfgcmd.GetDefaultValue(k)
		displayVal := val
		if displayVal == "" {
			displayVal = "(empty)"
		}
		displayDef := def
		if displayDef == "" {
			displayDef = "(empty)"
		}
		marker := " "
		if val != def {
			marker = "*"
		}
		fmt.Printf("%s %-*s  %-15s  %s\n", marker, maxLen, k, displayVal, displayDef)
	}
	fmt.Println()
	fmt.Println("* = differs from default")
}

func applyProfile(name string, useGlobal bool) {
	p, err := GetProfile(name)
	if err != nil {
		fmt.Printf("Error loading profile: %v\n", err)
		os.Exit(1)
	}
	if p == nil {
		fmt.Printf("Unknown profile: %s\n", name)
		fmt.Println()
		printAvailableProfiles()
		os.Exit(1)
	}

	// Load config
	var cfg *cfgtypes.GlobalConfig
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Validate all keys before applying
	for k := range p.Settings {
		if !cfgcmd.IsValidKey(k) {
			fmt.Printf("Error: profile contains invalid config key: %s\n", k)
			os.Exit(1)
		}
	}

	// Apply settings
	for k, v := range p.Settings {
		cfgcmd.SetValue(cfg, k, v)
	}

	// Save config
	if useGlobal {
		err = cfgtypes.SaveGlobalConfigFile(cfg)
	} else {
		err = cfgtypes.SaveProjectConfigFile(cfg)
	}
	if err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	scope := "project"
	if useGlobal {
		scope = "global"
	}
	fmt.Printf("Applied profile '%s' to %s config (%d settings)\n", name, scope, len(p.Settings))
}

func printAvailableProfiles() {
	profiles, err := GetProfiles()
	if err != nil {
		return
	}
	fmt.Println("Available profiles:")
	for _, p := range profiles {
		fmt.Printf("  %s\n", p.Name)
	}
}

func printHelp() {
	fmt.Println("Usage: addt profile <command>")
	fmt.Println()
	fmt.Println("Apply predefined configuration presets.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                  List available profiles")
	fmt.Println("  show <name>           Show what a profile would set")
	fmt.Println("  apply <name> [-g]     Apply profile to project (or global with -g)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -g, --global          Apply to global config instead of project")
	fmt.Println()

	profiles, err := GetProfiles()
	if err == nil && len(profiles) > 0 {
		fmt.Println("Profiles:")
		for _, p := range profiles {
			fmt.Printf("  %-12s %s\n", p.Name, p.Description)
		}
		fmt.Println()
	}

	fmt.Println("Examples:")
	fmt.Println("  addt profile list")
	fmt.Println("  addt profile show strict")
	fmt.Println("  addt profile apply strict")
	fmt.Println("  addt profile apply paranoia -g")
}
