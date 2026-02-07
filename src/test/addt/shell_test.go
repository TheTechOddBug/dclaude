//go:build addt

package addt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Helper to extract shell test markers from subprocess output ---

func extractShellResult(output, marker string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, marker) {
			return strings.TrimPrefix(line, marker)
		}
	}
	return ""
}

// --- Container tests (subprocess, both providers) ---

func TestShell_Addt_BasicExecution(t *testing.T) {
	// Scenario: User runs `addt shell claude -c "echo SHELL_TEST:hello"`.
	// The shell subcommand should execute the command inside the container
	// and output the echoed value, confirming the shell path works end-to-end.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellSubcommand(t, dir, "claude",
				"-c", "echo SHELL_TEST:hello")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("shell subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractShellResult(output, "SHELL_TEST:")
			if result != "hello" {
				t.Errorf("Expected SHELL_TEST:hello, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestShell_Addt_WorkdirMounted(t *testing.T) {
	// Scenario: User creates a marker file in the project directory, then
	// runs `addt shell claude` to check the file exists at /workspace/.
	// This confirms workdir mounting works in shell mode.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			// Create a marker file in the project directory
			markerFile := filepath.Join(dir, "shell_test_marker.txt")
			if err := os.WriteFile(markerFile, []byte("MARKER_FOUND"), 0o644); err != nil {
				t.Fatalf("Failed to write marker file: %v", err)
			}

			output, err := runShellSubcommand(t, dir, "claude",
				"-c", "cat /workspace/shell_test_marker.txt && echo WORKDIR_OK:yes || echo WORKDIR_OK:no")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("shell subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractShellResult(output, "WORKDIR_OK:")
			if result != "yes" {
				t.Errorf("Expected WORKDIR_OK:yes, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestShell_Addt_BashIsDefault(t *testing.T) {
	// Scenario: User runs `addt shell claude` and checks that
	// ADDT_COMMAND is set to /bin/bash, confirming the shell entrypoint override.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellSubcommand(t, dir, "claude",
				"-c", "echo SHELL_CMD:$ADDT_COMMAND")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("shell subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractShellResult(output, "SHELL_CMD:")
			if result != "/bin/bash" {
				t.Errorf("Expected ADDT_COMMAND=/bin/bash, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestShell_Addt_EnvVarsForwarded(t *testing.T) {
	// Scenario: User configures custom env vars in project config, then opens
	// a shell. The env vars should be available inside the container,
	// confirming env forwarding works through the shell subcommand path.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, `
env:
  - "SHELL_TEST_VAR=myvalue"
`)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellSubcommand(t, dir, "claude",
				"-c", "echo ENVVAR:${SHELL_TEST_VAR:-NOTSET}")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("shell subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractShellResult(output, "ENVVAR:")
			if result != "myvalue" {
				t.Errorf("Expected SHELL_TEST_VAR=myvalue, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestShell_Addt_UserIsAddt(t *testing.T) {
	// Scenario: User runs `whoami` inside the shell container and expects
	// the output to be "addt" (not root), confirming user mapping works.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "claude")

			output, err := runShellSubcommand(t, dir, "claude",
				"-c", "echo WHOAMI:$(whoami)")

			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("shell subcommand failed: %v\nOutput:\n%s", err, output)
			}

			result := extractShellResult(output, "WHOAMI:")
			if result != "addt" {
				t.Errorf("Expected whoami=addt, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}
