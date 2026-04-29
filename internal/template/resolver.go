package template

import (
	"context"
	"fmt"
	"strings"
)

// SecretProvider is the interface satisfied by vault/ssm/multi providers.
type SecretProvider interface {
	GetSecret(ctx context.Context, path string) (string, error)
}

// ProviderMap maps lowercase provider names ("vault", "ssm") to their
// implementations. It satisfies the Resolver interface used by Expand.
type ProviderMap struct {
	providers map[string]SecretProvider
}

// NewProviderMap creates a ProviderMap from the supplied name→provider pairs.
func NewProviderMap(pairs map[string]SecretProvider) *ProviderMap {
	p := make(map[string]SecretProvider, len(pairs))
	for k, v := range pairs {
		p[strings.ToLower(k)] = v
	}
	return &ProviderMap{providers: p}
}

// GetSecret implements Resolver. provider is the prefix (e.g. "vault"),
// path may contain a "#key" fragment already joined by Expand.
func (m *ProviderMap) GetSecret(ctx context.Context, provider, path string) (string, error) {
	p, ok := m.providers[strings.ToLower(provider)]
	if !ok {
		return "", fmt.Errorf("template/resolver: unknown provider %q", provider)
	}
	return p.GetSecret(ctx, path)
}

// ExpandEnv replaces secret references in every value of env, returning a new
// map with resolved values. Keys are left unchanged.
func ExpandEnv(ctx context.Context, env map[string]string, r Resolver) (map[string]string, error) {
	out := make(map[string]string, len(env))
	for k, v := range env {
		if !HasRefs(v) {
			out[k] = v
			continue
		}
		resolved, err := Expand(ctx, v, r)
		if err != nil {
			return nil, fmt.Errorf("template/resolver: env var %q: %w", k, err)
		}
		out[k] = resolved
	}
	return out, nil
}
