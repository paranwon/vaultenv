package provider

import (
	"context"
	"fmt"
)

// readOnlyProvider wraps a Provider and blocks any write-like operations
// by enforcing that path lookups are restricted to a declared allowlist.
// GetSecret and GetSecretsByPath are permitted; attempts to resolve paths
// outside the allowlist return ErrNotFound.
type readOnlyProvider struct {
	inner     Provider
	allowlist  map[string]struct{}
	hasFilter bool
}

// NewReadOnlyProvider returns a Provider that only allows access to paths
// present in allowedPaths. If allowedPaths is empty, all paths are permitted
// (the wrapper acts as a transparent pass-through with no filtering).
func NewReadOnlyProvider(inner Provider, allowedPaths []string) (Provider, error) {
	if inner == nil {
		return nil, fmt.Errorf("readOnlyProvider: inner provider must not be nil")
	}
	p := &readOnlyProvider{
		inner:    inner,
		allowlist: make(map[string]struct{}, len(allowedPaths)),
	}
	for _, path := range allowedPaths {
		if path == "" {
			return nil, fmt.Errorf("readOnlyProvider: allowedPaths must not contain empty strings")
		}
		p.allowlist[path] = struct{}{}
	}
	p.hasFilter = len(allowedPaths) > 0
	return p, nil
}

// isAllowed reports whether the given path is permitted under the current
// allowlist configuration. If no filter is active, all paths are allowed.
func (p *readOnlyProvider) isAllowed(path string) bool {
	if !p.hasFilter {
		return true
	}
	_, ok := p.allowlist[path]
	return ok
}

func (p *readOnlyProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	if !p.isAllowed(path) {
		return "", &NotFoundError{Key: path + "/" + key}
	}
	return p.inner.GetSecret(ctx, path, key)
}

func (p *readOnlyProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	if !p.isAllowed(path) {
		return nil, &NotFoundError{Key: path}
	}
	return p.inner.GetSecretsByPath(ctx, path)
}
