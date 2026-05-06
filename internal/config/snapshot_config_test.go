package config

import (
	"testing"
	"time"
)

func TestDefaultSnapshotConfig_IsDisabled(t *testing.T) {
	cfg := DefaultSnapshotConfig()
	if cfg.Enabled {
		t.Fatal("default snapshot config should be disabled")
	}
	if cfg.TTL != 0 {
		t.Fatalf("expected zero TTL, got %v", cfg.TTL)
	}
}

func TestSnapshotConfig_Validate_Disabled(t *testing.T) {
	cfg := SnapshotConfig{Enabled: false}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("disabled config should always be valid, got: %v", err)
	}
}

func TestSnapshotConfig_Validate_EnabledWithPaths(t *testing.T) {
	cfg := SnapshotConfig{
		Enabled: true,
		TTL:     5 * time.Minute,
		Paths:   []string{"secret/app"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
}

func TestSnapshotConfig_Validate_EnabledNoPaths(t *testing.T) {
	cfg := SnapshotConfig{
		Enabled: true,
		Paths:   nil,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for enabled snapshot with no paths")
	}
}

func TestSnapshotConfig_Validate_NegativeTTL(t *testing.T) {
	cfg := SnapshotConfig{
		Enabled: true,
		TTL:     -1 * time.Second,
		Paths:   []string{"secret/app"},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestSnapshotConfig_Validate_ZeroTTL_IsValid(t *testing.T) {
	cfg := SnapshotConfig{
		Enabled: true,
		TTL:     0,
		Paths:   []string{"secret/app"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("zero TTL (never expire) should be valid, got: %v", err)
	}
}
