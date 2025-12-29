package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cppenv",
	Short: "Reproducible C++ build environments",
	Long: `cppenv provides reproducible C++ build environments using pip-installable tools.

It manages Python, CMake, Ninja, Conan, Zig (for C/C++ compilation), and other
build tools in isolated per-project environments.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(toolchainCmd)
}
