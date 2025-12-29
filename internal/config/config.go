package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const ConfigFile = "cppenv.toml"

var DefaultTools = []string{
	"ziglang",
	"cmake",
	"ninja",
	"conan",
	"clang-tools",
}

type Config struct {
	Project ProjectConfig     `toml:"project"`
	Tools   map[string]string `toml:"tools"`
	Scripts map[string]string `toml:"scripts"`
}

type ProjectConfig struct {
	Name string `toml:"name"`
}

// FindConfig searches for cppenv.toml in the current directory
func FindConfig() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(cwd, ConfigFile)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

// Load reads and parses a cppenv.toml file
func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	if cfg.Tools == nil {
		cfg.Tools = make(map[string]string)
	}
	if cfg.Scripts == nil {
		cfg.Scripts = make(map[string]string)
	}
	return &cfg, nil
}

// CreateDefault creates a new Config with default tools
func CreateDefault(projectName string, tools map[string]string) *Config {
	return &Config{
		Project: ProjectConfig{Name: projectName},
		Tools:   tools,
		Scripts: make(map[string]string),
	}
}

// Write saves a Config to a TOML file
func Write(cfg *Config, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	return encoder.Encode(cfg)
}

// GetRequirements returns pip install requirements (e.g., ["ziglang==0.11.0", ...])
func (c *Config) GetRequirements() []string {
	reqs := make([]string, 0, len(c.Tools))
	for pkg, version := range c.Tools {
		reqs = append(reqs, pkg+"=="+version)
	}
	return reqs
}
