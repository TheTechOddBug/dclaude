package firewall

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedi4ever/addt/config"
)

func handleProject(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: addt firewall project <command>")
		fmt.Println("Commands: allow, deny, remove, list, reset")
		return
	}

	cmd := args[0]
	cfg := config.LoadProjectConfig()

	switch cmd {
	case "allow":
		projectAllow(cfg, args)
	case "deny":
		projectDeny(cfg, args)
	case "remove", "rm":
		projectRemove(cfg, args)
	case "list", "ls":
		projectList(cfg)
	case "reset":
		projectReset(cfg)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Commands: allow, deny, remove, list, reset")
	}
}

func projectAllow(cfg *config.GlobalConfig, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: addt firewall project allow <domain>")
		os.Exit(1)
	}
	domain := strings.TrimSpace(args[1])
	if containsString(cfg.FirewallAllowed, domain) {
		fmt.Printf("Domain '%s' already in project allowed list\n", domain)
		return
	}
	cfg.FirewallAllowed = append(cfg.FirewallAllowed, domain)
	saveProjectConfig(cfg)
	fmt.Printf("Added '%s' to project allowed domains\n", domain)
}

func projectDeny(cfg *config.GlobalConfig, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: addt firewall project deny <domain>")
		os.Exit(1)
	}
	domain := strings.TrimSpace(args[1])
	if containsString(cfg.FirewallDenied, domain) {
		fmt.Printf("Domain '%s' already in project denied list\n", domain)
		return
	}
	cfg.FirewallDenied = append(cfg.FirewallDenied, domain)
	saveProjectConfig(cfg)
	fmt.Printf("Added '%s' to project denied domains\n", domain)
}

func projectRemove(cfg *config.GlobalConfig, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: addt firewall project remove <domain>")
		os.Exit(1)
	}
	domain := strings.TrimSpace(args[1])
	removed := removeDomainFromConfig(cfg, domain)
	if removed {
		saveProjectConfig(cfg)
	} else {
		fmt.Printf("Domain '%s' not found in project config\n", domain)
	}
}

func projectList(cfg *config.GlobalConfig) {
	fmt.Println("Project firewall rules:")
	printDomainList("  Allowed", cfg.FirewallAllowed, nil, cfg.FirewallDenied)
	fmt.Printf("  Denied:\n")
	if len(cfg.FirewallDenied) == 0 {
		fmt.Printf("    (none)\n")
	} else {
		for _, d := range cfg.FirewallDenied {
			fmt.Printf("    - %s\n", d)
		}
	}
}

func projectReset(cfg *config.GlobalConfig) {
	cfg.FirewallAllowed = nil
	cfg.FirewallDenied = nil
	saveProjectConfig(cfg)
	fmt.Println("Cleared project firewall rules")
}
