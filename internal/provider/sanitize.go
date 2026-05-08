package provider

import (
	"fmt"
	"strings"
	"unicode"
)

// SanitizeFunc transforms a secret value into a sanitized form.
type SanitizeFunc func(value string) (string, error)

// sanitizeProvider wraps an inner Provider and applies a SanitizeFunc to
// every secret value returned by GetSecret and GetSecretsByPath.
type sanitizeProvider struct {
	inner    Provider
	sanitize SanitizeFunc
}

// NewSanitizeProvider returns a Provider that applies fn to every secret
// value produced by inner. fn must not be nil.
func NewSanitizeProvider(inner Provider, fn SanitizeFunc) (Provider, error) {
	if inner == nil {
		return nil, fmt.Errorf("sanitize: inner provider must not be nil")
	}
	if fn == nil {
		return nil, fmt.Errorf("sanitize: sanitize func must not be nil")
	}
	return &sanitizeProvider{inner: inner, sanitize: fn}, nil
}

func (p *sanitizeProvider) GetSecret(ctx interface{ Deadline() (interface{}, bool) }, path, key string) (string, error) {
	// Use the concrete context type via the standard library.
	return p.getSecret(path, key, func() (string, error) {
		return p.inner.(interface {
			GetSecret(interface{}, string, string) (string, error)
		}).GetSecret(ctx, path, key)
	})
}

func (p *sanitizeProvider) getSecret(path, key string, fetch func() (string, error)) (string, error) {
	val, err := fetch()
	if err != nil {
		return "", err
	}
	sanitized, err := p.sanitize(val)
	if err != nil {
		return "", fmt.Errorf("sanitize: path=%s key=%s: %w", path, key, err)
	}
	return sanitized, nil
}

// StripControlChars removes ASCII control characters (except tab and newline)
// from a secret value.
func StripControlChars(value string) (string, error) {
	var b strings.Builder
	for _, r := range value {
		if r == '\t' || r == '\n' || !unicode.IsControl(r) {
			b.WriteRune(r)
		}
	}
	return b.String(), nil
}

// TrimWhitespaceSanitize trims leading and trailing whitespace from a secret.
func TrimWhitespaceSanitize(value string) (string, error) {
	return strings.TrimSpace(value), nil
}
