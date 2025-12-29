package environment

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const VenvDir = ".cppenv"

// GetVenvPath returns the path to the venv directory
func GetVenvPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, VenvDir, "venv")
}

// GetBinPath returns the path to the venv bin/Scripts directory
func GetBinPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(GetVenvPath(), "Scripts")
	}
	return filepath.Join(GetVenvPath(), "bin")
}

// GetPip returns the path to pip in the venv
func GetPip() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(GetBinPath(), "pip.exe")
	}
	return filepath.Join(GetBinPath(), "pip")
}

// Exists checks if the venv exists
func Exists() bool {
	_, err := os.Stat(GetVenvPath())
	return err == nil
}

// Create creates a new venv using the specified Python
func Create(pythonPath string) error {
	venvPath := GetVenvPath()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(venvPath), 0755); err != nil {
		return fmt.Errorf("failed to create .cppenv directory: %w", err)
	}

	cmd := exec.Command(pythonPath, "-m", "venv", venvPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create venv: %w", err)
	}

	return nil
}

// InstallTools installs the given requirements into the venv
func InstallTools(reqs []string) error {
	pip := GetPip()

	// Upgrade pip first
	cmd := exec.Command(pip, "install", "--upgrade", "pip")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // Ignore errors on pip upgrade

	// Install all requirements
	args := append([]string{"install"}, reqs...)
	cmd = exec.Command(pip, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install tools: %w", err)
	}

	// Create symlinks for tools that don't put binaries in bin/
	createToolSymlinks()

	return nil
}

// createToolSymlinks creates symlinks for tools that store binaries elsewhere
func createToolSymlinks() {
	binPath := GetBinPath()
	venvPath := GetVenvPath()

	// Zig is stored in site-packages/ziglang/zig
	zigSource := filepath.Join(venvPath, "lib", "python3.11", "site-packages", "ziglang", "zig")
	zigTarget := filepath.Join(binPath, "zig")

	// Check for python 3.x directories if 3.11 doesn't exist
	if _, err := os.Stat(zigSource); os.IsNotExist(err) {
		// Try to find the correct python version directory
		libPath := filepath.Join(venvPath, "lib")
		entries, _ := os.ReadDir(libPath)
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "python") {
				zigSource = filepath.Join(libPath, entry.Name(), "site-packages", "ziglang", "zig")
				if _, err := os.Stat(zigSource); err == nil {
					break
				}
			}
		}
	}

	// Create zig symlink if source exists and target doesn't
	if _, err := os.Stat(zigSource); err == nil {
		if _, err := os.Stat(zigTarget); os.IsNotExist(err) {
			os.Symlink(zigSource, zigTarget)
		}
	}
}

// GetActivatedEnv returns environment variables with PATH prepended
func GetActivatedEnv() []string {
	binPath := GetBinPath()
	pathSep := ":"
	if runtime.GOOS == "windows" {
		pathSep = ";"
	}

	env := os.Environ()
	newEnv := make([]string, 0, len(env))

	pathSet := false
	for _, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			parts := strings.SplitN(e, "=", 2)
			newPath := binPath + pathSep + parts[1]
			newEnv = append(newEnv, "PATH="+newPath)
			pathSet = true
		} else {
			newEnv = append(newEnv, e)
		}
	}

	if !pathSet {
		newEnv = append(newEnv, "PATH="+binPath)
	}

	return newEnv
}

// resolveCommand looks up a command in the venv bin directory first
func resolveCommand(name string) string {
	binPath := GetBinPath()

	// Try to find the command in the venv bin directory
	cmdPath := filepath.Join(binPath, name)
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
		// Try with .exe extension on Windows
		exePath := cmdPath + ".exe"
		if _, err := os.Stat(exePath); err == nil {
			return exePath
		}
	}
	if _, err := os.Stat(cmdPath); err == nil {
		return cmdPath
	}

	// Fall back to the original name (will be looked up in PATH)
	return name
}

// RunCommand runs a command with the activated environment
func RunCommand(args []string) int {
	if len(args) == 0 {
		return 1
	}

	// Resolve the command to full path if it exists in venv
	cmdPath := resolveCommand(args[0])

	cmd := exec.Command(cmdPath, args[1:]...)
	cmd.Env = GetActivatedEnv()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

// CreateToolchainFile generates the CMake toolchain file for Zig
func CreateToolchainFile() (string, error) {
	cwd, _ := os.Getwd()
	toolchainPath := filepath.Join(cwd, VenvDir, "zig-toolchain.cmake")

	binPath := GetBinPath()
	var zigPath string
	if runtime.GOOS == "windows" {
		zigPath = filepath.Join(binPath, "zig.exe")
	} else {
		zigPath = filepath.Join(binPath, "zig")
	}

	// Convert to forward slashes for CMake
	zigPath = strings.ReplaceAll(zigPath, "\\", "/")

	content := fmt.Sprintf(`# Generated by cppenv - do not edit manually
set(CMAKE_C_COMPILER "%s")
set(CMAKE_CXX_COMPILER "%s")
set(CMAKE_C_COMPILER_ARG1 "cc")
set(CMAKE_CXX_COMPILER_ARG1 "c++")

# Zig-specific settings
set(CMAKE_C_COMPILER_ID "Clang")
set(CMAKE_CXX_COMPILER_ID "Clang")
`, zigPath, zigPath)

	if err := os.WriteFile(toolchainPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write toolchain file: %w", err)
	}

	return toolchainPath, nil
}

// AddToGitignore adds .cppenv/ to .gitignore if not already present
func AddToGitignore() error {
	cwd, _ := os.Getwd()
	gitignorePath := filepath.Join(cwd, ".gitignore")

	// Check if .gitignore exists and already has .cppenv
	if f, err := os.Open(gitignorePath); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == ".cppenv" || line == ".cppenv/" {
				f.Close()
				return nil // Already present
			}
		}
		f.Close()
	}

	// Append to .gitignore
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Check if file is empty or ends with newline
	info, _ := f.Stat()
	if info.Size() > 0 {
		// Add newline before entry if file doesn't end with one
		f.WriteString("\n")
	}
	_, err = f.WriteString(".cppenv/\n")
	return err
}
