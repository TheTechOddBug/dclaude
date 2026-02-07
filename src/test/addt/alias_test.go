//go:build addt

package addt

import (
	"strings"
	"testing"
)

// --- Alias (symlink) invocation tests ---
// These tests verify that binary name detection works correctly when addt
// is invoked via a symlink like "addt-codex". The detection logic in
// cmd/root.go sets ADDT_EXTENSIONS and ADDT_COMMAND based on the binary name.

func TestAlias_Addt_HelpShowsUsage(t *testing.T) {
	// Scenario: A user runs `addt-codex --help` to see available options.
	// When invoked via alias, --help is passed through to the container
	// (the extension shows its own help). This requires a built image.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()
			ensureAddtImage(t, dir, "codex")

			output, err := runAliasCommand(t, dir, "codex", "--help")
			t.Logf("Output:\n%s", output)

			// The command may exit non-zero (codex --help may return 0 or 1),
			// but the output should contain help/usage text from codex
			if err != nil {
				// Log but don't fail on exit code — codex --help may use non-zero
				t.Logf("exit error (may be expected): %v", err)
			}

			outputLower := strings.ToLower(output)
			if !strings.Contains(outputLower, "codex") && !strings.Contains(outputLower, "usage") && !strings.Contains(outputLower, "help") {
				t.Errorf("Expected help output to contain 'codex', 'usage', or 'help', got:\n%s", output)
			}
		})
	}
}

func TestAlias_Addt_SubcommandVersion(t *testing.T) {
	// Scenario: A user runs `addt-codex addt version` to check the addt version.
	// The "addt" subcommand namespace is handled at the host level (root.go:116-140),
	// so this does NOT need a container.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()

			output, err := runAliasCommand(t, dir, "codex", "addt", "version")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("alias command failed: %v\nOutput:\n%s", err, output)
			}

			if !strings.Contains(output, testVersion) {
				t.Errorf("Expected output to contain version %q, got:\n%s", testVersion, output)
			}
		})
	}
}

func TestAlias_Addt_SubcommandUsage(t *testing.T) {
	// Scenario: A user runs `addt-codex addt` (no subcommand) to see
	// available addt subcommands. This is handled at the host level
	// by printAddtSubcommandUsage() in cmd/cli.go, so no container needed.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDir(t, prov, ``)
			defer cleanup()

			output, err := runAliasCommand(t, dir, "codex", "addt")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("alias command failed: %v\nOutput:\n%s", err, output)
			}

			// Verify key subcommands are listed in the usage output
			for _, keyword := range []string{"build", "shell", "config", "version"} {
				if !strings.Contains(output, keyword) {
					t.Errorf("Expected usage output to contain %q, got:\n%s", keyword, output)
				}
			}
		})
	}
}
