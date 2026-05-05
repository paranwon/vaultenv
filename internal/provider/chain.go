package provider

import (
	"context"
	"errors"
	"fmt"
)

// ChainProvider tries each provider in order and returns the first successful
// result. Unlike MultiProvider, ChainProvider distinguishes between "not found"
// errors (continue to next) and hard errors (stop and propagate).
type ChainProvider struct {
	providers []Provider
}

// NewChainProvider creates a ChainProvider from the given ordered list of
// providers. At least one provider must be supplied.
func NewChainProvider(providers ...Provider) (*ChainProvider, error) {
	if len(providers) == 0 {
		return nil, errors.New("chain: at least one provider is required")
	}
	return &ChainProvider{providers: providers}, nil
}

// GetSecret iterates over the chain and returns the first non-not-found result.
// A hard (non-not-found) error from any provider short-circuits the chain.
func (c *ChainProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	var lastNotFound error
	for _, p := range c.providers {
		val, err := p.GetSecret(ctx, path, key)
		if err == nil {
			return val, nil
		}
		if IsNotFound(err) {
			lastNotFound = err
			continue
		}
		// Hard error — propagate immediately.
		return "", fmt.Errorf("chain: provider error: %w", err)
	}
	if lastNotFound != nil {
		return "", lastNotFound
	}
	return "", ErrNotFound(path, key)
}

// GetSecretsByPath iterates over the chain and returns the first successful
// result. Not-found responses cause the chain to advance; hard errors stop it.
func (c *ChainProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	for _, p := range c.providers {
		secrets, err := p.GetSecretsByPath(ctx, path)
		if err == nil {
			return secrets, nil
		}
		if IsNotFound(err) {
			continue
		}
		return nil, fmt.Errorf("chain: provider error: %w", err)
	}
	return nil, ErrNotFound(path, "")
}
