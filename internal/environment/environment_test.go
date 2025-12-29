package environment

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetVenvPath(t *testing.T) {
	// Save and restore working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origWd)

	// Create temp directory and change to it
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	venvPath := GetVenvPath()

	// Check that path ends with the expected suffix (avoids symlink issues like /var -> /private/var on macOS)
	expectedSuffix := filepath.Join(VenvDir, "venv")
	if !strings.HasSuffix(venvPath, expectedSuffix) {
		t.Errorf("expected path to end with %s, got %s", expectedSuffix, venvPath)
	}
}

func TestGetBinPath(t *testing.T) {
	// Save and restore working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origWd)

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	binPath := GetBinPath()

	// Check platform-specific path
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(binPath, "Scripts") {
			t.Errorf("expected path to end with Scripts on Windows, got %s", binPath)
		}
	} else {
		if !strings.HasSuffix(binPath, "bin") {
			t.Errorf("expected path to end with bin on Unix, got %s", binPath)
		}
	}
}

func TestGetActivatedEnv(t *testing.T) {
	// Save and restore working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origWd)

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	env := GetActivatedEnv()

	// Should have PATH set
	var pathFound bool
	binPath := GetBinPath()

	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			pathFound = true
			pathValue := strings.TrimPrefix(e, "PATH=")
			if !strings.HasPrefix(pathValue, binPath) {
				t.Errorf("expected PATH to start with %s, got %s", binPath, pathValue)
			}
		}
	}

	if !pathFound {
		t.Error("PATH not found in activated environment")
	}
}

func TestExists(t *testing.T) {
	// Save and restore working directory
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origWd)

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Should not exist initially
	if Exists() {
		t.Error("expected Exists() to return false in empty directory")
	}

	// Create venv directory structure
	venvPath := GetVenvPath()
	if err := os.MkdirAll(venvPath, 0755); err != nil {
		t.Fatalf("failed to create venv directory: %v", err)
	}

	// Should exist now
	if !Exists() {
		t.Error("expected Exists() to return true after creating venv directory")
	}
}
