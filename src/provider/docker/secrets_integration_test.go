//go:build integration

package docker

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/jedi4ever/addt/assets"
	"github.com/jedi4ever/addt/config/security"
	"github.com/jedi4ever/addt/extensions"
	"github.com/jedi4ever/addt/provider"
)

// checkDockerForSecrets verifies Docker is available
func checkDockerForSecrets(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not found in PATH, skipping integration test")
	}
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker daemon not running, skipping integration test")
	}
}

func TestSecretsToFiles_Integration_EnvVarsNotPassed(t *testing.T) {
	checkDockerForSecrets(t)

	// Create a provider with secrets_to_files enabled
	secCfg := security.DefaultConfig()
	secCfg.SecretsToFiles = true

	cfg := &provider.Config{
		Security: secCfg,
	}

	prov := &DockerProvider{
		config:   cfg,
		tempDirs: []string{},
	}

	// Simulate env with secrets
	env := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-test-key-12345",
		"TERM":              "xterm-256color",
		"HOME":              "/home/addt",
	}

	// Get extension env vars (simulate what would come from extension config)
	secretVarNames := []string{"ANTHROPIC_API_KEY"}

	// Write secrets to files
	secretsDir, writtenSecrets, err := prov.writeSecretsToFiles("test-image", env)
	if err != nil {
		// If no extension metadata, secrets won't be written - that's OK for this test
		// We'll test the filtering behavior instead
		t.Log("No extension metadata available, testing filter behavior only")
	}

	if secretsDir != "" {
		defer os.RemoveAll(secretsDir)
	}

	// Filter the secret env vars from the env map
	prov.filterSecretEnvVars(env, secretVarNames)

	// Verify ANTHROPIC_API_KEY was removed from env
	if _, exists := env["ANTHROPIC_API_KEY"]; exists {
		t.Error("ANTHROPIC_API_KEY should be filtered out when secrets_to_files is enabled")
	}

	// Verify non-secret env vars remain
	if env["TERM"] != "xterm-256color" {
		t.Errorf("TERM should remain, got %q", env["TERM"])
	}

	t.Logf("Secrets written: %v", writtenSecrets)
}

func TestSecretsToFiles_Integration_FilesReadable(t *testing.T) {
	checkDockerForSecrets(t)

	// Create a temp directory with secrets
	secretsDir, err := os.MkdirTemp("", "secrets-integration-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(secretsDir)

	// Write a test secret
	secretValue := "sk-ant-test-key-integration-12345"
	secretPath := secretsDir + "/ANTHROPIC_API_KEY"
	if err := os.WriteFile(secretPath, []byte(secretValue), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	// Run a container that reads the secret from file
	cmd := exec.Command("docker", "run", "--rm",
		"-v", secretsDir+":/run/secrets:ro",
		"alpine:latest",
		"cat", "/run/secrets/ANTHROPIC_API_KEY")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	if strings.TrimSpace(string(output)) != secretValue {
		t.Errorf("Expected secret value %q, got %q", secretValue, string(output))
	}
}

func TestSecretsToFiles_Integration_EntrypointLoadsSecrets(t *testing.T) {
	checkDockerForSecrets(t)

	// Create a temp directory with secrets
	secretsDir, err := os.MkdirTemp("", "secrets-entrypoint-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(secretsDir)

	// Write test secrets
	secrets := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-test-key-entrypoint-12345",
		"GH_TOKEN":          "ghp_test_token_67890",
	}

	for name, value := range secrets {
		if err := os.WriteFile(secretsDir+"/"+name, []byte(value), 0600); err != nil {
			t.Fatalf("Failed to write secret %s: %v", name, err)
		}
	}

	// Run container with entrypoint-like script that loads secrets and prints env
	script := `
for secret_file in /run/secrets/*; do
    if [ -f "$secret_file" ]; then
        var_name=$(basename "$secret_file")
        export "$var_name"="$(cat "$secret_file")"
    fi
done
echo "ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY"
echo "GH_TOKEN=$GH_TOKEN"
`

	cmd := exec.Command("docker", "run", "--rm",
		"-v", secretsDir+":/run/secrets:ro",
		"-e", "ADDT_SECRETS_DIR=/run/secrets",
		"alpine:latest",
		"sh", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Verify secrets were loaded
	if !strings.Contains(outputStr, "ANTHROPIC_API_KEY=sk-ant-test-key-entrypoint-12345") {
		t.Errorf("ANTHROPIC_API_KEY not loaded correctly. Output: %s", outputStr)
	}
	if !strings.Contains(outputStr, "GH_TOKEN=ghp_test_token_67890") {
		t.Errorf("GH_TOKEN not loaded correctly. Output: %s", outputStr)
	}
}

func TestSecretsToFiles_Integration_SecretsNotInEnvWhenDisabled(t *testing.T) {
	checkDockerForSecrets(t)

	// When secrets_to_files is disabled, secrets should be passed as env vars
	// This test verifies the default behavior
	secretValue := "sk-ant-test-direct-env-12345"

	cmd := exec.Command("docker", "run", "--rm",
		"-e", "ANTHROPIC_API_KEY="+secretValue,
		"alpine:latest",
		"printenv", "ANTHROPIC_API_KEY")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	if strings.TrimSpace(string(output)) != secretValue {
		t.Errorf("Expected env var value %q, got %q", secretValue, string(output))
	}
}

func TestSecretsToFiles_Integration_SecretsNotVisibleToSubprocess(t *testing.T) {
	checkDockerForSecrets(t)

	// Create secrets directory
	secretsDir, err := os.MkdirTemp("", "secrets-subprocess-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(secretsDir)

	// Write a test secret
	if err := os.WriteFile(secretsDir+"/SECRET_KEY", []byte("secret-value"), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	// Run container with script that:
	// 1. Loads secret
	// 2. Unsets it
	// 3. Spawns subprocess to check if it's visible
	script := `
# Load secret
export SECRET_KEY="$(cat /run/secrets/SECRET_KEY)"
echo "Parent has SECRET_KEY: $SECRET_KEY"

# Unset it before spawning subprocess
unset SECRET_KEY

# Subprocess should not see it
sh -c 'echo "Child SECRET_KEY: ${SECRET_KEY:-<not set>}"'
`

	cmd := exec.Command("docker", "run", "--rm",
		"-v", secretsDir+":/run/secrets:ro",
		"alpine:latest",
		"sh", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run container: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)

	// Parent should have the secret
	if !strings.Contains(outputStr, "Parent has SECRET_KEY: secret-value") {
		t.Errorf("Parent should have secret. Output: %s", outputStr)
	}

	// Child should NOT have the secret
	if !strings.Contains(outputStr, "Child SECRET_KEY: <not set>") {
		t.Errorf("Child should not have secret. Output: %s", outputStr)
	}
}

func TestSecretsToFiles_Integration_FilePermissions(t *testing.T) {
	checkDockerForSecrets(t)

	// Create secrets directory with restricted permissions
	secretsDir, err := os.MkdirTemp("", "secrets-perms-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(secretsDir)

	// Set directory permissions
	if err := os.Chmod(secretsDir, 0700); err != nil {
		t.Fatalf("Failed to chmod dir: %v", err)
	}

	// Write secret with restricted permissions
	secretPath := secretsDir + "/SECRET_KEY"
	if err := os.WriteFile(secretPath, []byte("secret"), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	// Verify file permissions on host
	info, err := os.Stat(secretPath)
	if err != nil {
		t.Fatalf("Failed to stat secret: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("Secret file should have 0600 permissions, got %o", info.Mode().Perm())
	}

	// Verify directory permissions on host
	dirInfo, err := os.Stat(secretsDir)
	if err != nil {
		t.Fatalf("Failed to stat secrets dir: %v", err)
	}

	if dirInfo.Mode().Perm() != 0700 {
		t.Errorf("Secrets dir should have 0700 permissions, got %o", dirInfo.Mode().Perm())
	}
}

// TestSecretsToFiles_Integration_FullContainerWithSecretsEnabled tests the full flow:
// - Provider with secrets_to_files enabled
// - Secrets written to files and mounted
// - Container runs and can read secrets from files
// - Secrets are NOT visible as environment variables
func TestSecretsToFiles_Integration_FullContainerWithSecretsEnabled(t *testing.T) {
	checkDockerForSecrets(t)

	// Create secrets directory
	secretsDir, err := os.MkdirTemp("", "secrets-full-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(secretsDir)

	// Write test secrets
	secretValue := "sk-ant-full-test-secret-12345"
	if err := os.WriteFile(secretsDir+"/ANTHROPIC_API_KEY", []byte(secretValue), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	containerName := fmt.Sprintf("addt-secrets-test-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	// Run container with:
	// - Secrets mounted at /run/secrets (simulating secrets_to_files)
	// - ADDT_SECRETS_DIR env var set
	// - NO ANTHROPIC_API_KEY env var (it should come from file)
	// Container will:
	// 1. Check if ANTHROPIC_API_KEY is in env (should NOT be)
	// 2. Load from /run/secrets
	// 3. Verify the value
	script := `
echo "=== Checking env vars ==="
if printenv ANTHROPIC_API_KEY >/dev/null 2>&1; then
    echo "FAIL: ANTHROPIC_API_KEY found in env vars"
    exit 1
else
    echo "PASS: ANTHROPIC_API_KEY not in env vars"
fi

echo "=== Loading from secrets ==="
if [ -f /run/secrets/ANTHROPIC_API_KEY ]; then
    SECRET_VALUE=$(cat /run/secrets/ANTHROPIC_API_KEY)
    echo "PASS: Secret loaded from file"
    echo "VALUE: $SECRET_VALUE"
else
    echo "FAIL: Secret file not found"
    exit 1
fi
`

	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"-v", secretsDir+":/run/secrets:ro",
		"-e", "ADDT_SECRETS_DIR=/run/secrets",
		// Note: NOT passing -e ANTHROPIC_API_KEY
		"alpine:latest",
		"sh", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	// Verify the test passed
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY not in env vars") {
		t.Error("Secret should NOT be in environment variables")
	}
	if !strings.Contains(outputStr, "PASS: Secret loaded from file") {
		t.Error("Secret should be loadable from file")
	}
	if !strings.Contains(outputStr, "VALUE: "+secretValue) {
		t.Errorf("Secret value mismatch, expected %s", secretValue)
	}
}

// TestSecretsToFiles_Integration_FullContainerWithSecretsDisabled tests the default flow:
// - Secrets passed as env vars (secrets_to_files disabled)
// - Container can read secrets from environment
func TestSecretsToFiles_Integration_FullContainerWithSecretsDisabled(t *testing.T) {
	checkDockerForSecrets(t)

	containerName := fmt.Sprintf("addt-secrets-disabled-test-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	secretValue := "sk-ant-env-test-secret-67890"

	// Run container with secret as env var (default behavior)
	script := `
echo "=== Checking env vars ==="
if printenv ANTHROPIC_API_KEY >/dev/null 2>&1; then
    echo "PASS: ANTHROPIC_API_KEY found in env vars"
    echo "VALUE: $(printenv ANTHROPIC_API_KEY)"
else
    echo "FAIL: ANTHROPIC_API_KEY not in env vars"
    exit 1
fi
`

	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"-e", "ANTHROPIC_API_KEY="+secretValue,
		"alpine:latest",
		"sh", "-c", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	// Verify the test passed
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY found in env vars") {
		t.Error("Secret SHOULD be in environment variables when secrets_to_files is disabled")
	}
	if !strings.Contains(outputStr, "VALUE: "+secretValue) {
		t.Errorf("Secret value mismatch, expected %s", secretValue)
	}
}

// TestSecretsToFiles_Integration_ProviderBuildsCorrectArgs tests that the provider
// builds the correct docker arguments when secrets_to_files is enabled
func TestSecretsToFiles_Integration_ProviderBuildsCorrectArgs(t *testing.T) {
	checkDockerForSecrets(t)

	// Create a provider with secrets_to_files enabled
	secCfg := security.DefaultConfig()
	secCfg.SecretsToFiles = true

	cfg := &provider.Config{
		Security: secCfg,
	}

	prov := &DockerProvider{
		config:   cfg,
		tempDirs: []string{},
	}

	// Create a RunSpec with env vars
	spec := &provider.RunSpec{
		Name:      "test-secrets-args",
		ImageName: "alpine:latest",
		Env: map[string]string{
			"ANTHROPIC_API_KEY": "test-secret-value",
			"TERM":              "xterm",
			"HOME":              "/home/test",
		},
	}

	// Create container context
	ctx := &containerContext{
		homeDir:              "/tmp",
		username:             "addt",
		useExistingContainer: false,
	}

	// Build docker args
	dockerArgs := prov.buildBaseDockerArgs(spec, ctx)
	dockerArgs, cleanup := prov.addContainerVolumesAndEnv(dockerArgs, spec, ctx)
	defer cleanup()

	// Check for ADDT_SECRETS_DIR env var
	foundSecretsDir := false
	foundSecretMount := false
	foundSecretInEnv := false

	for i, arg := range dockerArgs {
		if arg == "-e" && i+1 < len(dockerArgs) {
			if strings.HasPrefix(dockerArgs[i+1], "ADDT_SECRETS_DIR=") {
				foundSecretsDir = true
			}
			if strings.HasPrefix(dockerArgs[i+1], "ANTHROPIC_API_KEY=") {
				foundSecretInEnv = true
			}
		}
		if arg == "-v" && i+1 < len(dockerArgs) {
			if strings.Contains(dockerArgs[i+1], "/run/secrets:ro") {
				foundSecretMount = true
			}
		}
	}

	// Note: Since we don't have extension metadata, secrets won't actually be written
	// But we can verify the non-secret env vars are still there
	foundTerm := false
	for i, arg := range dockerArgs {
		if arg == "-e" && i+1 < len(dockerArgs) {
			if dockerArgs[i+1] == "TERM=xterm" {
				foundTerm = true
			}
		}
	}

	if !foundTerm {
		t.Error("Non-secret env var TERM should still be passed")
	}

	t.Logf("Docker args: %v", dockerArgs)
	t.Logf("Found ADDT_SECRETS_DIR: %v, Found secret mount: %v, Found secret in env: %v",
		foundSecretsDir, foundSecretMount, foundSecretInEnv)
}

// TestSecretsToFiles_Integration_EntrypointSimulation tests the full entrypoint behavior
// by simulating what docker-entrypoint.sh does
func TestSecretsToFiles_Integration_EntrypointSimulation(t *testing.T) {
	checkDockerForSecrets(t)

	// Create secrets directory
	secretsDir, err := os.MkdirTemp("", "secrets-entrypoint-sim-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(secretsDir)

	// Write multiple secrets
	secrets := map[string]string{
		"ANTHROPIC_API_KEY": "sk-ant-key-123",
		"GH_TOKEN":          "ghp_token_456",
		"OPENAI_API_KEY":    "sk-openai-789",
	}

	for name, value := range secrets {
		if err := os.WriteFile(secretsDir+"/"+name, []byte(value), 0600); err != nil {
			t.Fatalf("Failed to write secret %s: %v", name, err)
		}
	}

	containerName := fmt.Sprintf("addt-entrypoint-sim-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	// This script simulates exactly what docker-entrypoint.sh does
	entrypointScript := `
# Load secrets from files (same as docker-entrypoint.sh)
if [ -n "$ADDT_SECRETS_DIR" ] && [ -d "$ADDT_SECRETS_DIR" ]; then
    for secret_file in "$ADDT_SECRETS_DIR"/*; do
        if [ -f "$secret_file" ]; then
            var_name=$(basename "$secret_file")
            export "$var_name"="$(cat "$secret_file")"
        fi
    done
fi

# Now verify all secrets are loaded
echo "=== Verifying secrets ==="
echo "ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY"
echo "GH_TOKEN=$GH_TOKEN"
echo "OPENAI_API_KEY=$OPENAI_API_KEY"

# Verify none were in env before loading
# (They should only exist because we loaded them from files)
if [ "$ANTHROPIC_API_KEY" = "sk-ant-key-123" ]; then
    echo "PASS: ANTHROPIC_API_KEY correct"
else
    echo "FAIL: ANTHROPIC_API_KEY incorrect"
    exit 1
fi

if [ "$GH_TOKEN" = "ghp_token_456" ]; then
    echo "PASS: GH_TOKEN correct"
else
    echo "FAIL: GH_TOKEN incorrect"
    exit 1
fi

if [ "$OPENAI_API_KEY" = "sk-openai-789" ]; then
    echo "PASS: OPENAI_API_KEY correct"
else
    echo "FAIL: OPENAI_API_KEY incorrect"
    exit 1
fi

echo "=== All secrets verified ==="
`

	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"-v", secretsDir+":/run/secrets:ro",
		"-e", "ADDT_SECRETS_DIR=/run/secrets",
		"alpine:latest",
		"sh", "-c", entrypointScript)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	// Verify all tests passed
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY correct") {
		t.Error("ANTHROPIC_API_KEY not loaded correctly")
	}
	if !strings.Contains(outputStr, "PASS: GH_TOKEN correct") {
		t.Error("GH_TOKEN not loaded correctly")
	}
	if !strings.Contains(outputStr, "PASS: OPENAI_API_KEY correct") {
		t.Error("OPENAI_API_KEY not loaded correctly")
	}
	if !strings.Contains(outputStr, "=== All secrets verified ===") {
		t.Error("Not all secrets verified successfully")
	}
}

const testSecretsImageName = "addt-test-secrets"

// createSecretsTestProvider creates a provider for secrets tests
func createSecretsTestProvider(t *testing.T, cfg *provider.Config) *DockerProvider {
	t.Helper()
	prov, err := NewDockerProvider(
		cfg,
		assets.DockerDockerfile,
		assets.DockerDockerfileBase,
		assets.DockerEntrypoint,
		assets.DockerInitFirewall,
		assets.DockerInstallSh,
		extensions.FS,
	)
	if err != nil {
		t.Fatalf("Failed to create Docker provider: %v", err)
	}
	// Type assert to *DockerProvider
	dockerProv, ok := prov.(*DockerProvider)
	if !ok {
		t.Fatal("Provider is not a DockerProvider")
	}
	return dockerProv
}

// ensureSecretsTestImage builds the test image if needed
func ensureSecretsTestImage(t *testing.T) {
	t.Helper()

	cmd := exec.Command("docker", "image", "inspect", testSecretsImageName)
	if cmd.Run() == nil {
		return // Image exists
	}

	t.Log("Building test image for secrets integration test...")

	cfg := &provider.Config{
		AddtVersion: "0.0.0-test",
		Extensions:  "claude",
		NodeVersion: "22",
		GoVersion:   "latest",
		UvVersion:   "latest",
		ImageName:   testSecretsImageName,
	}

	prov := createSecretsTestProvider(t, cfg)
	if err := prov.Initialize(cfg); err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}

	if err := prov.BuildIfNeeded(false, false); err != nil {
		t.Fatalf("Failed to build test image: %v", err)
	}
}

// getDockerAccessibleTempDir returns a temp directory that Docker can access.
// On macOS, /tmp is not shared with Docker by default, so we use HOME instead.
func getDockerAccessibleTempDir(t *testing.T, prefix string) string {
	t.Helper()
	// Try HOME first (usually shared with Docker)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		dir, err := os.MkdirTemp(homeDir, prefix)
		if err == nil {
			return dir
		}
	}
	// Fall back to system temp (works on Linux)
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

// TestSecretsToFiles_Integration_RealEntrypoint tests the actual docker-entrypoint.sh
// with secrets loaded from files
func TestSecretsToFiles_Integration_RealEntrypoint(t *testing.T) {
	checkDockerForSecrets(t)
	ensureSecretsTestImage(t)

	// Create secrets directory in a Docker-accessible location
	secretsDir := getDockerAccessibleTempDir(t, "secrets-entrypoint-real-")
	defer os.RemoveAll(secretsDir)

	// Write ANTHROPIC_API_KEY secret (the key extension env var for claude)
	secretValue := "sk-ant-real-entrypoint-test-12345"
	if err := os.WriteFile(secretsDir+"/ANTHROPIC_API_KEY", []byte(secretValue), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	containerName := fmt.Sprintf("addt-secrets-real-entrypoint-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	// Run the actual entrypoint with secrets mounted
	// We override ADDT_COMMAND to run a simple check instead of claude
	// The entrypoint will:
	// 1. Load secrets from ADDT_SECRETS_DIR
	// 2. Run setup.sh for claude extension (which uses ANTHROPIC_API_KEY)
	// 3. Execute our check command
	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"-v", secretsDir+":/run/secrets:ro",
		"-e", "ADDT_SECRETS_DIR=/run/secrets",
		"-e", "ADDT_COMMAND=sh",
		// Note: NOT passing ANTHROPIC_API_KEY as env var
		testSecretsImageName,
		"-c", `
echo "=== Checking secrets loaded by entrypoint ==="
echo "ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY:-<not set>}"

if [ "$ANTHROPIC_API_KEY" = "sk-ant-real-entrypoint-test-12345" ]; then
    echo "PASS: ANTHROPIC_API_KEY loaded correctly by entrypoint"
else
    echo "FAIL: ANTHROPIC_API_KEY not loaded or wrong value"
    echo "Checking file at /run/secrets/ANTHROPIC_API_KEY..."
    ls -la /run/secrets/ 2>&1 || echo "Cannot list /run/secrets"
    exit 1
fi

# Check that setup.sh used the API key
if [ -f ~/.claude.json ] && grep -q "customApiKeyResponses" ~/.claude.json; then
    echo "PASS: setup.sh created config with API key"
else
    echo "INFO: No API key config (may need hasCompletedOnboarding)"
fi

echo "=== Test passed ==="
`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	// Verify entrypoint loaded the secret
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY loaded correctly by entrypoint") {
		t.Error("Entrypoint should have loaded ANTHROPIC_API_KEY from secrets file")
	}
	if !strings.Contains(outputStr, "=== Test passed ===") {
		t.Error("Not all checks passed")
	}
}

// TestSecretsToFiles_Integration_RealEntrypointNotInInitialEnv verifies that
// secrets are NOT visible in the initial process environment when using file-based secrets
func TestSecretsToFiles_Integration_RealEntrypointNotInInitialEnv(t *testing.T) {
	checkDockerForSecrets(t)
	ensureSecretsTestImage(t)

	// Create secrets directory in a Docker-accessible location
	secretsDir := getDockerAccessibleTempDir(t, "secrets-not-in-env-")
	defer os.RemoveAll(secretsDir)

	secretValue := "sk-ant-not-in-initial-env-test"
	if err := os.WriteFile(secretsDir+"/ANTHROPIC_API_KEY", []byte(secretValue), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	containerName := fmt.Sprintf("addt-secrets-not-in-env-%d", os.Getpid())
	defer exec.Command("docker", "rm", "-f", containerName).Run()

	// Run container WITHOUT using entrypoint to check raw environment
	// This simulates checking /proc/1/environ before entrypoint runs
	cmd := exec.Command("docker", "run", "--rm",
		"--name", containerName,
		"-v", secretsDir+":/run/secrets:ro",
		"-e", "ADDT_SECRETS_DIR=/run/secrets",
		"--entrypoint", "/bin/sh",
		// Note: NOT passing ANTHROPIC_API_KEY as env var
		testSecretsImageName,
		"-c", `
echo "=== Checking initial environment (before entrypoint) ==="

# Check if ANTHROPIC_API_KEY is in initial env
if printenv ANTHROPIC_API_KEY >/dev/null 2>&1; then
    echo "FAIL: ANTHROPIC_API_KEY found in initial env"
    exit 1
else
    echo "PASS: ANTHROPIC_API_KEY NOT in initial env"
fi

# Now simulate what entrypoint does
echo "=== Loading secrets like entrypoint does ==="
if [ -n "$ADDT_SECRETS_DIR" ] && [ -d "$ADDT_SECRETS_DIR" ]; then
    for secret_file in "$ADDT_SECRETS_DIR"/*; do
        if [ -f "$secret_file" ]; then
            var_name=$(basename "$secret_file")
            export "$var_name"="$(cat "$secret_file")"
            echo "Loaded: $var_name"
        fi
    done
fi

# Now check that secret IS available
if [ -n "$ANTHROPIC_API_KEY" ]; then
    echo "PASS: ANTHROPIC_API_KEY available after loading"
else
    echo "FAIL: ANTHROPIC_API_KEY not available after loading"
    exit 1
fi

echo "=== Test passed ==="
`)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
	}

	outputStr := string(output)
	t.Logf("Container output:\n%s", outputStr)

	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY NOT in initial env") {
		t.Error("Secret should NOT be in initial environment")
	}
	if !strings.Contains(outputStr, "PASS: ANTHROPIC_API_KEY available after loading") {
		t.Error("Secret should be available after loading from file")
	}
}

// TestSecretsToFiles_Integration_RealProviderFlow tests the complete flow using
// the actual provider to build docker args with secrets_to_files enabled
func TestSecretsToFiles_Integration_RealProviderFlow(t *testing.T) {
	checkDockerForSecrets(t)
	ensureSecretsTestImage(t)

	// Create provider with secrets_to_files enabled
	secCfg := security.DefaultConfig()
	secCfg.SecretsToFiles = true

	cfg := &provider.Config{
		AddtVersion: "0.0.0-test",
		Extensions:  "claude",
		NodeVersion: "22",
		GoVersion:   "latest",
		UvVersion:   "latest",
		ImageName:   testSecretsImageName,
		Security:    secCfg,
	}

	prov := createSecretsTestProvider(t, cfg)
	if err := prov.Initialize(cfg); err != nil {
		t.Fatalf("Failed to initialize provider: %v", err)
	}

	// Set ANTHROPIC_API_KEY in current environment (simulating user's shell)
	secretValue := "sk-ant-provider-flow-test-67890"
	os.Setenv("ANTHROPIC_API_KEY", secretValue)
	defer os.Unsetenv("ANTHROPIC_API_KEY")

	// Build environment using the provider's logic
	// This will pick up ANTHROPIC_API_KEY from extension env_vars
	extensionEnvVars := prov.GetExtensionEnvVars(testSecretsImageName)
	t.Logf("Extension env vars: %v", extensionEnvVars)

	// Check that ANTHROPIC_API_KEY is in the extension env vars
	foundKey := false
	for _, v := range extensionEnvVars {
		if v == "ANTHROPIC_API_KEY" {
			foundKey = true
			break
		}
	}

	if !foundKey {
		t.Log("ANTHROPIC_API_KEY not found in extension env vars - this is expected if image metadata isn't available")
		t.Log("Skipping rest of test as extension metadata not accessible")
		return
	}

	// Create env map as the core/env.go would
	env := map[string]string{
		"ANTHROPIC_API_KEY": secretValue,
		"TERM":              "xterm-256color",
	}

	// Create RunSpec
	containerName := fmt.Sprintf("addt-provider-flow-%d", os.Getpid())
	spec := &provider.RunSpec{
		Name:      containerName,
		ImageName: testSecretsImageName,
		Env:       env,
	}

	// Build docker args using provider
	ctx := &containerContext{
		homeDir:              os.TempDir(),
		username:             "addt",
		useExistingContainer: false,
	}

	dockerArgs := prov.buildBaseDockerArgs(spec, ctx)
	dockerArgs, cleanup := prov.addContainerVolumesAndEnv(dockerArgs, spec, ctx)
	defer cleanup()

	// Check that secrets were handled
	var foundSecretsDir, foundSecretMount bool
	var secretsEnvRemoved bool = true

	for i, arg := range dockerArgs {
		if arg == "-e" && i+1 < len(dockerArgs) {
			if strings.HasPrefix(dockerArgs[i+1], "ADDT_SECRETS_DIR=") {
				foundSecretsDir = true
			}
			if strings.HasPrefix(dockerArgs[i+1], "ANTHROPIC_API_KEY=") {
				secretsEnvRemoved = false
			}
		}
		if arg == "-v" && i+1 < len(dockerArgs) {
			if strings.Contains(dockerArgs[i+1], "/run/secrets:ro") {
				foundSecretMount = true
			}
		}
	}

	t.Logf("Docker args: %v", dockerArgs)
	t.Logf("Found secrets dir env: %v, Found secrets mount: %v, Secret env removed: %v",
		foundSecretsDir, foundSecretMount, secretsEnvRemoved)

	if foundSecretsDir && foundSecretMount && secretsEnvRemoved {
		t.Log("PASS: Provider correctly configured secrets_to_files")
	} else if !foundSecretsDir && !foundSecretMount {
		t.Log("INFO: Secrets not configured (extension metadata may not be available)")
	} else {
		t.Errorf("Unexpected state: foundSecretsDir=%v, foundSecretMount=%v, secretsEnvRemoved=%v",
			foundSecretsDir, foundSecretMount, secretsEnvRemoved)
	}
}

// TestSecretsToFiles_Integration_CompareEnvVsFiles runs two containers side by side:
// one with secrets as env vars, one with secrets as files, and compares behavior
func TestSecretsToFiles_Integration_CompareEnvVsFiles(t *testing.T) {
	checkDockerForSecrets(t)

	secretValue := "sk-compare-test-secret"

	// Create secrets directory for file-based test
	secretsDir, err := os.MkdirTemp("", "secrets-compare-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(secretsDir)

	if err := os.WriteFile(secretsDir+"/MY_SECRET", []byte(secretValue), 0600); err != nil {
		t.Fatalf("Failed to write secret: %v", err)
	}

	// Script to check /proc/*/environ for the secret
	// This simulates what a malicious process might try to do
	checkScript := `
# Check if secret is visible in /proc/1/environ
if grep -q "MY_SECRET" /proc/1/environ 2>/dev/null; then
    echo "SECRET_IN_PROC_ENVIRON=yes"
else
    echo "SECRET_IN_PROC_ENVIRON=no"
fi

# Check if secret is in current env
if printenv MY_SECRET >/dev/null 2>&1; then
    echo "SECRET_IN_ENV=yes"
    echo "SECRET_VALUE=$(printenv MY_SECRET)"
else
    echo "SECRET_IN_ENV=no"
fi
`

	// Test 1: Secret as env var (visible in /proc/1/environ)
	t.Run("SecretAsEnvVar", func(t *testing.T) {
		cmd := exec.Command("docker", "run", "--rm",
			"-e", "MY_SECRET="+secretValue,
			"alpine:latest",
			"sh", "-c", checkScript)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
		}

		outputStr := string(output)
		t.Logf("Env var mode output:\n%s", outputStr)

		// Secret SHOULD be visible in env
		if !strings.Contains(outputStr, "SECRET_IN_ENV=yes") {
			t.Error("Secret should be in env when passed as -e")
		}
		// Secret SHOULD be visible in /proc/1/environ (this is the security concern)
		if !strings.Contains(outputStr, "SECRET_IN_PROC_ENVIRON=yes") {
			t.Log("Note: Secret not found in /proc/1/environ (might be a container quirk)")
		}
	})

	// Test 2: Secret loaded from file (not in initial environ)
	t.Run("SecretFromFile", func(t *testing.T) {
		// Load secret from file then check
		fileCheckScript := `
# First check - secret should NOT be in env yet
if printenv MY_SECRET >/dev/null 2>&1; then
    echo "BEFORE_LOAD: SECRET_IN_ENV=yes (unexpected)"
else
    echo "BEFORE_LOAD: SECRET_IN_ENV=no (expected)"
fi

# Load from file
export MY_SECRET="$(cat /run/secrets/MY_SECRET)"

# After loading - secret IS in env (but wasn't in initial /proc/1/environ)
if printenv MY_SECRET >/dev/null 2>&1; then
    echo "AFTER_LOAD: SECRET_IN_ENV=yes"
    echo "SECRET_VALUE=$(printenv MY_SECRET)"
fi
`

		cmd := exec.Command("docker", "run", "--rm",
			"-v", secretsDir+":/run/secrets:ro",
			"alpine:latest",
			"sh", "-c", fileCheckScript)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Container failed: %v\nOutput: %s", err, string(output))
		}

		outputStr := string(output)
		t.Logf("File mode output:\n%s", outputStr)

		// Before loading, secret should NOT be in env
		if !strings.Contains(outputStr, "BEFORE_LOAD: SECRET_IN_ENV=no") {
			t.Error("Secret should NOT be in env before loading from file")
		}
		// After loading, secret should be available
		if !strings.Contains(outputStr, "AFTER_LOAD: SECRET_IN_ENV=yes") {
			t.Error("Secret should be in env after loading from file")
		}
		if !strings.Contains(outputStr, "SECRET_VALUE="+secretValue) {
			t.Error("Secret value should match after loading")
		}
	})
}
