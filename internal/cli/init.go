package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/michxymi/cppenv/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new cppenv project",
	Long:  `Creates a cppenv.toml file with default tools and their latest versions from PyPI.`,
	RunE:  runInit,
}

var nameFlag string

func init() {
	initCmd.Flags().StringVarP(&nameFlag, "name", "n", "", "Project name (default: current directory name)")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if config already exists
	if _, err := config.FindConfig(); err == nil {
		return fmt.Errorf("cppenv.toml already exists")
	}

	// Determine project name
	projectName := nameFlag
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		projectName = filepath.Base(cwd)
	}

	// Fetch latest versions from PyPI
	fmt.Println("Fetching latest tool versions...")
	tools := make(map[string]string)
	for _, pkg := range config.DefaultTools {
		fmt.Printf("  %s: ", pkg)
		version, err := config.GetLatestVersion(pkg)
		if err != nil {
			return fmt.Errorf("failed to get version for %s: %w", pkg, err)
		}
		fmt.Printf("%s\n", version)
		tools[pkg] = version
	}

	// Create and write config
	cfg := config.CreateDefault(projectName, tools)
	if err := config.Write(cfg, config.ConfigFile); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("\nCreated %s\n", config.ConfigFile)
	fmt.Println("Run 'cppenv install' to set up the environment")
	return nil
}
