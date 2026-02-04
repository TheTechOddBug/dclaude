package config

import (
	"os"
	"testing"
)

func TestLoadConfig_SecurityDefaults(t *testing.T) {
	// Setup isolated test environment
	globalDir := t.TempDir()
	projectDir := t.TempDir()
	origConfigDir := os.Getenv("ADDT_CONFIG_DIR")
	origCwd, _ := os.Getwd()

	os.Setenv("ADDT_CONFIG_DIR", globalDir)
	os.Chdir(projectDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("ADDT_CONFIG_DIR", origConfigDir)
		} else {
			os.Unsetenv("ADDT_CONFIG_DIR")
		}
		os.Chdir(origCwd)
	}()

	// Clear any security env vars
	os.Unsetenv("ADDT_SECURITY_PIDS_LIMIT")
	os.Unsetenv("ADDT_SECURITY_NO_NEW_PRIVILEGES")

	cfg := LoadConfig("0.0.0-test", "20", "1.21", "0.1.0", 30000)

	// Check defaults
	if cfg.Security.PidsLimit != 200 {
		t.Errorf("PidsLimit = %d, want 200 (default)", cfg.Security.PidsLimit)
	}
	if cfg.Security.UlimitNofile != "4096:8192" {
		t.Errorf("UlimitNofile = %q, want \"4096:8192\" (default)", cfg.Security.UlimitNofile)
	}
	if cfg.Security.UlimitNproc != "256:512" {
		t.Errorf("UlimitNproc = %q, want \"256:512\" (default)", cfg.Security.UlimitNproc)
	}
	if !cfg.Security.NoNewPrivileges {
		t.Error("NoNewPrivileges = false, want true (default)")
	}
	if len(cfg.Security.CapDrop) != 1 || cfg.Security.CapDrop[0] != "ALL" {
		t.Errorf("CapDrop = %v, want [ALL] (default)", cfg.Security.CapDrop)
	}
	expectedCapAdd := []string{"CHOWN", "SETUID", "SETGID"}
	if len(cfg.Security.CapAdd) != 3 {
		t.Errorf("CapAdd = %v, want %v (default)", cfg.Security.CapAdd, expectedCapAdd)
	}
	if cfg.Security.ReadOnlyRootfs {
		t.Error("ReadOnlyRootfs = true, want false (default)")
	}
}

func TestLoadConfig_SecurityEnvOverrides(t *testing.T) {
	// Setup isolated test environment
	globalDir := t.TempDir()
	projectDir := t.TempDir()
	origConfigDir := os.Getenv("ADDT_CONFIG_DIR")
	origCwd, _ := os.Getwd()

	os.Setenv("ADDT_CONFIG_DIR", globalDir)
	os.Chdir(projectDir)
	defer func() {
		if origConfigDir != "" {
			os.Setenv("ADDT_CONFIG_DIR", origConfigDir)
		} else {
			os.Unsetenv("ADDT_CONFIG_DIR")
		}
		os.Chdir(origCwd)
		os.Unsetenv("ADDT_SECURITY_PIDS_LIMIT")
		os.Unsetenv("ADDT_SECURITY_NO_NEW_PRIVILEGES")
		os.Unsetenv("ADDT_SECURITY_CAP_DROP")
		os.Unsetenv("ADDT_SECURITY_CAP_ADD")
	}()

	// Set env overrides
	os.Setenv("ADDT_SECURITY_PIDS_LIMIT", "500")
	os.Setenv("ADDT_SECURITY_NO_NEW_PRIVILEGES", "false")
	os.Setenv("ADDT_SECURITY_CAP_DROP", "NET_RAW,SYS_ADMIN")
	os.Setenv("ADDT_SECURITY_CAP_ADD", "MKNOD")

	cfg := LoadConfig("0.0.0-test", "20", "1.21", "0.1.0", 30000)

	if cfg.Security.PidsLimit != 500 {
		t.Errorf("PidsLimit = %d, want 500 (from env)", cfg.Security.PidsLimit)
	}
	if cfg.Security.NoNewPrivileges {
		t.Error("NoNewPrivileges = true, want false (from env)")
	}
	expectedCapDrop := []string{"NET_RAW", "SYS_ADMIN"}
	if len(cfg.Security.CapDrop) != 2 || cfg.Security.CapDrop[0] != "NET_RAW" || cfg.Security.CapDrop[1] != "SYS_ADMIN" {
		t.Errorf("CapDrop = %v, want %v (from env)", cfg.Security.CapDrop, expectedCapDrop)
	}
	if len(cfg.Security.CapAdd) != 1 || cfg.Security.CapAdd[0] != "MKNOD" {
		t.Errorf("CapAdd = %v, want [MKNOD] (from env)", cfg.Security.CapAdd)
	}
}
