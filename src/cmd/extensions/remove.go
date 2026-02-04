package extensions

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jedi4ever/addt/extensions"
)

// Remove deletes a local extension from ~/.addt/extensions/
func Remove(name string, force bool) {
	// Validate name
	if name == "" {
		fmt.Println("Error: extension name cannot be empty")
		os.Exit(1)
	}

	// Get local extensions directory
	localDir := extensions.GetLocalExtensionsDir()
	if localDir == "" {
		fmt.Println("Error: could not determine local extensions directory")
		os.Exit(1)
	}

	extDir := filepath.Join(localDir, name)

	// Check if extension exists locally
	info, err := os.Stat(extDir)
	if os.IsNotExist(err) {
		fmt.Printf("Error: no local extension '%s' found at %s\n", name, extDir)
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", extDir)
		os.Exit(1)
	}

	// Confirm removal unless --force is used
	if !force {
		fmt.Printf("This will remove the local extension '%s' at:\n", name)
		fmt.Printf("  %s\n", extDir)
		fmt.Println()

		if isBuiltinExtension(name) {
			fmt.Println("The built-in version will be used after removal.")
		} else {
			fmt.Println("Warning: This is a custom extension with no built-in fallback.")
		}

		fmt.Println()
		fmt.Print("Proceed? [y/N] ")

		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled.")
			return
		}
	}

	// Remove the directory
	if err := os.RemoveAll(extDir); err != nil {
		fmt.Printf("Error: failed to remove extension: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Removed local extension '%s'\n", name)

	if isBuiltinExtension(name) {
		fmt.Println()
		fmt.Printf("The built-in '%s' extension will now be used.\n", name)
		fmt.Printf("Rebuild with: addt build %s --force\n", name)
	}
}
