package profile

import (
	"testing"

	cfgcmd "github.com/jedi4ever/addt/cmd/config"
	cfgtypes "github.com/jedi4ever/addt/config"
)

func TestGetProfilesReturnsThree(t *testing.T) {
	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles() error: %v", err)
	}
	if len(profiles) != 3 {
		t.Errorf("GetProfiles() returned %d profiles, want 3", len(profiles))
	}
}

func TestGetProfilesSortedByName(t *testing.T) {
	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles() error: %v", err)
	}
	expected := []string{"develop", "paranoia", "strict"}
	for i, p := range profiles {
		if p.Name != expected[i] {
			t.Errorf("profiles[%d].Name = %q, want %q", i, p.Name, expected[i])
		}
	}
}

func TestGetProfileByName(t *testing.T) {
	tests := []struct {
		name        string
		expectFound bool
	}{
		{"develop", true},
		{"strict", true},
		{"paranoia", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		p, err := GetProfile(tt.name)
		if err != nil {
			t.Fatalf("GetProfile(%q) error: %v", tt.name, err)
		}
		if tt.expectFound && p == nil {
			t.Errorf("GetProfile(%q) = nil, want profile", tt.name)
		}
		if !tt.expectFound && p != nil {
			t.Errorf("GetProfile(%q) = %v, want nil", tt.name, p)
		}
	}
}

func TestProfilesHaveDescriptions(t *testing.T) {
	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles() error: %v", err)
	}
	for _, p := range profiles {
		if p.Description == "" {
			t.Errorf("profile %q has empty description", p.Name)
		}
	}
}

func TestProfilesHaveSettings(t *testing.T) {
	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles() error: %v", err)
	}
	for _, p := range profiles {
		if len(p.Settings) == 0 {
			t.Errorf("profile %q has no settings", p.Name)
		}
	}
}

func TestAllProfileKeysAreValidConfigKeys(t *testing.T) {
	profiles, err := GetProfiles()
	if err != nil {
		t.Fatalf("GetProfiles() error: %v", err)
	}
	for _, p := range profiles {
		for key := range p.Settings {
			if !cfgcmd.IsValidKey(key) {
				t.Errorf("profile %q has invalid config key: %q", p.Name, key)
			}
		}
	}
}

func TestApplyProfileSetsValues(t *testing.T) {
	p, err := GetProfile("strict")
	if err != nil {
		t.Fatalf("GetProfile(strict) error: %v", err)
	}
	if p == nil {
		t.Fatal("GetProfile(strict) returned nil")
	}

	cfg := &cfgtypes.GlobalConfig{}
	for k, v := range p.Settings {
		cfgcmd.SetValue(cfg, k, v)
	}

	// Verify a few key values were applied
	got := cfgcmd.GetValue(cfg, "firewall.enabled")
	if got != "true" {
		t.Errorf("after apply strict, firewall.enabled = %q, want %q", got, "true")
	}

	got = cfgcmd.GetValue(cfg, "security.audit_log")
	if got != "true" {
		t.Errorf("after apply strict, security.audit_log = %q, want %q", got, "true")
	}

	got = cfgcmd.GetValue(cfg, "container.memory")
	if got != "2g" {
		t.Errorf("after apply strict, container.memory = %q, want %q", got, "2g")
	}
}

func TestDevelopProfileMatchesDefaults(t *testing.T) {
	p, err := GetProfile("develop")
	if err != nil {
		t.Fatalf("GetProfile(develop) error: %v", err)
	}
	if p == nil {
		t.Fatal("GetProfile(develop) returned nil")
	}

	for key, val := range p.Settings {
		def := cfgcmd.GetDefaultValue(key)
		if val != def {
			t.Errorf("develop profile key %q = %q, differs from default %q", key, val, def)
		}
	}
}

func TestGetProfileNames(t *testing.T) {
	names := GetProfileNames()
	if len(names) != 3 {
		t.Errorf("GetProfileNames() returned %d names, want 3", len(names))
	}
	expected := map[string]bool{"develop": true, "strict": true, "paranoia": true}
	for _, n := range names {
		if !expected[n] {
			t.Errorf("unexpected profile name: %q", n)
		}
	}
}

func TestGetPresetsFS(t *testing.T) {
	fsys := GetPresetsFS()
	entries, err := fsys.ReadDir("presets")
	if err != nil {
		t.Fatalf("GetPresetsFS().ReadDir(presets) error: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("presets directory has %d entries, want 3", len(entries))
	}
}
