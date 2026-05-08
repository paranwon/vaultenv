package provider

import (
	"context"
	"fmt"
	"strings"
)

// ScopeProvider wraps an inner Provider and enforces that all secret paths
// fall within one of the declared allowed scopes (path prefixes). Requests
// outside the allowed scopes are rejected with ErrNotFound so callers cannot
// accidentally reach secrets they are not entitled to.
type ScopeProvider struct {
	inner  Provider
	scopes []string
}

// NewScopeProvider returns a ScopeProvider that only allows access to paths
// that begin with one of the provided scope prefixes.
//
// Rules:
//   - inner must not be nil.
//   - At least one scope must be provided.
//   - Each scope is normalised: leading/trailing whitespace is stripped and a
//     trailing "/" is appended if absent so that prefix matching is unambiguous.
func NewScopeProvider(inner Provider, scopes []string) (*ScopeProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("scope: inner provider must not be nil")
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("scope: at least one scope prefix is required")
	}

	normalised := make([]string, 0, len(scopes))
	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil, fmt.Errorf("scope: scope prefix must not be empty")
		}
		if !strings.HasSuffix(s, "/") {
			s += "/"
		}
		normalised = append(normalised, s)
	}

	return &ScopeProvider{inner: inner, scopes: normalised}, nil
}

// GetSecret returns the secret at path only if path falls within an allowed
// scope. Otherwise ErrNotFound is returned.
func (p *ScopeProvider) GetSecret(ctx context.Context, path string) (*Secret, error) {
	if !p.allowed(path) {
		return nil, ErrNotFound{Key: path}
	}
	return p.inner.GetSecret(ctx, path)
}

// GetSecretsByPath returns secrets under the given path only if it falls
// within an allowed scope.
func (p *ScopeProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	if !p.allowed(path) {
		return nil, ErrNotFound{Key: path}
	}
	return p.inner.GetSecretsByPath(ctx, path)
}

// allowed reports whether path starts with at least one of the declared scopes.
func (p *ScopeProvider) allowed(path string) bool {
	norm := path
	if !strings.HasSuffix(norm, "/") {
		norm += "/"
	}
	for _, s := range p.scopes {
		if strings.HasPrefix(norm, s) {
			return true
		}
	}
	return false
}
