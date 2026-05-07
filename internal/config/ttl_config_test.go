package config_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/config"
)

func TestDefaultTTLConfig_IsDisabled(t *testing.T) {
	cfg := config.DefaultTTLConfig()
	if cfg.Enabled {
		t.Error("expected default TTL config to be disabled")
	}
	if cfg.TTL <= 0 {
		t.Errorf("expected positive default TTL, got %s", cfg.TTL)
	}
}

func TestTTLConfig_Validate_Disabled(t *testing.T) {
	cfg := config.TTLConfig{Enabled: false, TTL: 0}
	if err := cfg.Validate(); err != nil {
		t.Errorf("disabled config should always be valid, got: %v", err)
	}
}

func TestTTLConfig_Validate_Valid(t *testing.T) {
	cfg := config.TTLConfig{Enabled: true, TTL: 10 * time.Minute}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTTLConfig_Validate_ZeroTTL(t *testing.T) {
	cfg := config.TTLConfig{Enabled: true, TTL: 0}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero TTL when enabled")
	}
}

func TestTTLConfig_Validate_NegativeTTL(t *testing.T) {
	cfg := config.TTLConfig{Enabled: true, TTL: -1 * time.Second}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for negative TTL when enabled")
	}
}

func TestTTLConfig_Validate_ExceedsMax(t *testing.T) {
	cfg := config.TTLConfig{Enabled: true, TTL: 25 * time.Hour}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for TTL exceeding 24h")
	}
}

func TestTTLConfig_Validate_ExactlyMaxAllowed(t *testing.T) {
	cfg := config.TTLConfig{Enabled: true, TTL: 24 * time.Hour}
	if err := cfg.Validate(); err != nil {
		t.Errorf("24h should be allowed, got: %v", err)
	}
}
