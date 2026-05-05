package provider

import (
	"context"
	"fmt"

	"github.com/your-org/vaultenv/internal/ratelimit"
)

// rateLimitedProvider wraps a Provider and enforces a rate limit on all calls.
type rateLimitedProvider struct {
	inner   Provider
	limiter *ratelimit.Limiter
}

// NewRateLimitedProvider wraps inner with a token-bucket rate limiter defined
// by policy. All GetSecret calls will block until a token is available or the
// context is cancelled.
func NewRateLimitedProvider(inner Provider, policy ratelimit.Policy) (Provider, error) {
	l, err := ratelimit.New(policy)
	if err != nil {
		return nil, fmt.Errorf("provider: rate limit: %w", err)
	}
	return &rateLimitedProvider{inner: inner, limiter: l}, nil
}

func (r *rateLimitedProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limit wait: %w", err)
	}
	return r.inner.GetSecret(ctx, path, key)
}

func (r *rateLimitedProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}
	return r.inner.GetSecretsByPath(ctx, path)
}
