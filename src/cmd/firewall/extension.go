package firewall

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/config"
)

func handleExtension(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: addt firewall extension <name> <command>")
		fmt.Println("Commands: allow, deny, remove, list, reset")
		return
	}

	extName := args[0]
	cmd := args[1]
	cfg := config.LoadGlobalConfig()

	// Ensure extensions map exists
	if cfg.Extensions == nil {
		cfg.Extensions = make(map[string]*config.ExtensionSettings)
	}
	if cfg.Extensions[extName] == nil {
		cfg.Extensions[extName] = &config.ExtensionSettings{}
	}
	ext := cfg.Extensions[extName]

	switch cmd {
	case "allow":
		extensionAllow(cfg, ext, extName, args)
	case "deny":
		extensionDeny(cfg, ext, extName, args)
	case "remove", "rm":
		extensionRemove(cfg, ext, extName, args)
	case "list", "ls":
		extensionList(ext, extName)
	case "reset":
		extensionReset(cfg, ext, extName)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Commands: allow, deny, remove, list, reset")
	}
}

func extensionAllow(cfg *config.GlobalConfig, ext *config.ExtensionSettings, extName string, args []string) {
	if len(args) < 3 {
		fmt.Printf("Usage: addt firewall extension %s allow <domain>\n", extName)
		os.Exit(1)
	}
	domain := strings.TrimSpace(args[2])
	if containsString(ext.FirewallAllowed, domain) {
		fmt.Printf("Domain '%s' already in %s allowed list\n", domain, extName)
		return
	}
	ext.FirewallAllowed = append(ext.FirewallAllowed, domain)
	saveGlobalConfig(cfg)
	fmt.Printf("Added '%s' to %s allowed domains\n", domain, extName)
}

func extensionDeny(cfg *config.GlobalConfig, ext *config.ExtensionSettings, extName string, args []string) {
	if len(args) < 3 {
		fmt.Printf("Usage: addt firewall extension %s deny <domain>\n", extName)
		os.Exit(1)
	}
	domain := strings.TrimSpace(args[2])
	if containsString(ext.FirewallDenied, domain) {
		fmt.Printf("Domain '%s' already in %s denied list\n", domain, extName)
		return
	}
	ext.FirewallDenied = append(ext.FirewallDenied, domain)
	saveGlobalConfig(cfg)
	fmt.Printf("Added '%s' to %s denied domains\n", domain, extName)
}

func extensionRemove(cfg *config.GlobalConfig, ext *config.ExtensionSettings, extName string, args []string) {
	if len(args) < 3 {
		fmt.Printf("Usage: addt firewall extension %s remove <domain>\n", extName)
		os.Exit(1)
	}
	domain := strings.TrimSpace(args[2])
	removed := false

	newAllowed := removeString(ext.FirewallAllowed, domain)
	if len(newAllowed) < len(ext.FirewallAllowed) {
		ext.FirewallAllowed = newAllowed
		removed = true
		fmt.Printf("Removed '%s' from %s allowed domains\n", domain, extName)
	}

	newDenied := removeString(ext.FirewallDenied, domain)
	if len(newDenied) < len(ext.FirewallDenied) {
		ext.FirewallDenied = newDenied
		removed = true
		fmt.Printf("Removed '%s' from %s denied domains\n", domain, extName)
	}

	if removed {
		saveGlobalConfig(cfg)
	} else {
		fmt.Printf("Domain '%s' not found in %s config\n", domain, extName)
	}
}

func extensionList(ext *config.ExtensionSettings, extName string) {
	fmt.Printf("Extension '%s' firewall rules:\n", extName)
	printDomainList("  Allowed", ext.FirewallAllowed, nil, ext.FirewallDenied)
	fmt.Printf("  Denied:\n")
	if len(ext.FirewallDenied) == 0 {
		fmt.Printf("    (none)\n")
	} else {
		for _, d := range ext.FirewallDenied {
			fmt.Printf("    - %s\n", d)
		}
	}
}

func extensionReset(cfg *config.GlobalConfig, ext *config.ExtensionSettings, extName string) {
	ext.FirewallAllowed = nil
	ext.FirewallDenied = nil
	saveGlobalConfig(cfg)
	fmt.Printf("Cleared %s firewall rules\n", extName)
}
