package config

import (
	"fmt"
	"os"
	"strings"

	cfgtypes "github.com/jedi4ever/addt/config"
	"github.com/jedi4ever/addt/extensions"
)

func listExtension(extName string, useGlobal bool) {
	// Get extension defaults from extension's config.yaml
	var extDefaults *extensions.ExtensionConfig
	exts, err := extensions.GetExtensions()
	if err == nil {
		for _, ext := range exts {
			if ext.Name == extName {
				extDefaults = &ext
				break
			}
		}
	}

	extNameUpper := strings.ToUpper(extName)
	scope := "project"
	if useGlobal {
		scope = "global"
	}
	fmt.Printf("Extension: %s (%s)\n\n", extName, scope)

	keys := GetExtensionKeys()

	// Load the appropriate config
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

	var extCfg *cfgtypes.ExtensionSettings
	if cfg.Extensions != nil {
		extCfg = cfg.Extensions[extName]
	}

	// Print header
	fmt.Printf("  %-10s   %-15s   %s\n", "Key", "Value", "Source")
	fmt.Printf("  %s   %s   %s\n", strings.Repeat("-", 10), strings.Repeat("-", 15), "--------")

	for _, k := range keys {
		envVar := fmt.Sprintf(k.EnvVar, extNameUpper)
		envValue := os.Getenv(envVar)

		var configValue, defaultValue string

		// Get config value
		if extCfg != nil {
			switch k.Key {
			case "version":
				configValue = extCfg.Version
			case "automount":
				if extCfg.Automount != nil {
					configValue = fmt.Sprintf("%v", *extCfg.Automount)
				}
			}
		}

		// Get extension default value
		if extDefaults != nil {
			switch k.Key {
			case "version":
				defaultValue = extDefaults.DefaultVersion
			case "automount":
				defaultValue = fmt.Sprintf("%v", extDefaults.AutoMount)
			}
		}

		// Determine effective value and source (env > config > default)
		var displayValue, source string
		if envValue != "" {
			displayValue = envValue
			source = "env"
		} else if configValue != "" {
			displayValue = configValue
			source = scope
		} else if defaultValue != "" {
			displayValue = defaultValue
			source = "default"
		} else {
			displayValue = "-"
			source = ""
		}

		if source == "env" || source == scope {
			fmt.Printf("* %-10s   %-15s   %s\n", k.Key, displayValue, source)
		} else {
			fmt.Printf("  %-10s   %-15s   %s\n", k.Key, displayValue, source)
		}
	}
}

func getExtension(extName, key string, useGlobal bool) {
	if !IsValidExtensionKey(key) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Println("Available keys: version, automount")
		os.Exit(1)
	}

	var cfg *cfgtypes.GlobalConfig
	var err error
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	var extCfg *cfgtypes.ExtensionSettings
	if cfg.Extensions != nil {
		extCfg = cfg.Extensions[extName]
	}

	if extCfg == nil {
		fmt.Printf("%s is not set\n", key)
		return
	}

	var val string
	switch key {
	case "version":
		val = extCfg.Version
	case "automount":
		if extCfg.Automount != nil {
			val = fmt.Sprintf("%v", *extCfg.Automount)
		}
	}

	if val == "" {
		fmt.Printf("%s is not set\n", key)
	} else {
		fmt.Println(val)
	}
}

func setExtension(extName, key, value string, useGlobal bool) {
	if !IsValidExtensionKey(key) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Println("Available keys: version, automount")
		os.Exit(1)
	}

	// Validate bool values
	if key == "automount" {
		value = strings.ToLower(value)
		if value != "true" && value != "false" {
			fmt.Printf("Invalid value for %s: must be 'true' or 'false'\n", key)
			os.Exit(1)
		}
	}

	var cfg *cfgtypes.GlobalConfig
	var err error
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize extensions map if needed
	if cfg.Extensions == nil {
		cfg.Extensions = make(map[string]*cfgtypes.ExtensionSettings)
	}

	// Initialize extension config if needed
	if cfg.Extensions[extName] == nil {
		cfg.Extensions[extName] = &cfgtypes.ExtensionSettings{}
	}

	extCfg := cfg.Extensions[extName]
	switch key {
	case "version":
		extCfg.Version = value
	case "automount":
		b := value == "true"
		extCfg.Automount = &b
	}

	scope := "project"
	if useGlobal {
		if err := cfgtypes.SaveGlobalConfigFile(cfg); err != nil {
			fmt.Printf("Error saving global config: %v\n", err)
			os.Exit(1)
		}
		scope = "global"
	} else {
		if err := cfgtypes.SaveProjectConfigFile(cfg); err != nil {
			fmt.Printf("Error saving project config: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Set %s.%s = %s (%s)\n", extName, key, value, scope)
}

func unsetExtension(extName, key string, useGlobal bool) {
	if !IsValidExtensionKey(key) {
		fmt.Printf("Unknown extension config key: %s\n", key)
		fmt.Println("Available keys: version, automount")
		os.Exit(1)
	}

	var cfg *cfgtypes.GlobalConfig
	var err error
	if useGlobal {
		cfg, err = cfgtypes.LoadGlobalConfigFile()
	} else {
		cfg, err = cfgtypes.LoadProjectConfigFile()
	}
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	scope := "project"
	if useGlobal {
		scope = "global"
	}

	if cfg.Extensions == nil || cfg.Extensions[extName] == nil {
		fmt.Printf("%s.%s is not set in %s config\n", extName, key, scope)
		return
	}

	extCfg := cfg.Extensions[extName]
	switch key {
	case "version":
		extCfg.Version = ""
	case "automount":
		extCfg.Automount = nil
	}

	// Clean up empty extension config
	if extCfg.Version == "" && extCfg.Automount == nil {
		delete(cfg.Extensions, extName)
	}

	// Clean up empty extensions map
	if len(cfg.Extensions) == 0 {
		cfg.Extensions = nil
	}

	if useGlobal {
		if err := cfgtypes.SaveGlobalConfigFile(cfg); err != nil {
			fmt.Printf("Error saving global config: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := cfgtypes.SaveProjectConfigFile(cfg); err != nil {
			fmt.Printf("Error saving project config: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Printf("Unset %s.%s (%s)\n", extName, key, scope)
}
