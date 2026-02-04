package extensions

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jedi4ever/addt/extensions"
)

// Clone copies a built-in extension to the local extensions directory for customization
// If targetName is empty, it uses the source name (overriding the built-in)
func Clone(sourceName, targetName string) {
	// Validate source name
	if sourceName == "" {
		fmt.Println("Error: extension name cannot be empty")
		os.Exit(1)
	}

	// If no target name, use source name
	if targetName == "" {
		targetName = sourceName
	}

	// Validate target name characters
	for _, c := range targetName {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			fmt.Println("Error: target name can only contain lowercase letters, numbers, hyphens, and underscores")
			os.Exit(1)
		}
	}

	// Check if extension exists in embedded FS
	if !isBuiltinExtension(sourceName) {
		fmt.Printf("Error: '%s' is not a built-in extension\n", sourceName)
		fmt.Println()
		fmt.Println("Use 'addt extensions list' to see available extensions.")
		fmt.Println("Use 'addt extensions new <name>' to create a new extension from scratch.")
		os.Exit(1)
	}

	// Get local extensions directory
	localDir := extensions.GetLocalExtensionsDir()
	if localDir == "" {
		fmt.Println("Error: could not determine local extensions directory")
		os.Exit(1)
	}

	destDir := filepath.Join(localDir, targetName)

	// Check if extension already exists locally
	if _, err := os.Stat(destDir); err == nil {
		fmt.Printf("Error: local extension '%s' already exists at %s\n", targetName, destDir)
		fmt.Println()
		fmt.Println("To re-clone, first remove the existing extension:")
		fmt.Printf("  addt extensions remove %s\n", targetName)
		os.Exit(1)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		fmt.Printf("Error: failed to create directory: %v\n", err)
		os.Exit(1)
	}

	// Copy all files from embedded FS
	filesCopied := 0
	renamed := sourceName != targetName
	err := fs.WalkDir(extensions.FS, sourceName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path within extension directory
		relPath, err := filepath.Rel(sourceName, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if d.IsDir() {
			if relPath != "." {
				return os.MkdirAll(destPath, 0755)
			}
			return nil
		}

		// Read file from embedded FS
		data, err := extensions.FS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// If renaming and this is config.yaml, update the name field
		if renamed && filepath.Base(path) == "config.yaml" {
			data = updateConfigName(data, sourceName, targetName)
		}

		// Determine file permissions (executable for .sh files)
		perm := os.FileMode(0644)
		if filepath.Ext(path) == ".sh" {
			perm = 0755
		}

		// Write file to local directory
		if err := os.WriteFile(destPath, data, perm); err != nil {
			return fmt.Errorf("failed to write %s: %w", destPath, err)
		}

		filesCopied++
		return nil
	})

	if err != nil {
		// Clean up on error
		os.RemoveAll(destDir)
		fmt.Printf("Error: failed to clone extension: %v\n", err)
		os.Exit(1)
	}

	if renamed {
		fmt.Printf("Cloned '%s' as '%s' to:\n", sourceName, targetName)
	} else {
		fmt.Printf("Cloned extension '%s' to:\n", sourceName)
	}
	fmt.Printf("  %s\n", destDir)
	fmt.Println()
	fmt.Printf("Files copied: %d\n", filesCopied)
	fmt.Println()
	if renamed {
		fmt.Printf("You now have a new extension '%s' based on '%s'.\n", targetName, sourceName)
	} else {
		fmt.Println("Your local copy now overrides the built-in extension.")
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Edit files in %s\n", destDir)
	fmt.Printf("  2. Build with: addt build %s\n", targetName)
	fmt.Printf("  3. Run with:   addt run %s\n", targetName)
	fmt.Println()
	fmt.Println("To remove this extension:")
	fmt.Printf("  addt extensions remove %s\n", targetName)
}

// isBuiltinExtension checks if an extension exists in the embedded FS
func isBuiltinExtension(name string) bool {
	entries, err := fs.ReadDir(extensions.FS, ".")
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == name {
			// Verify it has a config.yaml
			if _, err := extensions.FS.ReadFile(name + "/config.yaml"); err == nil {
				return true
			}
		}
	}
	return false
}

// updateConfigName replaces the name field in config.yaml content
func updateConfigName(data []byte, oldName, newName string) []byte {
	// Simple replacement of "name: oldName" with "name: newName"
	old := []byte("name: " + oldName)
	new := []byte("name: " + newName)
	return bytes.Replace(data, old, new, 1)
}
