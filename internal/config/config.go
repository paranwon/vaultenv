// Package config handles loading and validating vaultenv configuration.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Mapping is a single secret-to-env-var binding.
type Mapping struct {
	EnvVar string `yaml:"env"`
	Path   string `yaml:"path"`
	Key    string `yaml:"key"`
}

// RateLimit holds optional rate-limiting configuration for providers.
type RateLimit struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	Burst             int     `yaml:"burst"`
}

// Config is the top-level vaultenv configuration.
type Config struct {
	Provider  string    `yaml:"provider"`
	Mappings  []Mapping `yaml:"mappings"`
	RateLimit RateLimit `yaml:"rate_limit"`
}

// Load reads and validates a Config from the YAML file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	return parse(data)
}

func parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(cfg *Config) error {
	cfg.Provider = strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch cfg.Provider {
	case "vault", "ssm":
	default:
		return fmt.Errorf("config: unsupported provider %q (want vault or ssm)", cfg.Provider)
	}
	for i, m := range cfg.Mappings {
		if m.EnvVar == "" {
			return fmt.Errorf("config: mapping[%d]: env must not be empty", i)
		}
		if m.Path == "" {
			return fmt.Errorf("config: mapping[%d]: path must not be empty", i)
		}
	}
	if rl := cfg.RateLimit; (rl.RequestsPerSecond != 0 || rl.Burst != 0) &&
		(rl.RequestsPerSecond <= 0 || rl.Burst <= 0) {
		return fmt.Errorf("config: rate_limit: both requests_per_second and burst must be positive")
	}
	return nil
}
