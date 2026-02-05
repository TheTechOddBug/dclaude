package extensions

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunCredentialScript runs an extension's credential script and returns env vars
// The script runs on the host and outputs KEY=value pairs to stdout
// Returns a map of environment variable names to values
func RunCredentialScript(ext *ExtensionConfig) (map[string]string, error) {
	if ext.CredentialScript == "" {
		return nil, nil
	}

	// Find the script path
	scriptPath, err := findCredentialScript(ext)
	if err != nil {
		return nil, err
	}

	if scriptPath == "" {
		return nil, nil
	}

	// Run the script
	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Stderr = os.Stderr // Show errors to user

	output, err := cmd.Output()
	if err != nil {
		// Script failure is not fatal - might just mean no credentials available
		return nil, nil
	}

	// Parse KEY=value output
	envVars := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=value
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])

			// Only accept keys that look like env vars (uppercase, underscores)
			if isValidEnvVarName(key) {
				envVars[key] = value
			}
		}
	}

	return envVars, nil
}

// findCredentialScript locates the credential script for an extension
func findCredentialScript(ext *ExtensionConfig) (string, error) {
	scriptName := ext.CredentialScript

	// Check local extension directory first (~/.addt/extensions/<name>/)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		localPath := filepath.Join(homeDir, ".addt", "extensions", ext.Name, scriptName)
		if fileExists(localPath) {
			return localPath, nil
		}
	}

	// For embedded extensions, we need to extract the script to a temp location
	// Check if script exists in embedded filesystem
	embeddedPath := fmt.Sprintf("%s/%s", ext.Name, scriptName)
	content, err := FS.ReadFile(embeddedPath)
	if err != nil {
		// Script doesn't exist - not an error, just no credentials
		return "", nil
	}

	// Write to temp file
	tmpDir, err := os.MkdirTemp("", "addt-cred-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	tmpScript := filepath.Join(tmpDir, scriptName)
	if err := os.WriteFile(tmpScript, content, 0700); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to write temp script: %w", err)
	}

	return tmpScript, nil
}

// isValidEnvVarName checks if a string is a valid environment variable name
func isValidEnvVarName(name string) bool {
	if name == "" {
		return false
	}

	for i, c := range name {
		if i == 0 {
			// First char must be letter or underscore
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_') {
				return false
			}
		} else {
			// Subsequent chars can be letter, digit, or underscore
			if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}

	return true
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
