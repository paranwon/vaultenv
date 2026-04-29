package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Mapping represents a single secret-to-env-var binding.
type Mapping struct {
	EnvVar string `yaml:"env"`
	Path   string `yaml:"path"`
	Key    string `yaml:"key"`
}

// Config holds the full vaultenv configuration.
type Config struct {
	Provider       string    `yaml:"provider"`
	VaultAddr      string    `yaml:"vault_addr"`
	VaultToken     string    `yaml:"vault_token"`
	VaultMount     string    `yaml:"vault_mount"`
	SSMRegion      string    `yaml:"ssm_region"`
	Mappings       []Mapping `yaml:"mappings"`
	// MultiProviders is used when Provider is "multi".
	MultiProviders []*Config `yaml:"multi_providers"`
}

// Load reads and validates a Config from the YAML file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}
	cfg, err := parse(data)
	if err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}
	return &cfg, nil
}

func validate(cfg *Config) error {
	validProviders := map[string]bool{"vault": true, "ssm": true, "multi": true}
	if !validProviders[cfg.Provider] {
		return fmt.Errorf("config: unknown provider %q", cfg.Provider)
	}
	var errs []error
	for i, m := range cfg.Mappings {
		if m.EnvVar == "" {
			errs = append(errs, fmt.Errorf("mapping[%d]: env is required", i))
		}
		if m.Path == "" {
			errs = append(errs, fmt.Errorf("mapping[%d]: path is required", i))
		}
	}
	if cfg.Provider == "multi" && len(cfg.MultiProviders) == 0 {
		errs = append(errs, errors.New("multi provider requires at least one entry in multi_providers"))
	}
	return errors.Join(errs...)
}
