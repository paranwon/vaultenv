package config

import (
	"fmt"
	"strings"
)

// NamespaceConfig controls the NamespaceProvider middleware.
type NamespaceConfig struct {
	// Enabled activates namespace scoping when true.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Namespace is the prefix prepended to every secret path.
	// Must be non-empty when Enabled is true.
	Namespace string `yaml:"namespace" json:"namespace"`
}

// DefaultNamespaceConfig returns a disabled NamespaceConfig.
func DefaultNamespaceConfig() NamespaceConfig {
	return NamespaceConfig{
		Enabled:   false,
		Namespace: "",
	}
}

// Validate returns an error if the configuration is inconsistent.
func (c NamespaceConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	ns := strings.TrimSpace(c.Namespace)
	if ns == "" {
		return fmt.Errorf("namespace config: namespace must not be empty when enabled")
	}
	return nil
}
