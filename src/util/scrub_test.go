package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScrubFile(t *testing.T) {
	t.Run("overwrites file content with random data", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "secret.txt")
		original := []byte("super-secret-api-key-12345")

		if err := os.WriteFile(path, original, 0600); err != nil {
			t.Fatal(err)
		}

		if err := ScrubFile(path); err != nil {
			t.Fatalf("ScrubFile failed: %v", err)
		}

		// File should still exist with same size
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("file should still exist: %v", err)
		}
		if info.Size() != int64(len(original)) {
			t.Fatalf("expected size %d, got %d", len(original), info.Size())
		}

		// Content should be different from original
		scrubbed, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read scrubbed file: %v", err)
		}
		if string(scrubbed) == string(original) {
			t.Fatal("scrubbed content should differ from original")
		}
	})

	t.Run("handles empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")

		if err := os.WriteFile(path, []byte{}, 0600); err != nil {
			t.Fatal(err)
		}

		if err := ScrubFile(path); err != nil {
			t.Fatalf("ScrubFile should succeed on empty file: %v", err)
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		err := ScrubFile("/nonexistent/path/file.txt")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestScrubAndRemove(t *testing.T) {
	t.Run("scrubs and removes file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "secret.txt")

		if err := os.WriteFile(path, []byte("secret-data"), 0600); err != nil {
			t.Fatal(err)
		}

		ScrubAndRemove(path)

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("file should have been removed")
		}
	})

	t.Run("handles nonexistent file without panic", func(t *testing.T) {
		// Should not panic on nonexistent file
		ScrubAndRemove("/nonexistent/path/file.txt")
	})

	t.Run("handles already-removed file without panic", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "gone.txt")

		// File never created — ScrubAndRemove should not panic
		ScrubAndRemove(path)
	})
}
