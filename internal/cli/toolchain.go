package cli

import (
	"fmt"

	"github.com/michxymi/cppenv/internal/environment"
	"github.com/spf13/cobra"
)

var toolchainCmd = &cobra.Command{
	Use:   "toolchain",
	Short: "Regenerate the CMake toolchain file",
	Long: `Regenerates the CMake toolchain file for using Zig as the C/C++ compiler.

Use this file with CMake:
  cmake -B build -DCMAKE_TOOLCHAIN_FILE=.cppenv/zig-toolchain.cmake`,
	RunE: runToolchain,
}

func runToolchain(cmd *cobra.Command, args []string) error {
	if !environment.Exists() {
		return fmt.Errorf("environment not found, run 'cppenv install' first")
	}

	path, err := environment.CreateToolchainFile()
	if err != nil {
		return err
	}

	fmt.Printf("Toolchain file created: %s\n", path)
	fmt.Println()
	fmt.Println("Use with CMake:")
	fmt.Printf("  cmake -B build -DCMAKE_TOOLCHAIN_FILE=%s\n", path)

	return nil
}
