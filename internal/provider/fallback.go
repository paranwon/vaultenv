package provider

import (
	"context"
	"errors"
	"fmt"
)

// FallbackProvider wraps a primary provider and a fallback provider.
// If the primary returns a not-found error, the fallback is tried.
// Other errors from the primary are returned immediately without consulting
// the fallback.
type FallbackProvider struct {
	primary  Provider
	fallback Provider
}

// NewFallbackProvider creates a FallbackProvider that consults fallback only
// when the primary cannot find the requested secret (ErrNotFound).
func NewFallbackProvider(primary, fallback Provider) *FallbackProvider {
	if primary == nil {
		panic("fallback: primary provider must not be nil")
	}
	if fallback == nil {
		panic("fallback: fallback provider must not be nil")
	}
	return &FallbackProvider{primary: primary, fallback: fallback}
}

// GetSecret retrieves a secret from the primary provider. If the primary
// returns ErrNotFound, the fallback provider is consulted instead.
func (f *FallbackProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	val, err := f.primary.GetSecret(ctx, path, key)
	if err == nil {
		return val, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("fallback primary: %w", err)
	}
	// Primary did not find the secret — try the fallback.
	val, err = f.fallback.GetSecret(ctx, path, key)
	if err != nil {
		return "", fmt.Errorf("fallback secondary: %w", err)
	}
	return val, nil
}

// GetSecretsByPath retrieves secrets from the primary provider. If the primary
// returns ErrNotFound, the fallback provider is consulted instead.
func (f *FallbackProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	vals, err := f.primary.GetSecretsByPath(ctx, path)
	if err == nil {
		return vals, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("fallback primary: %w", err)
	}
	// Primary did not find the path — try the fallback.
	vals, err = f.fallback.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("fallback secondary: %w", err)
	}
	return vals, nil
}
