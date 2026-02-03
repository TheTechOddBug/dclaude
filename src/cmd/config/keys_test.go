package config

import (
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestKeyValidation(t *testing.T) {
	validKeys := []string{
		"docker_cpus", "docker_memory", "dind", "dind_mode",
		"firewall", "firewall_mode", "node_version", "go_version",
		"persistent", "workdir", "workdir_automount",
	}

	for _, key := range validKeys {
		if !IsValidKey(key) {
			t.Errorf("IsValidKey(%q) = false, want true", key)
		}
	}

	invalidKeys := []string{"invalid", "foo", "bar", "version"}
	for _, key := range invalidKeys {
		if IsValidKey(key) {
			t.Errorf("IsValidKey(%q) = true, want false", key)
		}
	}
}

func TestExtensionKeyValidation(t *testing.T) {
	validKeys := []string{"version", "automount"}
	for _, key := range validKeys {
		if !IsValidExtensionKey(key) {
			t.Errorf("IsValidExtensionKey(%q) = false, want true", key)
		}
	}

	invalidKeys := []string{"invalid", "foo", "node_version"}
	for _, key := range invalidKeys {
		if IsValidExtensionKey(key) {
			t.Errorf("IsValidExtensionKey(%q) = true, want false", key)
		}
	}
}

func TestGetValue(t *testing.T) {
	persistent := true
	portStart := 35000
	cfg := &cfgtypes.GlobalConfig{
		NodeVersion:    "20",
		DockerCPUs:     "4",
		Persistent:     &persistent,
		PortRangeStart: &portStart,
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"node_version", "20"},
		{"docker_cpus", "4"},
		{"persistent", "true"},
		{"port_range_start", "35000"},
		{"go_version", ""}, // not set
	}

	for _, tt := range tests {
		got := GetValue(cfg, tt.key)
		if got != tt.expected {
			t.Errorf("GetValue(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}

func TestSetValue(t *testing.T) {
	cfg := &cfgtypes.GlobalConfig{}

	SetValue(cfg, "node_version", "18")
	if cfg.NodeVersion != "18" {
		t.Errorf("NodeVersion = %q, want %q", cfg.NodeVersion, "18")
	}

	SetValue(cfg, "persistent", "true")
	if cfg.Persistent == nil || *cfg.Persistent != true {
		t.Errorf("Persistent = %v, want true", cfg.Persistent)
	}

	SetValue(cfg, "port_range_start", "40000")
	if cfg.PortRangeStart == nil || *cfg.PortRangeStart != 40000 {
		t.Errorf("PortRangeStart = %v, want 40000", cfg.PortRangeStart)
	}
}

func TestUnsetValue(t *testing.T) {
	persistent := true
	cfg := &cfgtypes.GlobalConfig{
		NodeVersion: "20",
		Persistent:  &persistent,
	}

	UnsetValue(cfg, "node_version")
	if cfg.NodeVersion != "" {
		t.Errorf("NodeVersion = %q, want empty", cfg.NodeVersion)
	}

	UnsetValue(cfg, "persistent")
	if cfg.Persistent != nil {
		t.Errorf("Persistent = %v, want nil", cfg.Persistent)
	}
}

func TestGetDefaultValue(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"node_version", "22"},
		{"firewall", "false"},
		{"firewall_mode", "strict"},
		{"persistent", "false"},
		{"workdir_automount", "true"},
		{"port_range_start", "30000"},
		{"ssh_forward", "agent"},
	}

	for _, tt := range tests {
		got := GetDefaultValue(tt.key)
		if got != tt.expected {
			t.Errorf("GetDefaultValue(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}
}
