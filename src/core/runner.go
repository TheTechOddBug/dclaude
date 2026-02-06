package core

import (
	"fmt"

	"github.com/jedi4ever/addt/provider"
	"github.com/jedi4ever/addt/util"
)

var runnerLogger = util.Log("runner")

// Runner coordinates container execution
type Runner struct {
	provider provider.Provider
	config   *provider.Config
}

// NewRunner creates a new runner
func NewRunner(p provider.Provider, cfg *provider.Config) *Runner {
	return &Runner{
		provider: p,
		config:   cfg,
	}
}

// Run executes the container with the configured extension
func (r *Runner) Run(args []string) error {
	runnerLogger.Debugf("Runner.Run called with args: %v", args)
	return r.execute(args, false)
}

// Shell opens an interactive shell in the container
func (r *Runner) Shell(args []string) error {
	runnerLogger.Debugf("Runner.Shell called with args: %v", args)
	return r.execute(args, true)
}

// execute is the common execution logic for Run and Shell
func (r *Runner) execute(args []string, openShell bool) error {
	runnerLogger.Debugf("execute called with args: %v", args)
	runnerLogger.Debugf("Runner.execute called: openShell=%v, args=%v", openShell, args)

	// Determine container name
	runnerLogger.Debug("Generating container name")
	name := r.generateName()
	runnerLogger.Debugf("Generated container name: %s", name)

	// Build run options
	runnerLogger.Debug("Building run options")
	opts := BuildRunOptions(r.provider, r.config, name, args, openShell)
	runnerLogger.Debugf("Run options: Name=%s, ImageName=%s, Args=%v, Interactive=%v, Persistent=%v",
		opts.Name, opts.ImageName, opts.Args, opts.Interactive, opts.Persistent)

	// Display status
	runnerLogger.Debug("Displaying status")
	DisplayStatus(r.provider, r.config, name)

	// Execute via provider
	if openShell {
		runnerLogger.Debug("Calling provider.Shell")
		err := r.provider.Shell(opts)
		if err != nil {
			runnerLogger.Errorf("Provider.Shell failed: %v", err)
		} else {
			runnerLogger.Debug("Provider.Shell completed successfully")
		}
		return err
	}
	runnerLogger.Debug("Calling provider.Run")
	err := r.provider.Run(opts)
	if err != nil {
		runnerLogger.Errorf("Provider.Run failed: %v", err)
	} else {
		runnerLogger.Debug("Provider.Run completed successfully")
	}
	return err
}

// generateName generates the container name based on persistence mode
func (r *Runner) generateName() string {
	if r.config.Persistent {
		return r.provider.GeneratePersistentName()
	}
	return r.provider.GenerateEphemeralName()
}

// GetExtensionName returns the current extension name
func (r *Runner) GetExtensionName() string {
	if r.config.Command != "" {
		return r.config.Command
	}
	return "claude"
}

// DisplayWarning shows the experimental warning
func (r *Runner) DisplayWarning() {
	fmt.Printf("⚠ addt:%s is experimental - things are not perfect yet\n", r.GetExtensionName())
}
