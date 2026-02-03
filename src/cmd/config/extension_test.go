package config

import (
	"testing"

	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestExtensionSettings(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create config with extension settings
	automount := true
	cfg := &cfgtypes.GlobalConfig{
		Extensions: map[string]*cfgtypes.ExtensionSettings{
			"claude": {
				Version:   "1.0.5",
				Automount: &automount,
			},
		},
	}

	err := cfgtypes.SaveGlobalConfigFile(cfg)
	if err != nil {
		t.Fatalf("SaveGlobalConfigFile() error = %v", err)
	}

	// Load and verify
	loaded, err := cfgtypes.LoadGlobalConfigFile()
	if err != nil {
		t.Fatalf("LoadGlobalConfigFile() error = %v", err)
	}

	if loaded.Extensions == nil {
		t.Fatal("Extensions is nil")
	}

	claudeCfg := loaded.Extensions["claude"]
	if claudeCfg == nil {
		t.Fatal("claude extension config is nil")
	}

	if claudeCfg.Version != "1.0.5" {
		t.Errorf("claude.Version = %q, want %q", claudeCfg.Version, "1.0.5")
	}
	if claudeCfg.Automount == nil || *claudeCfg.Automount != true {
		t.Errorf("claude.Automount = %v, want true", claudeCfg.Automount)
	}
}

func TestExtensionSettingsInProjectConfig(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create project config with extension settings
	automount := false
	cfg := &cfgtypes.GlobalConfig{
		Extensions: map[string]*cfgtypes.ExtensionSettings{
			"claude": {
				Automount: &automount,
			},
		},
	}

	err := cfgtypes.SaveProjectConfigFile(cfg)
	if err != nil {
		t.Fatalf("SaveProjectConfigFile() error = %v", err)
	}

	// Load and verify
	loaded, err := cfgtypes.LoadProjectConfigFile()
	if err != nil {
		t.Fatalf("LoadProjectConfigFile() error = %v", err)
	}

	if loaded.Extensions == nil {
		t.Fatal("Extensions is nil")
	}

	claudeCfg := loaded.Extensions["claude"]
	if claudeCfg == nil {
		t.Fatal("claude extension config is nil")
	}

	if claudeCfg.Automount == nil || *claudeCfg.Automount != false {
		t.Errorf("claude.Automount = %v, want false", claudeCfg.Automount)
	}
}
