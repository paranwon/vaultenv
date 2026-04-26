package provider

import (
	"context"
	"fmt"
)

// SecretProvider is the common interface for all secret backends.
type SecretProvider interface {
	// GetSecret fetches secrets at the given path and returns a map of key→value.
	GetSecret(ctx context.Context, path string) (map[string]string, error)
}

// Type represents a supported secret backend.
type Type string

const (
	TypeVault Type = "vault"
	TypeSSM   Type = "ssm"
)

// Config holds generic provider configuration.
type Config struct {
	Type Type

	// Vault-specific
	VaultAddress   string
	VaultToken     string
	VaultNamespace string

	// SSM-specific
	AWSRegion string
}

// New constructs a SecretProvider from the given Config.
func New(cfg Config) (SecretProvider, error) {
	switch cfg.Type {
	case TypeVault:
		return NewVaultProvider(VaultConfig{
			Address:   cfg.VaultAddress,
			Token:     cfg.VaultToken,
			Namespace: cfg.VaultNamespace,
		})
	case TypeSSM:
		return nil, fmt.Errorf("provider: SSM not yet implemented")
	default:
		return nil, fmt.Errorf("provider: unknown type %q", cfg.Type)
	}
}
