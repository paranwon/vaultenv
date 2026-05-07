package config

import "fmt"

// DedupeConfig controls the deduplication middleware.
type DedupeConfig struct {
	// Enabled turns the deduplication provider wrapper on or off.
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// DefaultDedupeConfig returns a safe default with deduplication disabled.
func DefaultDedupeConfig() DedupeConfig {
	return DedupeConfig{Enabled: false}
}

// Validate returns an error if the configuration is invalid.
// DedupeConfig has no additional constraints beyond its boolean flag, but the
// method is provided for consistency with other config types.
func (c DedupeConfig) Validate() error {
	// Nothing to validate for a boolean-only config, but we keep the signature
	// so callers can range over a slice of validators uniformly.
	_ = fmt.Sprintf("dedupe.enabled=%v", c.Enabled)
	return nil
}
