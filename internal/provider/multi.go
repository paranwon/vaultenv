package provider

import (
	"context"
	"errors"
	"fmt"
)

// MultiProvider tries each provider in order and returns the first successful result.
// If all providers fail, it returns a combined error.
type MultiProvider struct {
	providers []Provider
}

// NewMultiProvider creates a MultiProvider from the given list of providers.
// At least one provider must be supplied.
func NewMultiProvider(providers ...Provider) (*MultiProvider, error) {
	if len(providers) == 0 {
		return nil, errors.New("multi provider: at least one provider is required")
	}
	return &MultiProvider{providers: providers}, nil
}

// GetSecret iterates through the providers and returns the first successful secret.
// Errors from each provider are collected and returned if all providers fail.
func (m *MultiProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	var errs []error
	for _, p := range m.providers {
		val, err := p.GetSecret(ctx, path, key)
		if err == nil {
			return val, nil
		}
		errs = append(errs, err)
	}
	return "", fmt.Errorf("multi provider: all providers failed: %w", errors.Join(errs...))
}

// GetSecretsByPath iterates through the providers and returns the first successful map.
// Errors from each provider are collected and returned if all providers fail.
func (m *MultiProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	var errs []error
	for _, p := range m.providers {
		vals, err := p.GetSecretsByPath(ctx, path)
		if err == nil {
			return vals, nil
		}
		errs = append(errs, err)
	}
	return nil, fmt.Errorf("multi provider: all providers failed: %w", errors.Join(errs...))
}
