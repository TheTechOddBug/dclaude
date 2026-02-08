package profile

import (
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed presets/*.yaml
var presetsFS embed.FS

// Profile represents a named configuration preset
type Profile struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Settings    map[string]string `yaml:"settings"`
}

// GetProfiles returns all available profiles sorted by name
func GetProfiles() ([]Profile, error) {
	entries, err := presetsFS.ReadDir("presets")
	if err != nil {
		return nil, fmt.Errorf("reading presets directory: %w", err)
	}

	var profiles []Profile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := presetsFS.ReadFile(filepath.Join("presets", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading preset %s: %w", entry.Name(), err)
		}
		var p Profile
		if err := yaml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("parsing preset %s: %w", entry.Name(), err)
		}
		profiles = append(profiles, p)
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	return profiles, nil
}

// GetProfile returns a profile by name, or nil if not found
func GetProfile(name string) (*Profile, error) {
	profiles, err := GetProfiles()
	if err != nil {
		return nil, err
	}
	for _, p := range profiles {
		if p.Name == name {
			return &p, nil
		}
	}
	return nil, nil
}

// GetPresetsFS returns the embedded presets filesystem for use in asset hashing
func GetPresetsFS() embed.FS {
	return presetsFS
}

// GetProfileNames returns the names of all available profiles
func GetProfileNames() []string {
	profiles, err := GetProfiles()
	if err != nil {
		return nil
	}
	var names []string
	for _, p := range profiles {
		names = append(names, p.Name)
	}
	return names
}
