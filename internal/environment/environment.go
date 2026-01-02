package environment

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const VenvDir = ".cppenv"

// GetCppenvDir returns the path to the .cppenv directory
func GetCppenvDir() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, VenvDir)
}

// GetVenvPath returns the path to the venv directory
func GetVenvPath() string {
	return filepath.Join(GetCppenvDir(), "venv")
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
		// Create zig-cc and zig-c++ wrapper scripts for CMake/Conan compatibility
		// Place them in .cppenv directory, not in venv/bin
		cppenvDir := GetCppenvDir()
		os.MkdirAll(cppenvDir, 0755)
		createZigWrapperScripts(cppenvDir, zigTarget)
	}
}

// createZigWrapperScripts creates zig-cc and zig-c++ wrapper scripts
// These are needed because CMake/Conan need to invoke "zig cc" and "zig c++"
// but they expect a single executable path
// targetDir is the directory where the wrappers should be created (.cppenv directory)
func createZigWrapperScripts(targetDir, zigPath string) {
	// Convert to forward slashes for shell scripts
	zigPath = strings.ReplaceAll(zigPath, "\\", "/")

	if runtime.GOOS == "windows" {
		// Create batch files on Windows
		ccContent := fmt.Sprintf("@echo off\r\n\"%s\" cc %%*\r\n", zigPath)
		cxxContent := fmt.Sprintf("@echo off\r\n\"%s\" c++ %%*\r\n", zigPath)
		os.WriteFile(filepath.Join(targetDir, "zig-cc.bat"), []byte(ccContent), 0755)
		os.WriteFile(filepath.Join(targetDir, "zig-c++.bat"), []byte(cxxContent), 0755)
	} else {
		// Create shell scripts on Unix
		ccContent := fmt.Sprintf("#!/bin/sh\nexec \"%s\" cc \"$@\"\n", zigPath)
		cxxContent := fmt.Sprintf("#!/bin/sh\nexec \"%s\" c++ \"$@\"\n", zigPath)
		ccPath := filepath.Join(targetDir, "zig-cc")
		cxxPath := filepath.Join(targetDir, "zig-c++")
		os.WriteFile(ccPath, []byte(ccContent), 0755)
		os.WriteFile(cxxPath, []byte(cxxContent), 0755)
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
// If targetDir is empty, uses the current working directory
func CreateToolchainFile(targetDir string) (string, error) {
	if targetDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
		targetDir = cwd
	}

	// Ensure .cppenv directory exists in target directory
	cppenvDir := filepath.Join(targetDir, VenvDir)
	if err := os.MkdirAll(cppenvDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create .cppenv directory: %w", err)
	}

	toolchainPath := filepath.Join(cppenvDir, "zig_toolchain.cmake")

	// Wrappers are now in .cppenv directory, not in venv/bin
	var zigCCPath, zigCXXPath string
	if runtime.GOOS == "windows" {
		zigCCPath = filepath.Join(cppenvDir, "zig-cc.bat")
		zigCXXPath = filepath.Join(cppenvDir, "zig-c++.bat")
	} else {
		zigCCPath = filepath.Join(cppenvDir, "zig-cc")
		zigCXXPath = filepath.Join(cppenvDir, "zig-c++")
	}

	// Convert to forward slashes for CMake
	zigCCPath = strings.ReplaceAll(zigCCPath, "\\", "/")
	zigCXXPath = strings.ReplaceAll(zigCXXPath, "\\", "/")

	content := fmt.Sprintf(`# Generated by cppenv - do not edit manually
set(CMAKE_C_COMPILER "%s")
set(CMAKE_CXX_COMPILER "%s")

# Zig-specific settings
set(CMAKE_C_COMPILER_ID "Clang")
set(CMAKE_CXX_COMPILER_ID "Clang")
`, zigCCPath, zigCXXPath)

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

// CreateCMakeUserPresets creates CMakeUserPresets.json in the .cppenv directory
func CreateCMakeUserPresets() error {
	cppenvDir := GetCppenvDir()
	if err := os.MkdirAll(cppenvDir, 0755); err != nil {
		return fmt.Errorf("failed to create .cppenv directory: %w", err)
	}

	presetsPath := filepath.Join(cppenvDir, "CMakeUserPresets.json")

	presets := map[string]interface{}{
		"version": 6,
		"configurePresets": []map[string]interface{}{
			{
				"name":      "cppenv",
				"generator": "Ninja",
				"binaryDir": "${sourceDir}/build",
				"cacheVariables": map[string]interface{}{
					"CMAKE_PROJECT_TOP_LEVEL_INCLUDES": "${sourceDir}/.cppenv/conan_provider.cmake",
					"CMAKE_C_COMPILER":                 "${sourceDir}/.cppenv/zig-cc",
					"CMAKE_CXX_COMPILER":               "${sourceDir}/.cppenv/zig-c++",
				},
			},
		},
	}

	jsonData, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(presetsPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write CMakeUserPresets.json: %w", err)
	}

	return nil
}
