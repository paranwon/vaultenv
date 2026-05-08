package config

import (
	"testing"
	"time"
)

func TestDefaultCacheAsideConfig_IsDisabled(t *testing.T) {
	cfg := DefaultCacheAsideConfig()
	if cfg.Enabled {
		t.Fatal("default config should be disabled")
	}
}

func TestDefaultCacheAsideConfig_HasSensibleTTL(t *testing.T) {
	cfg := DefaultCacheAsideConfig()
	if cfg.TTL <= 0 {
		t.Fatalf("default TTL should be positive, got %s", cfg.TTL)
	}
}

func TestCacheAsideConfig_Validate_Disabled(t *testing.T) {
	cfg := CacheAsideConfig{Enabled: false, TTL: 0}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("disabled config should always validate, got: %v", err)
	}
}

func TestCacheAsideConfig_Validate_ValidEnabled(t *testing.T) {
	cfg := CacheAsideConfig{Enabled: true, TTL: time.Minute}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestCacheAsideConfig_Validate_ZeroTTL(t *testing.T) {
	cfg := CacheAsideConfig{Enabled: true, TTL: 0}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for zero TTL when enabled")
	}
}

func TestCacheAsideConfig_Validate_NegativeTTL(t *testing.T) {
	cfg := CacheAsideConfig{Enabled: true, TTL: -time.Second}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative TTL when enabled")
	}
}

func TestCacheAsideConfig_Validate_DisabledIgnoresZeroTTL(t *testing.T) {
	cfg := CacheAsideConfig{Enabled: false, TTL: -99 * time.Second}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("disabled config should ignore TTL value, got: %v", err)
	}
}
