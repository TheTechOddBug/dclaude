//go:build addt

package addt

import (
	"os"
	"path/filepath"
	"testing"
)

// setupCredentialScript creates a credential script for the debug-creds extension
// under ADDT_HOME/extensions/debug-creds/credentials.sh. Returns ADDT_HOME path
// and a cleanup function that restores the original ADDT_HOME.
func setupCredentialScript(t *testing.T) (string, func()) {
	t.Helper()

	addtHome := t.TempDir()
	credDir := filepath.Join(addtHome, "extensions", "debug-creds")
	if err := os.MkdirAll(credDir, 0755); err != nil {
		t.Fatalf("Failed to create credential script dir: %v", err)
	}

	// Credential script outputs a secret env var on the HOST before container start.
	// The entrypoint should scrub and unset it after extension setup.
	credScript := "#!/bin/bash\necho TEST_CRED_SECRET=secret-value-12345\n"
	if err := os.WriteFile(filepath.Join(credDir, "credentials.sh"), []byte(credScript), 0700); err != nil {
		t.Fatalf("Failed to write credentials.sh: %v", err)
	}

	origHome := os.Getenv("ADDT_HOME")
	os.Setenv("ADDT_HOME", addtHome)

	cleanup := func() {
		if origHome != "" {
			os.Setenv("ADDT_HOME", origHome)
		} else {
			os.Unsetenv("ADDT_HOME")
		}
	}

	return addtHome, cleanup
}

// --- Container tests (subprocess, both providers) ---

// Scenario: A developer's extension has a credential script that provides
// a secret API key. After the container starts and extension setup completes,
// the credential env var should be scrubbed and unset — it must NOT leak
// into the user's shell session or the running process environment.
func TestScrub_Addt_CredentialVarsNotLeaked(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			// Use isolate_secrets: false so credential vars are passed as -e flags
			// and scrubbed by the entrypoint's ADDT_CREDENTIAL_VARS block
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
security:
  isolate_secrets: false
`)
			defer cleanup()

			_, credCleanup := setupCredentialScript(t)
			defer credCleanup()

			ensureAddtImage(t, dir, "debug-creds")

			// Verify both the credential var and the meta-var are gone
			output, err := runRunSubcommand(t, dir, "debug-creds",
				"-c", "echo CRED:${TEST_CRED_SECRET:-NOTSET} && echo META:${ADDT_CREDENTIAL_VARS:-NOTSET}")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			cred := extractMarker(output, "CRED:")
			if cred != "NOTSET" {
				t.Errorf("Expected TEST_CRED_SECRET to be scrubbed and unset (NOTSET), got %q", cred)
			}

			meta := extractMarker(output, "META:")
			if meta != "NOTSET" {
				t.Errorf("Expected ADDT_CREDENTIAL_VARS to be scrubbed and unset (NOTSET), got %q", meta)
			}
		})
	}
}

// Scenario: With secrets isolation enabled (the default), credentials are
// delivered through a tmpfs file at /run/secrets/.secrets. After the
// entrypoint loads them, the file should be overwritten with random data
// and deleted — preventing recovery from disk or tmpfs.
func TestScrub_Addt_SecretsFileRemoved(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
security:
  isolate_secrets: true
`)
			defer cleanup()

			_, credCleanup := setupCredentialScript(t)
			defer credCleanup()

			ensureAddtImage(t, dir, "debug-creds")

			// Verify the secrets file does not exist after the entrypoint ran
			output, err := runRunSubcommand(t, dir, "debug-creds",
				"-c", "if [ -f /run/secrets/.secrets ]; then echo SECRETS_FILE:EXISTS; else echo SECRETS_FILE:GONE; fi")
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "SECRETS_FILE:")
			if result != "GONE" {
				t.Errorf("Expected secrets file to be scrubbed and removed (GONE), got %q", result)
			}
		})
	}
}

// Scenario: With secrets isolation enabled, a credential var that was
// delivered through the secrets file mechanism should NOT appear in
// /proc/1/environ — meaning it was never passed as a container -e flag.
func TestScrub_Addt_CredentialNotInProcEnviron(t *testing.T) {
	providers := requireProviders(t)

	for _, prov := range providers {
		t.Run(prov, func(t *testing.T) {
			dir, cleanup := setupAddtDirWithExtensions(t, prov, `
security:
  isolate_secrets: true
`)
			defer cleanup()

			_, credCleanup := setupCredentialScript(t)
			defer credCleanup()

			ensureAddtImage(t, dir, "debug-creds")

			// Check /proc/1/environ for the credential value —
			// it should NOT be there when delivered via secrets file
			output, err := runRunSubcommand(t, dir, "debug-creds",
				"-c", procEnvLeakCommand("TEST_CRED_SECRET"))
			t.Logf("Output:\n%s", output)
			if err != nil {
				t.Fatalf("run failed: %v\nOutput:\n%s", err, output)
			}

			result := extractMarker(output, "PROC_RESULT:")
			if result != "ISOLATED" {
				t.Errorf("Expected TEST_CRED_SECRET to be isolated from /proc/1/environ, got PROC_RESULT:%s", result)
			}
		})
	}
}
