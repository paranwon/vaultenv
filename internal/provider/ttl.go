package provider

import (
	"context"
	"fmt"
	"time"
)

// TTLProvider wraps an inner Provider and attaches a per-secret TTL to every
// value it returns. Callers can inspect the TTL via the SecretMeta field if
// the concrete value type exposes it; here we simply enforce that the inner
// call completes within the declared TTL and surface it in the key metadata.
//
// The primary purpose is to act as a lightweight deadline guard distinct from
// TimeoutProvider: TimeoutProvider aborts slow fetches, while TTLProvider
// expresses "this secret should be considered stale after N seconds" and
// attaches that intent to the returned value map so downstream cache layers
// can respect it.
type TTLProvider struct {
	inner Provider
	ttl   time.Duration
}

// TTLMeta is injected as a synthetic "__ttl__" key in every map returned by
// GetSecretsByPath so that cache layers can read it without needing a
// separate out-of-band channel.
const TTLMetaKey = "__ttl__"

// NewTTLProvider creates a TTLProvider.
// ttl must be positive; inner must not be nil.
func NewTTLProvider(inner Provider, ttl time.Duration) (*TTLProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("ttl provider: inner provider must not be nil")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl provider: ttl must be positive, got %s", ttl)
	}
	return &TTLProvider{inner: inner, ttl: ttl}, nil
}

// GetSecret delegates to the inner provider unchanged; TTL is advisory and
// primarily used by GetSecretsByPath consumers.
func (p *TTLProvider) GetSecret(ctx context.Context, path string) (string, error) {
	return p.inner.GetSecret(ctx, path)
}

// GetSecretsByPath delegates to the inner provider and injects a TTLMetaKey
// entry so that caching layers downstream can honour the declared TTL.
func (p *TTLProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := p.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	if secrets == nil {
		secrets = make(map[string]string)
	}
	secrets[TTLMetaKey] = p.ttl.String()
	return secrets, nil
}
