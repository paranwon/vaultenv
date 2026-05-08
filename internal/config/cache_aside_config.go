package config

import (
	"fmt"
	"time"
)

// CacheAsideConfig controls the in-process cache-aside layer that sits in
// front of a provider to reduce backend round-trips.
type CacheAsideConfig struct {
	// Enabled turns the cache on or off.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// TTL is how long a cached entry is considered fresh.
	// Must be a positive duration when Enabled is true.
	TTL time.Duration `yaml:"ttl" json:"ttl"`
}

// DefaultCacheAsideConfig returns a sensible default (disabled).
func DefaultCacheAsideConfig() CacheAsideConfig {
	return CacheAsideConfig{
		Enabled: false,
		TTL:     30 * time.Second,
	}
}

// Validate returns an error if the configuration is inconsistent.
func (c CacheAsideConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.TTL <= 0 {
		return fmt.Errorf("cache_aside: ttl must be positive when enabled, got %s", c.TTL)
	}
	return nil
}
