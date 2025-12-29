package config

import (
	"testing"
)

func TestGetLatestVersionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test against real PyPI - cmake is a stable, well-known package
	version, err := GetLatestVersion("cmake")
	if err != nil {
		t.Fatalf("GetLatestVersion(cmake) failed: %v", err)
	}

	if version == "" {
		t.Error("expected non-empty version string")
	}

	t.Logf("Latest cmake version: %s", version)
}

func TestGetLatestVersionNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	_, err := GetLatestVersion("this-package-definitely-does-not-exist-12345")
	if err == nil {
		t.Error("expected error for nonexistent package, got nil")
	}
}
