package config_test

import (
	"encoding/hex"
	"testing"

	"github.com/your-org/vaultenv/internal/config"
)

func TestEncryptConfig_Validate_Disabled(t *testing.T) {
	cfg := config.EncryptConfig{Enabled: false}
	if err := cfg.Validate(); err != nil {
		t.Errorf("disabled config should always be valid, got: %v", err)
	}
}

func TestEncryptConfig_Validate_ValidAESGCM(t *testing.T) {
	cfg := config.EncryptConfig{Enabled: true, Algorithm: "aes-gcm", KeyEnvVar: "MY_KEY"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("valid config: %v", err)
	}
}

func TestEncryptConfig_Validate_UnsupportedAlgorithm(t *testing.T) {
	cfg := config.EncryptConfig{Enabled: true, Algorithm: "rsa", KeyEnvVar: "MY_KEY"}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for unsupported algorithm")
	}
}

func TestEncryptConfig_Validate_EmptyKeyEnvVar(t *testing.T) {
	cfg := config.EncryptConfig{Enabled: true, Algorithm: "aes-gcm", KeyEnvVar: ""}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty key_env_var")
	}
}

func TestEncryptConfig_ResolveKey_Success(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	t.Setenv("TEST_DECRYPT_KEY", hex.EncodeToString(key))

	cfg := config.EncryptConfig{Enabled: true, Algorithm: "aes-gcm", KeyEnvVar: "TEST_DECRYPT_KEY"}
	got, err := cfg.ResolveKey()
	if err != nil {
		t.Fatalf("ResolveKey: %v", err)
	}
	if len(got) != 32 {
		t.Errorf("want 32 bytes, got %d", len(got))
	}
}

func TestEncryptConfig_ResolveKey_MissingEnv(t *testing.T) {
	cfg := config.EncryptConfig{Enabled: true, Algorithm: "aes-gcm", KeyEnvVar: "VAULTENV_NONEXISTENT_KEY_XYZ"}
	_, err := cfg.ResolveKey()
	if err == nil {
		t.Error("expected error for missing env var")
	}
}

func TestEncryptConfig_ResolveKey_WrongLength(t *testing.T) {
	t.Setenv("TEST_SHORT_KEY", hex.EncodeToString([]byte("tooshort")))
	cfg := config.EncryptConfig{Enabled: true, Algorithm: "aes-gcm", KeyEnvVar: "TEST_SHORT_KEY"}
	_, err := cfg.ResolveKey()
	if err == nil {
		t.Error("expected error for wrong key length")
	}
}

func TestDefaultEncryptConfig_IsDisabled(t *testing.T) {
	cfg := config.DefaultEncryptConfig()
	if cfg.Enabled {
		t.Error("default config should be disabled")
	}
	if cfg.Algorithm != "aes-gcm" {
		t.Errorf("default algorithm: want aes-gcm got %q", cfg.Algorithm)
	}
}
