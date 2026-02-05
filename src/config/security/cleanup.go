package security

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CleanupStaleTempDirs removes stale addt temporary directories from previous runs.
// This handles cases where the process was killed (SIGKILL) or crashed without cleanup.
// Only removes directories older than maxAge to avoid removing directories from
// concurrent processes.
func CleanupStaleTempDirs(maxAge time.Duration) error {
	tmpDir := os.TempDir()

	// Patterns for addt temp directories
	patterns := []string{
		"ssh-proxy-*",
		"gpg-proxy-*",
		"addt-secrets-*",
		"ssh-safe-*",
		"gpg-safe-*",
	}

	cutoff := time.Now().Add(-maxAge)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(tmpDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue
			}

			// Only remove if directory is older than maxAge
			if info.ModTime().Before(cutoff) {
				os.RemoveAll(match)
			}
		}
	}

	return nil
}

// CleanupOldSeccompProfiles removes stale seccomp profiles from temp directory.
func CleanupOldSeccompProfiles(maxAge time.Duration) error {
	tmpDir := os.TempDir()
	cutoff := time.Now().Add(-maxAge)

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasPrefix(entry.Name(), "addt-seccomp-") && strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(tmpDir, entry.Name())
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				os.Remove(path)
			}
		}
	}

	return nil
}

// CleanupAll performs all cleanup operations for stale temporary files.
// Should be called during provider initialization.
func CleanupAll() {
	// Clean up directories older than 1 hour
	maxAge := 1 * time.Hour

	CleanupStaleTempDirs(maxAge)
	CleanupOldSeccompProfiles(maxAge)
}
