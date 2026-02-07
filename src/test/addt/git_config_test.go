//go:build addt

package addt

import (
	"strings"
	"testing"

	"github.com/jedi4ever/addt/cmd"
	configcmd "github.com/jedi4ever/addt/cmd/config"
)

// --- Git config tests (in-process, no container needed) ---

func TestGitConfig_Addt_DefaultEnabled(t *testing.T) {
	// Scenario: User starts with no git config and checks defaults.
	// git.forward_config should default to true.
	_, cleanup := setupAddtDir(t, "", ``)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "git.forward_config") {
			if !strings.Contains(line, "true") {
				t.Errorf("Expected git.forward_config default=true, got line: %s", line)
			}
			if !strings.Contains(line, "default") {
				t.Errorf("Expected git.forward_config source=default, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected output to contain git.forward_config, got:\n%s", output)
}

func TestGitConfig_Addt_ConfigViaSet(t *testing.T) {
	// Scenario: User disables git config forwarding via 'config set git.forward_config false',
	// then verifies it appears as false in config list.
	_, cleanup := setupAddtDir(t, "", ``)
	defer cleanup()

	captureOutput(t, func() {
		configcmd.HandleCommand([]string{"set", "git.forward_config", "false"})
	})

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "git.forward_config") {
			if !strings.Contains(line, "false") {
				t.Errorf("Expected git.forward_config=false after config set, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected git.forward_config source=project after config set, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected output to contain git.forward_config, got:\n%s", output)
}

func TestGitConfig_Addt_ConfigPathSet(t *testing.T) {
	// Scenario: User sets a custom .gitconfig path via 'config set git.config_path /tmp/.gitconfig',
	// then verifies it appears in config list.
	_, cleanup := setupAddtDir(t, "", ``)
	defer cleanup()

	captureOutput(t, func() {
		configcmd.HandleCommand([]string{"set", "git.config_path", "/tmp/.gitconfig"})
	})

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "git.config_path") {
			if !strings.Contains(line, "/tmp/.gitconfig") {
				t.Errorf("Expected git.config_path=/tmp/.gitconfig, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected git.config_path source=project, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected output to contain git.config_path, got:\n%s", output)
}

func TestGitConfig_Addt_CompletionIncluded(t *testing.T) {
	// Scenario: User generates bash completion and verifies that git config keys
	// are dynamically injected into the completion script.
	_, cleanup := setupAddtDir(t, "", "")
	defer cleanup()

	output := captureOutput(t, func() {
		cmd.HandleCompletionCommand([]string{"bash"})
	})

	for _, key := range []string{"git.forward_config", "git.config_path"} {
		if !strings.Contains(output, key) {
			t.Errorf("Expected bash completion to include config key %q, got:\n%s", key, output)
		}
	}
}
