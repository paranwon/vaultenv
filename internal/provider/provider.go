package provider

import (
	"context"
	"fmt"

	"github.com/your-org/vaultenv/internal/config"
)

// Provider is the interface all secret backends must implement.
type Provider interface {
	// GetSecret retrieves a single secret value by path and key.
	GetSecret(ctx context.Context, path, key string) (string, error)
	// GetSecretsByPath retrieves all key/value pairs under a path.
	GetSecretsByPath(ctx context.Context, path string) (map[string]string, error)
}

// New constructs the appropriate Provider implementation based on the config.
// It returns an error if the provider type is unknown or misconfigured.
func New(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "vault":
		return NewVaultProvider(cfg)
	case "ssm":
		return NewSSMProvider(cfg)
	case "multi":
		var providers []Provider
		for _, sub := range cfg.MultiProviders {
			p, err := New(sub)
			if err != nil {
				return nil, fmt.Errorf("multi provider: failed to init sub-provider %q: %w", sub.Provider, err)
			}
			providers = append(providers, p)
		}
		return NewMultiProvider(providers...)
	default:
		return nil, fmt.Errorf("unknown provider %q", cfg.Provider)
	}
}
