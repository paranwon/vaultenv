package config_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/config"
)

func TestDefaultCircuitBreakerConfig_IsDisabled(t *testing.T) {
	cfg := config.DefaultCircuitBreakerConfig()
	if cfg.Enabled {
		t.Error("default circuit breaker config should be disabled")
	}
	if cfg.MaxFailures <= 0 {
		t.Errorf("default MaxFailures should be > 0, got %d", cfg.MaxFailures)
	}
	if cfg.ResetTimeout <= 0 {
		t.Errorf("default ResetTimeout should be > 0, got %s", cfg.ResetTimeout)
	}
}

func TestCircuitBreakerConfig_Validate_Disabled(t *testing.T) {
	cfg := config.CircuitBreakerConfig{Enabled: false}
	if err := cfg.Validate(); err != nil {
		t.Errorf("disabled config should always be valid, got %v", err)
	}
}

func TestCircuitBreakerConfig_Validate_Valid(t *testing.T) {
	cfg := config.CircuitBreakerConfig{
		Enabled:      true,
		MaxFailures:  3,
		ResetTimeout: 10 * time.Second,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got %v", err)
	}
}

func TestCircuitBreakerConfig_Validate_ZeroMaxFailures(t *testing.T) {
	cfg := config.CircuitBreakerConfig{
		Enabled:      true,
		MaxFailures:  0,
		ResetTimeout: 10 * time.Second,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for MaxFailures=0")
	}
}

func TestCircuitBreakerConfig_Validate_ZeroResetTimeout(t *testing.T) {
	cfg := config.CircuitBreakerConfig{
		Enabled:     true,
		MaxFailures: 3,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero ResetTimeout")
	}
}

func TestCircuitBreakerConfig_Validate_NegativeMaxFailures(t *testing.T) {
	cfg := config.CircuitBreakerConfig{
		Enabled:      true,
		MaxFailures:  -1,
		ResetTimeout: time.Second,
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative MaxFailures")
	}
}
