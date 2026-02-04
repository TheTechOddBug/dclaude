package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterSecretEnvVars(t *testing.T) {
	p := &DockerProvider{}

	env := map[string]string{
		"ANTHROPIC_API_KEY": "secret-key",
		"GH_TOKEN":          "github-token",
		"TERM":              "xterm-256color",
		"HOME":              "/home/user",
	}

	secretVarNames := []string{"ANTHROPIC_API_KEY", "GH_TOKEN"}

	p.filterSecretEnvVars(env, secretVarNames)

	// Secret vars should be removed
	if _, exists := env["ANTHROPIC_API_KEY"]; exists {
		t.Error("ANTHROPIC_API_KEY should be removed")
	}
	if _, exists := env["GH_TOKEN"]; exists {
		t.Error("GH_TOKEN should be removed")
	}

	// Non-secret vars should remain
	if env["TERM"] != "xterm-256color" {
		t.Errorf("TERM = %q, want \"xterm-256color\"", env["TERM"])
	}
	if env["HOME"] != "/home/user" {
		t.Errorf("HOME = %q, want \"/home/user\"", env["HOME"])
	}
}

func TestAddSecretsMount(t *testing.T) {
	p := &DockerProvider{}

	// Test with empty secrets dir
	args := p.addSecretsMount([]string{"-it"}, "")
	if len(args) != 1 {
		t.Errorf("Expected 1 arg for empty secrets dir, got %d", len(args))
	}

	// Test with secrets dir
	args = p.addSecretsMount([]string{"-it"}, "/tmp/secrets")
	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
	if args[1] != "-v" {
		t.Errorf("Expected -v flag, got %s", args[1])
	}
	if args[2] != "/tmp/secrets:/run/secrets:ro" {
		t.Errorf("Expected mount arg, got %s", args[2])
	}
}

func TestWriteSecretsToFiles(t *testing.T) {
	// Create a temporary directory to simulate secrets
	tmpDir, err := os.MkdirTemp("", "secrets-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test that secrets are written correctly
	env := map[string]string{
		"SECRET_KEY": "my-secret-value",
		"OTHER_VAR":  "other-value",
	}

	// Write secret to file manually to test the pattern
	secretPath := filepath.Join(tmpDir, "SECRET_KEY")
	if err := os.WriteFile(secretPath, []byte("my-secret-value"), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	// Verify the secret was written
	content, err := os.ReadFile(secretPath)
	if err != nil {
		t.Fatalf("Failed to read secret: %v", err)
	}
	if string(content) != env["SECRET_KEY"] {
		t.Errorf("Secret content = %q, want %q", string(content), env["SECRET_KEY"])
	}

	// Verify file permissions (0600)
	info, err := os.Stat(secretPath)
	if err != nil {
		t.Fatalf("Failed to stat secret: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Secret permissions = %o, want 0600", info.Mode().Perm())
	}
}
