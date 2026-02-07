//go:build addt

package addt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	configcmd "github.com/jedi4ever/addt/cmd/config"
)

// --- Config tests (in-process, no container needed) ---

func TestEnvFile_Addt_DefaultValues(t *testing.T) {
	// Scenario: User starts with no env file config. The defaults should be
	// env_file_load=true and env_file=.env, both with source=default.
	_, cleanup := setupAddtDir(t, "", ``)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")

	foundLoad := false
	foundFile := false
	for _, line := range lines {
		if strings.Contains(line, "env_file_load") {
			foundLoad = true
			if !strings.Contains(line, "true") {
				t.Errorf("Expected env_file_load default=true, got line: %s", line)
			}
			if !strings.Contains(line, "default") {
				t.Errorf("Expected env_file_load source=default, got line: %s", line)
			}
		}
		if strings.Contains(line, "env_file") && !strings.Contains(line, "env_file_load") {
			foundFile = true
			if !strings.Contains(line, ".env") {
				t.Errorf("Expected env_file default=.env, got line: %s", line)
			}
			if !strings.Contains(line, "default") {
				t.Errorf("Expected env_file source=default, got line: %s", line)
			}
		}
	}
	if !foundLoad {
		t.Errorf("Expected env_file_load key in config list, got:\n%s", output)
	}
	if !foundFile {
		t.Errorf("Expected env_file key in config list, got:\n%s", output)
	}
}

func TestEnvFile_Addt_ConfigLoaded(t *testing.T) {
	// Scenario: User sets env_file_load: true and env_file: custom.env in
	// .addt.yaml project config. Verify values and source=project.
	_, cleanup := setupAddtDir(t, "", `
env_file_load: true
env_file: custom.env
`)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.Contains(line, "env_file_load") {
			if !strings.Contains(line, "true") {
				t.Errorf("Expected env_file_load=true, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected env_file_load source=project, got line: %s", line)
			}
		}
		if strings.Contains(line, "env_file") && !strings.Contains(line, "env_file_load") {
			if !strings.Contains(line, "custom.env") {
				t.Errorf("Expected env_file=custom.env, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected env_file source=project, got line: %s", line)
			}
		}
	}
}

func TestEnvFile_Addt_ConfigViaSet(t *testing.T) {
	// Scenario: User sets env_file via 'config set' command, then verifies
	// it appears in config list with the correct value and source=project.
	_, cleanup := setupAddtDir(t, "", ``)
	defer cleanup()

	captureOutput(t, func() {
		configcmd.HandleCommand([]string{"set", "env_file", "custom.env"})
	})

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "env_file") && !strings.Contains(line, "env_file_load") {
			if !strings.Contains(line, "custom.env") {
				t.Errorf("Expected env_file=custom.env after config set, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected env_file source=project after config set, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected env_file key in config list, got:\n%s", output)
}

func TestEnvFile_Addt_DisabledViaConfig(t *testing.T) {
	// Scenario: User sets env_file_load: false in .addt.yaml to disable
	// env file loading. Verify value=false and source=project.
	_, cleanup := setupAddtDir(t, "", `
env_file_load: false
`)
	defer cleanup()

	output := captureOutput(t, func() {
		configcmd.HandleCommand([]string{"list"})
	})

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "env_file_load") {
			if !strings.Contains(line, "false") {
				t.Errorf("Expected env_file_load=false, got line: %s", line)
			}
			if !strings.Contains(line, "project") {
				t.Errorf("Expected env_file_load source=project, got line: %s", line)
			}
			return
		}
	}
	t.Errorf("Expected env_file_load key in config list, got:\n%s", output)
}

// --- Container tests (subprocess, both providers) ---

func TestEnvFile_Addt_VarsLoadedInContainer(t *testing.T) {
	// Scenario: User creates a .env file with TEST_VAR=from_envfile. After
	// running a command, the variable should be available inside the container.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
env_file_load: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// Set env vars for subprocess robustness
			origLoad := os.Getenv("ADDT_ENV_FILE_LOAD")
			os.Setenv("ADDT_ENV_FILE_LOAD", "true")
			defer func() {
				if origLoad != "" {
					os.Setenv("ADDT_ENV_FILE_LOAD", origLoad)
				} else {
					os.Unsetenv("ADDT_ENV_FILE_LOAD")
				}
			}()

			// Create .env file in project directory
			envContent := "TEST_VAR=from_envfile\n"
			if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0o644); err != nil {
				t.Fatalf("Failed to write .env: %v", err)
			}

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo ENVVAR:${TEST_VAR:-NOTSET}")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "ENVVAR:")
			if result != "from_envfile" {
				t.Errorf("Expected ENVVAR:from_envfile, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestEnvFile_Addt_MultipleVars(t *testing.T) {
	// Scenario: User creates a .env file with multiple variables. All of
	// them should be available inside the container.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
env_file_load: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			origLoad := os.Getenv("ADDT_ENV_FILE_LOAD")
			os.Setenv("ADDT_ENV_FILE_LOAD", "true")
			defer func() {
				if origLoad != "" {
					os.Setenv("ADDT_ENV_FILE_LOAD", origLoad)
				} else {
					os.Unsetenv("ADDT_ENV_FILE_LOAD")
				}
			}()

			// Create .env file with multiple variables
			envContent := "FIRST_VAR=hello\nSECOND_VAR=world\n"
			if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0o644); err != nil {
				t.Fatalf("Failed to write .env: %v", err)
			}

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo FIRST:${FIRST_VAR:-NOTSET} && echo SECOND:${SECOND_VAR:-NOTSET}")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			first := extractMarker(output, "FIRST:")
			if first != "hello" {
				t.Errorf("Expected FIRST:hello, got %q", first)
			}

			second := extractMarker(output, "SECOND:")
			if second != "world" {
				t.Errorf("Expected SECOND:world, got %q", second)
			}
		})
	}
}

func TestEnvFile_Addt_CustomFilePath(t *testing.T) {
	// Scenario: User points env_file to a custom path (custom.env) instead
	// of the default .env. Variables from that file should be loaded.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
env_file_load: true
env_file: custom.env
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// Set env vars for subprocess robustness
			origLoad := os.Getenv("ADDT_ENV_FILE_LOAD")
			origFile := os.Getenv("ADDT_ENV_FILE")
			os.Setenv("ADDT_ENV_FILE_LOAD", "true")
			os.Setenv("ADDT_ENV_FILE", "custom.env")
			defer func() {
				if origLoad != "" {
					os.Setenv("ADDT_ENV_FILE_LOAD", origLoad)
				} else {
					os.Unsetenv("ADDT_ENV_FILE_LOAD")
				}
				if origFile != "" {
					os.Setenv("ADDT_ENV_FILE", origFile)
				} else {
					os.Unsetenv("ADDT_ENV_FILE")
				}
			}()

			// Create custom.env file in project directory
			envContent := "CUSTOM_VAR=from_custom\n"
			if err := os.WriteFile(filepath.Join(dir, "custom.env"), []byte(envContent), 0o644); err != nil {
				t.Fatalf("Failed to write custom.env: %v", err)
			}

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo CUSTOMVAR:${CUSTOM_VAR:-NOTSET}")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "CUSTOMVAR:")
			if result != "from_custom" {
				t.Errorf("Expected CUSTOMVAR:from_custom, got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestEnvFile_Addt_DisabledNoVars(t *testing.T) {
	// Scenario: User sets env_file_load: false. Even with a .env file
	// present, variables should NOT be loaded into the container.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
env_file_load: false
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			// Set env vars for subprocess robustness
			origLoad := os.Getenv("ADDT_ENV_FILE_LOAD")
			os.Setenv("ADDT_ENV_FILE_LOAD", "false")
			defer func() {
				if origLoad != "" {
					os.Setenv("ADDT_ENV_FILE_LOAD", origLoad)
				} else {
					os.Unsetenv("ADDT_ENV_FILE_LOAD")
				}
			}()

			// Create .env file that should be ignored
			envContent := "DISABLED_VAR=should_not_appear\n"
			if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0o644); err != nil {
				t.Fatalf("Failed to write .env: %v", err)
			}

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo DISABLED:${DISABLED_VAR:-NOTSET}")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "DISABLED:")
			if result != "NOTSET" {
				t.Errorf("Expected DISABLED:NOTSET (env file disabled), got %q\nFull output:\n%s", result, output)
			}
		})
	}
}

func TestEnvFile_Addt_CommentsAndEmptyLines(t *testing.T) {
	// Scenario: User's .env file contains comments, empty lines, and quoted
	// values. Only valid variable assignments should be loaded.
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
env_file_load: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			origLoad := os.Getenv("ADDT_ENV_FILE_LOAD")
			os.Setenv("ADDT_ENV_FILE_LOAD", "true")
			defer func() {
				if origLoad != "" {
					os.Setenv("ADDT_ENV_FILE_LOAD", origLoad)
				} else {
					os.Unsetenv("ADDT_ENV_FILE_LOAD")
				}
			}()

			// Create .env file with comments, empty lines, and a plain value
			envContent := "# This is a comment\n\nPLAIN_VAR=plain_value\n"
			if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0o644); err != nil {
				t.Fatalf("Failed to write .env: %v", err)
			}

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo PLAIN:${PLAIN_VAR:-NOTSET}")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			plain := extractMarker(output, "PLAIN:")
			if plain != "plain_value" {
				t.Errorf("Expected PLAIN:plain_value, got %q", plain)
			}
		})
	}
}

func TestEnvFile_Addt_MissingFileNoError(t *testing.T) {
	// Scenario: No .env file exists in the project directory. The container
	// should start normally without errors (default behavior is silent skip).
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
env_file_load: true
`)
			defer cleanup()
			ensureAddtImage(t, dir, "debug")

			origLoad := os.Getenv("ADDT_ENV_FILE_LOAD")
			os.Setenv("ADDT_ENV_FILE_LOAD", "true")
			defer func() {
				if origLoad != "" {
					os.Setenv("ADDT_ENV_FILE_LOAD", origLoad)
				} else {
					os.Unsetenv("ADDT_ENV_FILE_LOAD")
				}
			}()

			// Deliberately do NOT create any .env file

			output, err := runRunSubcommand(t, dir, "debug",
				"-c", "echo MISSING:ok")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "MISSING:")
			if result != "ok" {
				t.Errorf("Expected MISSING:ok (container starts without .env), got %q\nFull output:\n%s", result, output)
			}
		})
	}
}
