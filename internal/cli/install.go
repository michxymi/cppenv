package cli

import (
	"fmt"

	"github.com/michxymi/cppenv/internal/config"
	"github.com/michxymi/cppenv/internal/environment"
	"github.com/michxymi/cppenv/internal/python"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install tools into the project environment",
	Long:  `Downloads Python if needed, creates a virtual environment, and installs all configured tools.`,
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Find and load config
	configPath, err := config.FindConfig()
	if err != nil {
		return fmt.Errorf("no cppenv.toml found, run 'cppenv init' first")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Ensure Python is available
	fmt.Println("Checking Python...")
	pythonPath, err := python.Ensure()
	if err != nil {
		return fmt.Errorf("failed to set up Python: %w", err)
	}
	fmt.Printf("Using Python at: %s\n", pythonPath)

	// Create venv if needed
	if !environment.Exists() {
		fmt.Println("Creating environment...")
		if err := environment.Create(pythonPath); err != nil {
			return err
		}
	} else {
		fmt.Println("Environment already exists")
	}

	// Install tools
	fmt.Println("\nInstalling tools...")
	reqs := cfg.GetRequirements()
	for _, req := range reqs {
		fmt.Printf("  %s\n", req)
	}
	if err := environment.InstallTools(reqs); err != nil {
		return err
	}

	// Update .gitignore
	if added, err := environment.AddToGitignore(); err != nil {
		fmt.Printf("Warning: could not update .gitignore: %v\n", err)
	} else if added {
		fmt.Println("Added .cppenv/ to .gitignore")
	}

	// Create conan_provider.cmake if it doesn't exist
	if created, err := environment.CreateConanProvider(); err != nil {
		return fmt.Errorf("failed to create conan_provider.cmake: %w", err)
	} else if created {
		fmt.Println("Created conan_provider.cmake in .cppenv/")
	}

	// Create CMakeUserPresets.json if it doesn't exist
	if created, err := environment.CreateCMakeUserPresets(); err != nil {
		return fmt.Errorf("failed to create CMakeUserPresets.json: %w", err)
	} else if created {
		fmt.Println("Created CMakeUserPresets.json in .cppenv/")
	}

	fmt.Println("\nDone! You can now use 'cppenv run <command>' to run tools.")
	return nil
}
