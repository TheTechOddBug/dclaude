package extensions

import (
	"testing"
)

func TestIsBuiltinExtension(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"claude", true},
		{"codex", true},
		{"gemini", true},
		{"nonexistent", false},
		{"", false},
		{"CLAUDE", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBuiltinExtension(tt.name)
			if result != tt.expected {
				t.Errorf("isBuiltinExtension(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}
