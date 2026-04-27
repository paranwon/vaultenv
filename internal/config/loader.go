package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// DefaultPaths lists candidate config file locations searched in order.
var DefaultPaths = []string{
	".vaultenv.yaml",
	".vaultenv.yml",
	"vaultenv.yaml",
	"vaultenv.yml",
}

// LoadDefault attempts to load a config file from the default search paths.
// It returns an error if none of the candidates exist.
func LoadDefault() (*Config, error) {
	for _, p := range DefaultPaths {
		if _, err := os.Stat(p); err == nil {
			return Load(p)
		}
	}
	return nil, fmt.Errorf("no config file found; tried: %v", DefaultPaths)
}

// LoadFromEnv loads the config file path from the VAULTENV_CONFIG environment
// variable, falling back to LoadDefault if the variable is not set.
func LoadFromEnv() (*Config, error) {
	if path := os.Getenv("VAULTENV_CONFIG"); path != "" {
		abs, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolving config path: %w", err)
		}
		return Load(abs)
	}
	return LoadDefault()
}
