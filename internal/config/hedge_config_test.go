package config_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/config"
)

func TestDefaultHedgeConfig_IsDisabled(t *testing.T) {
	cfg := config.DefaultHedgeConfig()
	if cfg.Enabled {
		t.Error("default hedge config should be disabled")
	}
}

func TestDefaultHedgeConfig_HasSensibleDelay(t *testing.T) {
	cfg := config.DefaultHedgeConfig()
	if cfg.Delay <= 0 {
		t.Errorf("default delay should be positive, got %v", cfg.Delay)
	}
}

func TestHedgeConfig_Validate_Disabled(t *testing.T) {
	cfg := config.HedgeConfig{Enabled: false, Delay: 0}
	if err := cfg.Validate(); err != nil {
		t.Errorf("disabled config should always be valid, got: %v", err)
	}
}

func TestHedgeConfig_Validate_ValidEnabled(t *testing.T) {
	cfg := config.HedgeConfig{Enabled: true, Delay: 25 * time.Millisecond}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestHedgeConfig_Validate_ZeroDelay(t *testing.T) {
	cfg := config.HedgeConfig{Enabled: true, Delay: 0}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero delay")
	}
}

func TestHedgeConfig_Validate_NegativeDelay(t *testing.T) {
	cfg := config.HedgeConfig{Enabled: true, Delay: -1 * time.Millisecond}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative delay")
	}
}

func TestHedgeConfig_Validate_ExceedsMaxDelay(t *testing.T) {
	cfg := config.HedgeConfig{Enabled: true, Delay: 10 * time.Second}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for delay exceeding 5s maximum")
	}
}

func TestHedgeConfig_Validate_ExactlyMaxDelay(t *testing.T) {
	cfg := config.HedgeConfig{Enabled: true, Delay: 5 * time.Second}
	if err := cfg.Validate(); err != nil {
		t.Errorf("5s should be accepted as maximum, got: %v", err)
	}
}
