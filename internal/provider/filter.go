package provider

import (
	"context"
	"strings"
)

// FilterFunc decides whether a secret key should be included.
// It receives the full path and the key name; returning true keeps the secret.
type FilterFunc func(path, key string) bool

// filterProvider wraps an inner Provider and applies a FilterFunc to every
// result returned by GetSecret and GetSecretsByPath.
type filterProvider struct {
	inner  Provider
	filter FilterFunc
}

// NewFilterProvider returns a Provider that passes results through fn.
// Only secrets for which fn returns true are surfaced to the caller.
// Secrets that are filtered out are silently dropped; if GetSecret targets a
// filtered key the provider returns ErrNotFound.
func NewFilterProvider(inner Provider, fn FilterFunc) (Provider, error) {
	if inner == nil {
		return nil, ErrNotFound // reuse sentinel; callers should guard
	}
	if fn == nil {
		return inner, nil // no-op
	}
	return &filterProvider{inner: inner, filter: fn}, nil
}

func (f *filterProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	if !f.filter(path, key) {
		return "", &NotFoundError{Key: path + "/" + key}
	}
	return f.inner.GetSecret(ctx, path, key)
}

func (f *filterProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := f.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	filtered := make(map[string]string, len(secrets))
	for k, v := range secrets {
		if f.filter(path, k) {
			filtered[k] = v
		}
	}
	return filtered, nil
}

// PrefixFilter returns a FilterFunc that allows only keys whose names start
// with the given prefix (case-sensitive).
func PrefixFilter(prefix string) FilterFunc {
	return func(_, key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

// SuffixFilter returns a FilterFunc that allows only keys whose names end
// with the given suffix (case-sensitive).
func SuffixFilter(suffix string) FilterFunc {
	return func(_, key string) bool {
		return strings.HasSuffix(key, suffix)
	}
}
