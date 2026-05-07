package config_test

import (
	"testing"

	"github.com/yourusername/vaultenv/internal/config"
)

func TestDefaultDedupeConfig_IsDisabled(t *testing.T) {
	cfg := config.DefaultDedupeConfig()
	if cfg.Enabled {
		t.Error("expected default dedupe config to be disabled")
	}
}

func TestDedupeConfig_Validate_Disabled(t *testing.T) {
	cfg := config.DedupeConfig{Enabled: false}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestDedupeConfig_Validate_Enabled(t *testing.T) {
	cfg := config.DedupeConfig{Enabled: true}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected validation error for enabled config: %v", err)
	}
}

func TestDedupeConfig_DefaultThenEnable(t *testing.T) {
	cfg := config.DefaultDedupeConfig()
	cfg.Enabled = true
	if !cfg.Enabled {
		t.Error("expected enabled to be true after assignment")
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}
