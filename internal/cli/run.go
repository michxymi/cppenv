package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/michxymi/cppenv/internal/config"
	"github.com/michxymi/cppenv/internal/environment"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <command or script> [args...]",
	Short: "Run a script or command using the project environment",
	Long: `Executes a command with the cppenv tools in PATH.

If the first argument matches a script name defined in cppenv.toml's [scripts]
section, that script will be executed. Otherwise, the command is run directly.

Examples:
  cppenv run cmake --version
  cppenv run build              # runs script named "build" from cppenv.toml`,
	Args:               cobra.MinimumNArgs(1),
	RunE:               runRun,
	DisableFlagParsing: true,
}

func runRun(cmd *cobra.Command, args []string) error {
	if !environment.Exists() {
		return fmt.Errorf("environment not found, run 'cppenv install' first")
	}

	// Try to load config for scripts
	var cfg *config.Config
	if configPath, err := config.FindConfig(); err == nil {
		cfg, _ = config.Load(configPath)
	}

	// Check if first arg is a script name
	if cfg != nil && cfg.Scripts != nil {
		if script, ok := cfg.Scripts[args[0]]; ok {
			fmt.Printf("→ %s\n", script)
			return runScript(script)
		}
	}

	// Run as direct command
	exitCode := environment.RunCommand(args)
	os.Exit(exitCode)
	return nil
}

func runScript(script string) error {
	var args []string
	if runtime.GOOS == "windows" {
		args = []string{"cmd", "/c", script}
	} else {
		args = []string{"sh", "-c", script}
	}

	exitCode := environment.RunCommand(args)
	os.Exit(exitCode)
	return nil
}
