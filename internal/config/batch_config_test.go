package config_test

import (
	"testing"

	"github.com/your-org/vaultenv/internal/config"
)

func TestDefaultBatchConfig_IsDisabled(t *testing.T) {
	cfg := config.DefaultBatchConfig()
	if cfg.Enabled {
		t.Error("expected default batch config to be disabled")
	}
	if cfg.Workers != 8 {
		t.Errorf("expected default workers = 8, got %d", cfg.Workers)
	}
}

func TestBatchConfig_Validate_Disabled(t *testing.T) {
	cfg := config.BatchConfig{Enabled: false, Workers: 0}
	if err := cfg.Validate(); err != nil {
		t.Errorf("disabled config should always be valid, got: %v", err)
	}
}

func TestBatchConfig_Validate_Valid(t *testing.T) {
	cfg := config.BatchConfig{Enabled: true, Workers: 16}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestBatchConfig_Validate_ZeroWorkers(t *testing.T) {
	cfg := config.BatchConfig{Enabled: true, Workers: 0}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero workers")
	}
}

func TestBatchConfig_Validate_TooManyWorkers(t *testing.T) {
	cfg := config.BatchConfig{Enabled: true, Workers: 512}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for workers > 256")
	}
}

func TestBatchConfig_Effective_FillsZeroWorkers(t *testing.T) {
	cfg := config.BatchConfig{Enabled: true, Workers: 0}
	eff := cfg.Effective()
	if eff.Workers != 8 {
		t.Errorf("expected workers = 8 after Effective(), got %d", eff.Workers)
	}
}

func TestBatchConfig_Effective_PreservesExplicitWorkers(t *testing.T) {
	cfg := config.BatchConfig{Enabled: true, Workers: 32}
	eff := cfg.Effective()
	if eff.Workers != 32 {
		t.Errorf("expected workers = 32 after Effective(), got %d", eff.Workers)
	}
}
