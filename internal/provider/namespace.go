package provider

import (
	"context"
	"fmt"
	"strings"
)

// NamespaceProvider wraps an inner Provider and scopes all secret lookups
// to a fixed namespace prefix. The namespace is prepended to every path
// before delegating, and stripped from keys returned by GetSecretsByPath.
type NamespaceProvider struct {
	inner     Provider
	namespace string
}

// NewNamespaceProvider creates a NamespaceProvider that prefixes every
// path with namespace. The namespace must be non-empty and is normalised
// to have exactly one trailing slash.
func NewNamespaceProvider(inner Provider, namespace string) (*NamespaceProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("namespace provider: inner provider must not be nil")
	}
	ns := strings.TrimSpace(namespace)
	if ns == "" {
		return nil, fmt.Errorf("namespace provider: namespace must not be empty")
	}
	// Normalise: ensure exactly one trailing slash.
	ns = strings.TrimRight(ns, "/") + "/"
	return &NamespaceProvider{inner: inner, namespace: ns}, nil
}

// GetSecret prepends the configured namespace to path before delegating.
func (n *NamespaceProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	return n.inner.GetSecret(ctx, n.qualify(path), key)
}

// GetSecretsByPath prepends the namespace to path, delegates, then strips
// the namespace prefix from every key in the returned map so callers see
// unqualified keys.
func (n *NamespaceProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	results, err := n.inner.GetSecretsByPath(ctx, n.qualify(path))
	if err != nil {
		return nil, err
	}
	stripped := make(map[string]string, len(results))
	for k, v := range results {
		stripped[n.strip(k)] = v
	}
	return stripped, nil
}

func (n *NamespaceProvider) qualify(path string) string {
	clean := strings.TrimLeft(path, "/")
	return n.namespace + clean
}

func (n *NamespaceProvider) strip(key string) string {
	return strings.TrimPrefix(key, n.namespace)
}
