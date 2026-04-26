package provider

import (
	"context"
	"fmt"
	"strings"
)

// SecretProvider is the common interface implemented by all secret backends.
type SecretProvider interface {
	// GetSecret retrieves a single secret value by its backend-specific reference.
	GetSecret(ctx context.Context, ref string) (string, error)
}

// PathProvider is an optional extension for providers that support bulk
// retrieval of secrets under a path prefix.
type PathProvider interface {
	SecretProvider
	GetSecretsByPath(ctx context.Context, path string) (map[string]string, error)
}

// BackendType enumerates the supported secret backends.
type BackendType string

const (
	BackendVault BackendType = "vault"
	BackendSSM   BackendType = "ssm"
)

// Config holds the configuration needed to instantiate a provider.
type Config struct {
	Backend BackendType
	// Vault-specific
	VaultAddr  string
	VaultToken string
	// SSM-specific
	AWSRegion string
}

// New returns a SecretProvider for the given configuration.
func New(ctx context.Context, cfg Config) (SecretProvider, error) {
	switch strings.ToLower(string(cfg.Backend)) {
	case string(BackendVault):
		return NewVaultProvider(cfg.VaultAddr, cfg.VaultToken)
	case string(BackendSSM):
		return NewSSMProvider(ctx, cfg.AWSRegion)
	default:
		return nil, fmt.Errorf("provider: unknown backend %q (supported: vault, ssm)", cfg.Backend)
	}
}
