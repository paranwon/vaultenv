package provider

import (
	"context"
	"fmt"
	"time"
)

// TimeoutProvider wraps a Provider and enforces a per-call deadline.
type TimeoutProvider struct {
	inner   Provider
	timeout time.Duration
}

// NewTimeoutProvider returns a Provider that cancels any call exceeding d.
func NewTimeoutProvider(inner Provider, d time.Duration) (*TimeoutProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("timeout provider: inner provider must not be nil")
	}
	if d <= 0 {
		return nil, fmt.Errorf("timeout provider: timeout must be positive, got %s", d)
	}
	return &TimeoutProvider{inner: inner, timeout: d}, nil
}

// GetSecret fetches a single secret, aborting after the configured timeout.
func (t *TimeoutProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.inner.GetSecret(ctx, path, key)
}

// GetSecretsByPath fetches all secrets under path, aborting after the configured timeout.
func (t *TimeoutProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.inner.GetSecretsByPath(ctx, path)
}
