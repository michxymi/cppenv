package python

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetPythonHome(t *testing.T) {
	home := GetPythonHome()

	if home == "" {
		t.Fatal("GetPythonHome() returned empty string")
	}

	// Should be under user's home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get user home: %v", err)
	}

	if !strings.HasPrefix(home, userHome) {
		t.Errorf("expected path under %s, got %s", userHome, home)
	}

	// Should end with .cppenv/python
	if !strings.HasSuffix(home, filepath.Join(".cppenv", "python")) {
		t.Errorf("expected path to end with .cppenv/python, got %s", home)
	}
}

func TestGetPythonPath(t *testing.T) {
	pythonPath := GetPythonPath()

	if pythonPath == "" {
		t.Fatal("GetPythonPath() returned empty string")
	}

	// Should be under PythonHome
	home := GetPythonHome()
	if !strings.HasPrefix(pythonPath, home) {
		t.Errorf("expected path under %s, got %s", home, pythonPath)
	}

	// Check platform-specific binary name
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(pythonPath, "python.exe") {
			t.Errorf("expected path to end with python.exe on Windows, got %s", pythonPath)
		}
	} else {
		if !strings.HasSuffix(pythonPath, "python3") && !strings.HasSuffix(pythonPath, "python") {
			t.Errorf("expected path to end with python3 or python, got %s", pythonPath)
		}
	}
}

func TestGetDownloadURL(t *testing.T) {
	url, err := getDownloadURL()
	if err != nil {
		t.Fatalf("getDownloadURL() failed: %v", err)
	}

	if url == "" {
		t.Fatal("getDownloadURL() returned empty string")
	}

	// Should be a GitHub release URL
	if !strings.Contains(url, "github.com/indygreg/python-build-standalone") {
		t.Errorf("expected python-build-standalone URL, got %s", url)
	}

	// Should contain Python version
	if !strings.Contains(url, PythonVersion) {
		t.Errorf("expected URL to contain version %s, got %s", PythonVersion, url)
	}

	// Should be platform-appropriate
	switch runtime.GOOS {
	case "linux":
		if !strings.Contains(url, "linux") {
			t.Errorf("expected linux in URL on Linux, got %s", url)
		}
	case "darwin":
		if !strings.Contains(url, "apple-darwin") {
			t.Errorf("expected apple-darwin in URL on macOS, got %s", url)
		}
	case "windows":
		if !strings.Contains(url, "windows") {
			t.Errorf("expected windows in URL on Windows, got %s", url)
		}
	}
}
