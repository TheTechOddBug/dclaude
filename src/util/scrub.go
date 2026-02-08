package util

import (
	"crypto/rand"
	"fmt"
	"os"
)

// ScrubFile overwrites a file with random data (same size) before deletion.
// This prevents recovery of sensitive data from disk or filesystem caches.
func ScrubFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	size := info.Size()
	if size == 0 {
		return nil
	}

	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file for scrubbing: %w", err)
	}
	defer f.Close()

	randomData := make([]byte, size)
	if _, err := rand.Read(randomData); err != nil {
		return fmt.Errorf("failed to generate random data: %w", err)
	}

	if _, err := f.WriteAt(randomData, 0); err != nil {
		return fmt.Errorf("failed to overwrite file: %w", err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

// ScrubAndRemove overwrites a file with random data then removes it.
// Best-effort: logs warnings on failure but does not return errors.
func ScrubAndRemove(path string) {
	log := Log("scrub")
	if err := ScrubFile(path); err != nil {
		log.Warning("failed to scrub file before removal: %s: %v", path, err)
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		log.Warning("failed to remove file after scrubbing: %s: %v", path, err)
	}
}
