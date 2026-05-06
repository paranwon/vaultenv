package config

import (
	"errors"
	"time"
)

// SnapshotConfig controls the behaviour of the snapshot provider wrapper.
// When Enabled is true, vaultenv fetches all secrets for the configured paths
// once and serves them from an in-memory snapshot for the lifetime of the
// process (or until TTL elapses).
type SnapshotConfig struct {
	// Enabled turns the snapshot layer on or off.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// TTL is how long the snapshot is considered fresh.
	// A zero value means the snapshot never expires.
	TTL time.Duration `yaml:"ttl" json:"ttl"`

	// Paths is the list of secret paths to eagerly fetch.
	// If empty and Enabled is true, validation returns an error.
	Paths []string `yaml:"paths" json:"paths"`
}

// DefaultSnapshotConfig returns a safe default (disabled).
func DefaultSnapshotConfig() SnapshotConfig {
	return SnapshotConfig{
		Enabled: false,
		TTL:     0,
	}
}

// Validate checks that the SnapshotConfig is self-consistent.
func (s SnapshotConfig) Validate() error {
	if !s.Enabled {
		return nil
	}
	if len(s.Paths) == 0 {
		return errors.New("snapshot: at least one path must be specified when snapshot is enabled")
	}
	if s.TTL < 0 {
		return errors.New("snapshot: ttl must be >= 0")
	}
	return nil
}
