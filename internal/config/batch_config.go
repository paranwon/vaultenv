package config

import "fmt"

// BatchConfig controls the concurrent batch-fetch behaviour of vaultenv.
type BatchConfig struct {
	// Enabled turns batch fetching on or off.
	Enabled bool `yaml:"enabled" env:"VAULTENV_BATCH_ENABLED"`

	// Workers is the maximum number of concurrent secret-fetch goroutines.
	// Defaults to 8 when Enabled is true.
	Workers int `yaml:"workers" env:"VAULTENV_BATCH_WORKERS"`
}

// DefaultBatchConfig returns a safe, disabled BatchConfig.
func DefaultBatchConfig() BatchConfig {
	return BatchConfig{
		Enabled: false,
		Workers: 8,
	}
}

// Validate returns an error if the BatchConfig contains illegal values.
func (b BatchConfig) Validate() error {
	if !b.Enabled {
		return nil
	}
	if b.Workers < 1 {
		return fmt.Errorf("batch: workers must be >= 1, got %d", b.Workers)
	}
	if b.Workers > 256 {
		return fmt.Errorf("batch: workers must be <= 256, got %d", b.Workers)
	}
	return nil
}

// Effective returns b with zero-value Workers replaced by the default.
func (b BatchConfig) Effective() BatchConfig {
	if b.Workers == 0 {
		b.Workers = DefaultBatchConfig().Workers
	}
	return b
}
