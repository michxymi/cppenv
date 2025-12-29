package cli

import (
	"fmt"

	"github.com/michxymi/cppenv/internal/config"
	"github.com/michxymi/cppenv/internal/environment"
	"github.com/michxymi/cppenv/internal/python"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status and configured tools",
	Long:  `Displays the current project configuration, installed status, and tool versions.`,
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Try to load config
	configPath, err := config.FindConfig()
	if err != nil {
		fmt.Println("No cppenv.toml found in current directory.")
		fmt.Println("Run 'cppenv init' to create one.")
		return nil
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Project info
	fmt.Printf("Project: %s\n", cfg.Project.Name)
	fmt.Println()

	// Python status
	fmt.Println("Python:")
	if python.IsInstalled() {
		fmt.Printf("  Installed: %s\n", python.GetPythonPath())
	} else {
		fmt.Println("  Not installed (will download on 'cppenv install')")
	}
	fmt.Println()

	// Environment status
	fmt.Println("Environment:")
	if environment.Exists() {
		fmt.Printf("  Path: %s\n", environment.GetVenvPath())
		fmt.Println("  Status: installed")
	} else {
		fmt.Println("  Status: not installed")
		fmt.Println("  Run 'cppenv install' to set up")
	}
	fmt.Println()

	// Tools
	fmt.Println("Tools:")
	for pkg, version := range cfg.Tools {
		fmt.Printf("  %s: %s\n", pkg, version)
	}

	// Scripts
	if len(cfg.Scripts) > 0 {
		fmt.Println()
		fmt.Println("Scripts:")
		for name, script := range cfg.Scripts {
			fmt.Printf("  %s: %s\n", name, script)
		}
	}

	return nil
}
