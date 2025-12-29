package python

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	PythonVersion = "3.11.7"
	ReleaseDate   = "20240107"
	BaseURL       = "https://github.com/indygreg/python-build-standalone/releases/download"
)

// GetCppenvHome returns the global cppenv directory (~/.cppenv)
func GetCppenvHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cppenv")
}

// GetPythonHome returns the Python installation directory
func GetPythonHome() string {
	return filepath.Join(GetCppenvHome(), "python")
}

// GetPythonPath returns the path to the Python executable
func GetPythonPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(GetPythonHome(), "python", "python.exe")
	}
	return filepath.Join(GetPythonHome(), "python", "bin", "python3")
}

// IsInstalled checks if the managed Python is installed
func IsInstalled() bool {
	_, err := os.Stat(GetPythonPath())
	return err == nil
}

// getDownloadURL returns the download URL for the current platform
func getDownloadURL() (string, error) {
	var target string

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			target = "x86_64-unknown-linux-gnu"
		case "arm64":
			target = "aarch64-unknown-linux-gnu"
		default:
			return "", fmt.Errorf("unsupported Linux architecture: %s", runtime.GOARCH)
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			target = "x86_64-apple-darwin"
		case "arm64":
			target = "aarch64-apple-darwin"
		default:
			return "", fmt.Errorf("unsupported macOS architecture: %s", runtime.GOARCH)
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			target = "x86_64-pc-windows-msvc-shared"
		default:
			return "", fmt.Errorf("unsupported Windows architecture: %s", runtime.GOARCH)
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	filename := fmt.Sprintf("cpython-%s+%s-%s-install_only.tar.gz", PythonVersion, ReleaseDate, target)
	return fmt.Sprintf("%s/%s/%s", BaseURL, ReleaseDate, filename), nil
}

// Install downloads and extracts Python to ~/.cppenv/python/
func Install() error {
	url, err := getDownloadURL()
	if err != nil {
		return err
	}

	fmt.Printf("Downloading Python %s...\n", PythonVersion)

	pythonHome := GetPythonHome()
	if err := os.MkdirAll(pythonHome, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download Python: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Python: HTTP %d", resp.StatusCode)
	}

	if strings.HasSuffix(url, ".tar.gz") {
		if err := extractTarGz(resp.Body, pythonHome); err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
	} else if strings.HasSuffix(url, ".zip") {
		tmpFile, err := os.CreateTemp("", "python-*.zip")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())
		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			return err
		}
		tmpFile.Close()
		if err := extractZip(tmpFile.Name(), pythonHome); err != nil {
			return fmt.Errorf("failed to extract archive: %w", err)
		}
	}

	fmt.Println("Python installed successfully.")
	return nil
}

// Ensure makes sure Python is available, downloading if necessary
func Ensure() (string, error) {
	if IsInstalled() {
		return GetPythonPath(), nil
	}
	if err := Install(); err != nil {
		return "", err
	}
	return GetPythonPath(), nil
}

func extractTarGz(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
