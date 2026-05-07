package config

import (
	"fmt"
	"time"
)

// TTLConfig controls the advisory TTL attached to secrets by TTLProvider.
type TTLConfig struct {
	// Enabled turns the TTL wrapper on or off.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// TTL is the duration after which a fetched secret should be considered
	// stale by downstream cache layers.
	TTL time.Duration `yaml:"ttl" json:"ttl"`
}

// DefaultTTLConfig returns a disabled TTLConfig with a sensible default TTL
// that becomes active when the caller sets Enabled = true.
func DefaultTTLConfig() TTLConfig {
	return TTLConfig{
		Enabled: false,
		TTL:     5 * time.Minute,
	}
}

// Validate returns an error when the configuration is inconsistent.
func (c TTLConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.TTL <= 0 {
		return fmt.Errorf("ttl_config: ttl must be positive when enabled, got %s", c.TTL)
	}
	if c.TTL > 24*time.Hour {
		return fmt.Errorf("ttl_config: ttl %s exceeds maximum of 24h", c.TTL)
	}
	return nil
}
