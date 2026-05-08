package provider

import (
	"context"
	"fmt"
)

// CloneProvider wraps an inner Provider and deep-copies every secret value
// before returning it, ensuring that callers cannot accidentally mutate shared
// state held inside the inner provider's cache.
type CloneProvider struct {
	inner Provider
}

// NewCloneProvider returns a CloneProvider that wraps inner.
// It returns an error when inner is nil.
func NewCloneProvider(inner Provider) (*CloneProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("cloneprovider: inner provider must not be nil")
	}
	return &CloneProvider{inner: inner}, nil
}

// GetSecret delegates to the inner provider and returns a freshly allocated
// copy of the secret value so the caller owns the string exclusively.
func (c *CloneProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	v, err := c.inner.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}
	return clone(v), nil
}

// GetSecretsByPath delegates to the inner provider and returns a new map
// whose values are independent copies of the originals.
func (c *CloneProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := c.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[clone(k)] = clone(v)
	}
	return out, nil
}

// clone returns a freshly allocated copy of s.
func clone(s string) string {
	if s == "" {
		return ""
	}
	b := make([]byte, len(s))
	copy(b, s)
	return string(b)
}
