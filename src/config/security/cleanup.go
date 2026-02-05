package security

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const pidFileName = ".addt.pid"

// WritePIDFile writes the current process PID to a file in the given directory.
// This allows cleanup to identify orphaned directories from crashed processes.
func WritePIDFile(dir string) error {
	pidFile := filepath.Join(dir, pidFileName)
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0600)
}

// isProcessRunning checks if a process with the given PID is still running.
func isProcessRunning(pid int) bool {
	// On Unix, sending signal 0 checks if process exists without actually sending a signal
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Signal 0 doesn't actually send a signal, just checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// isOrphanedDir checks if a directory was created by a process that is no longer running.
// Returns true if the directory is orphaned (safe to clean up), false if owner is still running.
func isOrphanedDir(dir string) bool {
	pidFile := filepath.Join(dir, pidFileName)

	data, err := os.ReadFile(pidFile)
	if err != nil {
		// No PID file - could be old directory from before PID tracking
		// Fall back to age-based check: only consider orphaned if very old (24+ hours)
		info, err := os.Stat(dir)
		if err != nil {
			return false
		}
		// Be conservative: 24 hours old without PID file means likely orphaned
		return info.ModTime().Before(time.Now().Add(-24 * time.Hour))
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return false
	}

	// If process is not running, directory is orphaned
	return !isProcessRunning(pid)
}

// CleanupStaleTempDirs removes orphaned addt temporary directories from previous runs.
// Only removes directories where the creating process is no longer running.
func CleanupStaleTempDirs() error {
	tmpDir := os.TempDir()

	// Patterns for addt temp directories
	patterns := []string{
		"ssh-proxy-*",
		"gpg-proxy-*",
		"addt-secrets-*",
		"ssh-safe-*",
		"gpg-safe-*",
	}

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(tmpDir, pattern))
		if err != nil {
			continue
		}

		for _, match := range matches {
			if isOrphanedDir(match) {
				os.RemoveAll(match)
			}
		}
	}

	return nil
}

// CleanupOldSeccompProfiles removes stale seccomp profiles from temp directory.
// Seccomp profiles are single files without PID tracking, so use age-based cleanup.
// Only removes profiles older than 24 hours.
func CleanupOldSeccompProfiles() error {
	tmpDir := os.TempDir()
	cutoff := time.Now().Add(-24 * time.Hour)

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
	CleanupStaleTempDirs()
	CleanupOldSeccompProfiles()
}
