package provider

import (
	"context"
	"strings"
)

// PrefixProvider wraps a Provider and strips a fixed prefix from every
// secret path before delegating to the underlying provider. This is useful
// when environment-variable mappings use a short alias (e.g. "db/password")
// but the real backend path includes a team or environment prefix
// (e.g. "prod/myapp/db/password").
type PrefixProvider struct {
	inner  Provider
	prefix string
}

// NewPrefixProvider returns a PrefixProvider that prepends prefix to every
// path passed to GetSecret. prefix should NOT contain a trailing slash;
one will be added automatically.
func NewPrefixProvider(inner Provider, prefix string) *PrefixProvider {
	return &PrefixProvider{
		inner:  inner,
		prefix: strings.TrimRight(prefix, "/"),
	}
}

// GetSecret prepends the configured prefix to path and delegates to the
// wrapped provider.
func (p *PrefixProvider) GetSecret(ctx context.Context, path string) (string, error) {
	full := p.prefix + "/" + strings.TrimLeft(path, "/")
	return p.inner.GetSecret(ctx, full)
}

// GetSecretsByPath prepends the configured prefix to path and delegates to
// the wrapped provider.
func (p *PrefixProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	full := p.prefix + "/" + strings.TrimLeft(path, "/")
	return p.inner.GetSecretsByPath(ctx, full)
}
