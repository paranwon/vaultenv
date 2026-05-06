package provider

import (
	"context"
	"strings"
)

// StaticProvider serves secrets from an in-memory map. It is primarily useful
// for testing and for injecting well-known default values without a remote
// backend.
type StaticProvider struct {
	secrets map[string]string
}

// NewStaticProvider returns a Provider backed by the supplied key/value map.
// Keys are stored and looked up case-insensitively (lowercased).
func NewStaticProvider(secrets map[string]string) *StaticProvider {
	norm := make(map[string]string, len(secrets))
	for k, v := range secrets {
		norm[strings.ToLower(k)] = v
	}
	return &StaticProvider{secrets: norm}
}

// GetSecret returns the value for the given path/key pair.
// path and key are joined as "path/key" and looked up case-insensitively.
// Returns ErrNotFound when the key is absent.
func (p *StaticProvider) GetSecret(_ context.Context, path, key string) (string, error) {
	lookup := strings.ToLower(path + "/" + key)
	if val, ok := p.secrets[lookup]; ok {
		return val, nil
	}
	return "", ErrNotFound{Key: path + "/" + key}
}

// GetSecretsByPath returns all secrets whose key starts with the given path
// prefix. The returned map uses the original casing of the stored keys.
func (p *StaticProvider) GetSecretsByPath(_ context.Context, path string) (map[string]string, error) {
	prefix := strings.ToLower(strings.TrimSuffix(path, "/") + "/")
	result := make(map[string]string)
	for k, v := range p.secrets {
		if strings.HasPrefix(k, prefix) {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil, ErrNotFound{Key: path}
	}
	return result, nil
}
