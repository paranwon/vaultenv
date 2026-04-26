package provider

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// VaultProvider fetches secrets from HashiCorp Vault.
type VaultProvider struct {
	client *vaultapi.Client
}

// VaultConfig holds configuration for connecting to Vault.
type VaultConfig struct {
	Address   string
	Token     string
	Namespace string
}

// NewVaultProvider creates a new VaultProvider with the given config.
func NewVaultProvider(cfg VaultConfig) (*VaultProvider, error) {
	config := vaultapi.DefaultConfig()
	config.Address = cfg.Address

	client, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to create client: %w", err)
	}

	client.SetToken(cfg.Token)

	if cfg.Namespace != "" {
		client.SetNamespace(cfg.Namespace)
	}

	return &VaultProvider{client: client}, nil
}

// GetSecret retrieves a secret at the given path and returns key=value pairs.
// path format: "secret/data/myapp" or "secret/data/myapp#key" for a specific key.
func (v *VaultProvider) GetSecret(ctx context.Context, path string) (map[string]string, error) {
	path, keyFilter := splitPathKey(path)

	secret, err := v.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("vault: read %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("vault: no secret found at %q", path)
	}

	// KV v2 wraps data under secret.Data["data"]
	data := secret.Data
	if nested, ok := secret.Data["data"]; ok {
		if m, ok := nested.(map[string]interface{}); ok {
			data = m
		}
	}

	result := make(map[string]string)
	for k, val := range data {
		if keyFilter != "" && k != keyFilter {
			continue
		}
		if s, ok := val.(string); ok {
			result[k] = s
		}
	}

	if keyFilter != "" && len(result) == 0 {
		return nil, fmt.Errorf("vault: key %q not found at path %q", keyFilter, path)
	}

	return result, nil
}

// splitPathKey splits "path#key" into ("path", "key").
func splitPathKey(raw string) (string, string) {
	parts := strings.SplitN(raw, "#", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return raw, ""
}
