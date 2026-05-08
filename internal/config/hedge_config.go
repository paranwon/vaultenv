package config

import (
	"fmt"
	"time"
)

// HedgeConfig controls the hedging middleware for provider calls.
type HedgeConfig struct {
	// Enabled turns hedging on or off.
	Enabled bool `yaml:"enabled" env:"VAULTENV_HEDGE_ENABLED"`

	// Delay is how long to wait before issuing a duplicate (hedge) request.
	// Must be positive when Enabled is true.
	Delay time.Duration `yaml:"delay" env:"VAULTENV_HEDGE_DELAY"`
}

// DefaultHedgeConfig returns a safe, disabled configuration.
func DefaultHedgeConfig() HedgeConfig {
	return HedgeConfig{
		Enabled: false,
		Delay:   20 * time.Millisecond,
	}
}

// Validate checks that the configuration is internally consistent.
func (c HedgeConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Delay <= 0 {
		return fmt.Errorf("hedge: delay must be positive, got %v", c.Delay)
	}
	if c.Delay > 5*time.Second {
		return fmt.Errorf("hedge: delay %v exceeds maximum of 5s", c.Delay)
	}
	return nil
}
