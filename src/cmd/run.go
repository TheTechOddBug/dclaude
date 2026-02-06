package cmd

import (
	"fmt"
	"os"

	extcmd "github.com/jedi4ever/addt/cmd/extensions"
	"github.com/jedi4ever/addt/util"
)

// HandleRunCommand handles the "addt run <extension>" command.
// It validates the extension and sets up environment variables.
// Returns the remaining args for execution, or nil if the command was fully handled (help/error).
func HandleRunCommand(args []string) []string {
	runLogger := util.Log("run")
	runLogger.Debugf("HandleRunCommand called with args: %v", args)

	if len(args) < 1 {
		runLogger.Debug("No extension specified, showing help")
		printRunHelp()
		return nil
	}

	extName := args[0]
	runLogger.Debugf("Extension name: %s", extName)

	// Check for help flag
	if extName == "--help" || extName == "-h" {
		runLogger.Debug("Help flag detected, showing help")
		printRunHelp()
		return nil
	}

	// Validate extension exists
	runLogger.Debugf("Checking if extension '%s' exists", extName)
	if !extcmd.Exists(extName) {
		runLogger.Errorf("Extension '%s' does not exist", extName)
		fmt.Printf("Error: extension '%s' does not exist\n", extName)
		fmt.Println("Run 'addt extensions list' to see available extensions")
		os.Exit(1)
	}
	runLogger.Debugf("Extension '%s' exists", extName)

	// Get entrypoint
	entrypoint := extcmd.GetEntrypoint(extName)
	runLogger.Debugf("Extension entrypoint: %s", entrypoint)

	// Set the extension environment variables
	runLogger.Debugf("Setting ADDT_EXTENSIONS=%s", extName)
	os.Setenv("ADDT_EXTENSIONS", extName)
	runLogger.Debugf("Setting ADDT_COMMAND=%s", entrypoint)
	os.Setenv("ADDT_COMMAND", entrypoint)

	// Return remaining args for execution
	if len(args) > 1 {
		remainingArgs := args[1:]
		runLogger.Debugf("Returning remaining args for execution: %v", remainingArgs)
		return remainingArgs
	}
	runLogger.Debug("No remaining args, returning empty slice")
	return []string{}
}

func printRunHelp() {
	fmt.Println("Usage: addt run <extension> [args...]")
	fmt.Println()
	fmt.Println("Run a specific extension in a container.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  <extension>    Name of the extension to run")
	fmt.Println("  [args...]      Arguments to pass to the extension")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  addt run claude \"Fix the bug\"")
	fmt.Println("  addt run codex --help")
	fmt.Println("  addt run gemini")
	fmt.Println()
	fmt.Println("To see available extensions:")
	fmt.Println("  addt extensions list")
}
