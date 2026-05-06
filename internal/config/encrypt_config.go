package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// EncryptConfig holds client-side decryption settings.
type EncryptConfig struct {
	// Enabled indicates whether client-side decryption is active.
	Enabled bool
	// Algorithm is the decryption algorithm. Currently only "aes-gcm" is supported.
	Algorithm string
	// KeyEnvVar is the name of the environment variable that holds the hex-encoded key.
	KeyEnvVar string
}

// Validate returns an error if the EncryptConfig is invalid.
func (e EncryptConfig) Validate() error {
	if !e.Enabled {
		return nil
	}
	algo := strings.ToLower(strings.TrimSpace(e.Algorithm))
	if algo != "aes-gcm" {
		return fmt.Errorf("encrypt: unsupported algorithm %q (supported: aes-gcm)", e.Algorithm)
	}
	if strings.TrimSpace(e.KeyEnvVar) == "" {
		return fmt.Errorf("encrypt: key_env_var must not be empty when enabled")
	}
	return nil
}

// ResolveKey reads and decodes the AES key from the environment variable
// specified in KeyEnvVar. The value must be a 32-byte (256-bit) hex string.
func (e EncryptConfig) ResolveKey() ([]byte, error) {
	raw := os.Getenv(e.KeyEnvVar)
	if raw == "" {
		return nil, fmt.Errorf("encrypt: env var %q is not set or empty", e.KeyEnvVar)
	}
	key, err := hex.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return nil, fmt.Errorf("encrypt: decode key from %q: %w", e.KeyEnvVar, err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encrypt: key must be 32 bytes (256-bit), got %d", len(key))
	}
	return key, nil
}

// DefaultEncryptConfig returns an EncryptConfig with sensible defaults (disabled).
func DefaultEncryptConfig() EncryptConfig {
	return EncryptConfig{
		Enabled:   false,
		Algorithm: "aes-gcm",
		KeyEnvVar: "VAULTENV_DECRYPT_KEY",
	}
}
