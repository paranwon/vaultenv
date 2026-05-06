package config

import (
	"fmt"
	"time"
)

// CircuitBreakerConfig holds circuit-breaker settings loaded from configuration.
type CircuitBreakerConfig struct {
	Enabled      bool          `yaml:"enabled"`
	MaxFailures  int           `yaml:"max_failures"`
	ResetTimeout time.Duration `yaml:"reset_timeout"`
}

// DefaultCircuitBreakerConfig returns a sensible default (disabled).
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:      false,
		MaxFailures:  5,
		ResetTimeout: 30 * time.Second,
	}
}

// Validate checks that the configuration is consistent.
func (c CircuitBreakerConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.MaxFailures <= 0 {
		return fmt.Errorf("circuit_breaker: max_failures must be > 0, got %d", c.MaxFailures)
	}
	if c.ResetTimeout <= 0 {
		return fmt.Errorf("circuit_breaker: reset_timeout must be > 0, got %s", c.ResetTimeout)
	}
	return nil
}
