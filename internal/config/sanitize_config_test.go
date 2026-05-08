package config_test

import (
	"testing"

	"github.com/nicholasgasior/vaultenv/internal/config"
)

func TestDefaultSanitizeConfig_IsDisabled(t *testing.T) {
	cfg := config.DefaultSanitizeConfig()
	if cfg.Enabled {
		t.Error("expected default config to be disabled")
	}
	if cfg.Mode != config.SanitizeModeNone {
		t.Errorf("expected mode %q, got %q", config.SanitizeModeNone, cfg.Mode)
	}
}

func TestSanitizeConfig_Validate_Disabled(t *testing.T) {
	cfg := config.SanitizeConfig{Enabled: false, Mode: config.SanitizeModeNone}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSanitizeConfig_Validate_TrimSpace(t *testing.T) {
	cfg := config.SanitizeConfig{Enabled: true, Mode: config.SanitizeModeTrimSpace}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSanitizeConfig_Validate_StripControl(t *testing.T) {
	cfg := config.SanitizeConfig{Enabled: true, Mode: config.SanitizeModeStripControl}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSanitizeConfig_Validate_EnabledNoMode(t *testing.T) {
	cfg := config.SanitizeConfig{Enabled: true, Mode: config.SanitizeModeNone}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error when enabled with mode=none")
	}
}

func TestSanitizeConfig_Validate_UnknownMode(t *testing.T) {
	cfg := config.SanitizeConfig{Enabled: true, Mode: "base64"}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for unknown mode")
	}
}
