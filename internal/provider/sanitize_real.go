package provider

import (
	"context"
	"fmt"
)

// Ensure sanitizeProvider implements Provider at compile time.
var _ Provider = (*sanitizeProviderReal)(nil)

type sanitizeProviderReal struct {
	inner    Provider
	sanitize SanitizeFunc
}

// NewSanitizeProvider returns a Provider that applies fn to every secret
// value produced by inner.
func NewSanitizeProvider(inner Provider, fn SanitizeFunc) (Provider, error) {
	if inner == nil {
		return nil, fmt.Errorf("sanitize: inner provider must not be nil")
	}
	if fn == nil {
		return nil, fmt.Errorf("sanitize: sanitize func must not be nil")
	}
	return &sanitizeProviderReal{inner: inner, sanitize: fn}, nil
}

func (p *sanitizeProviderReal) GetSecret(ctx context.Context, path, key string) (string, error) {
	val, err := p.inner.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}
	sanitized, err := p.sanitize(val)
	if err != nil {
		return "", fmt.Errorf("sanitize: path=%s key=%s: %w", path, key, err)
	}
	return sanitized, nil
}

func (p *sanitizeProviderReal) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := p.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(secrets))
	for k, v := range secrets {
		sanitized, err := p.sanitize(v)
		if err != nil {
			return nil, fmt.Errorf("sanitize: path=%s key=%s: %w", path, k, err)
		}
		result[k] = sanitized
	}
	return result, nil
}
