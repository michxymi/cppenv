package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp directory
	dir := t.TempDir()
	configPath := filepath.Join(dir, "cppenv.toml")

	content := `
[project]
name = "test-project"

[tools]
cmake = "3.28.1"
ninja = "1.11.1.1"

[scripts]
build = "cmake --build build"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got '%s'", cfg.Project.Name)
	}

	if cfg.Tools["cmake"] != "3.28.1" {
		t.Errorf("expected cmake version '3.28.1', got '%s'", cfg.Tools["cmake"])
	}

	if cfg.Tools["ninja"] != "1.11.1.1" {
		t.Errorf("expected ninja version '1.11.1.1', got '%s'", cfg.Tools["ninja"])
	}

	if cfg.Scripts["build"] != "cmake --build build" {
		t.Errorf("expected build script 'cmake --build build', got '%s'", cfg.Scripts["build"])
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/cppenv.toml")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestGetRequirements(t *testing.T) {
	cfg := &Config{
		Tools: map[string]string{
			"cmake": "3.28.1",
			"ninja": "1.11.1.1",
		},
	}

	reqs := cfg.GetRequirements()

	if len(reqs) != 2 {
		t.Fatalf("expected 2 requirements, got %d", len(reqs))
	}

	// Check that both requirements are present (order may vary due to map)
	reqMap := make(map[string]bool)
	for _, req := range reqs {
		reqMap[req] = true
	}

	if !reqMap["cmake==3.28.1"] {
		t.Error("expected 'cmake==3.28.1' in requirements")
	}
	if !reqMap["ninja==1.11.1.1"] {
		t.Error("expected 'ninja==1.11.1.1' in requirements")
	}
}

func TestCreateDefault(t *testing.T) {
	tools := map[string]string{
		"cmake": "3.28.1",
	}

	cfg := CreateDefault("my-project", tools)

	if cfg.Project.Name != "my-project" {
		t.Errorf("expected project name 'my-project', got '%s'", cfg.Project.Name)
	}

	if cfg.Tools["cmake"] != "3.28.1" {
		t.Errorf("expected cmake version '3.28.1', got '%s'", cfg.Tools["cmake"])
	}

	if cfg.Scripts == nil {
		t.Error("expected Scripts map to be initialized")
	}
}

func TestWriteAndLoad(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "cppenv.toml")

	original := &Config{
		Project: ProjectConfig{Name: "roundtrip-test"},
		Tools: map[string]string{
			"cmake": "3.28.1",
		},
		Scripts: map[string]string{
			"build": "make",
		},
	}

	if err := Write(original, configPath); err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded.Project.Name != original.Project.Name {
		t.Errorf("project name mismatch: expected '%s', got '%s'",
			original.Project.Name, loaded.Project.Name)
	}

	if loaded.Tools["cmake"] != original.Tools["cmake"] {
		t.Errorf("cmake version mismatch: expected '%s', got '%s'",
			original.Tools["cmake"], loaded.Tools["cmake"])
	}
}
