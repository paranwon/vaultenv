package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Provider represents a secret backend provider type.
type Provider string

const (
	ProviderVault Provider = "vault"
	ProviderSSM   Provider = "ssm"
)

// SecretMapping defines a single secret-to-env-var mapping.
type SecretMapping struct {
	Path   string `yaml:"path"`
	Key    string `yaml:"key,omitempty"`
	EnvVar string `yaml:"env_var"`
}

// VaultConfig holds Vault-specific configuration.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	MountPath string `yaml:"mount_path"`
}

// SSMConfig holds AWS SSM-specific configuration.
type SSMConfig struct {
	Region string `yaml:"region"`
}

// Config is the top-level vaultenv configuration.
type Config struct {
	Provider Provider        `yaml:"provider"`
	Vault    VaultConfig     `yaml:"vault,omitempty"`
	SSM      SSMConfig       `yaml:"ssm,omitempty"`
	Secrets  []SecretMapping `yaml:"secrets"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}
	return parse(data)
}

func parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Provider != ProviderVault && cfg.Provider != ProviderSSM {
		return fmt.Errorf("invalid provider %q: must be \"vault\" or \"ssm\"", cfg.Provider)
	}
	for i, s := range cfg.Secrets {
		if s.Path == "" {
			return fmt.Errorf("secrets[%d]: path must not be empty", i)
		}
		if s.EnvVar == "" {
			return fmt.Errorf("secrets[%d]: env_var must not be empty", i)
		}
	}
	return nil
}
