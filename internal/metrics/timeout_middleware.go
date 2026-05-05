package metrics

import (
	"context"
	"errors"
	"time"
)

// TimeoutAwareProvider wraps an instrumented provider and records timeout occurrences
// as a dedicated counter so dashboards can distinguish timeouts from other errors.
type TimeoutAwareProvider struct {
	inner    SecretProvider
	timeouts *Counter
}

// SecretProvider is the minimal interface consumed by middleware in this package.
type SecretProvider interface {
	GetSecret(ctx context.Context, path, key string) (string, error)
	GetSecretsByPath(ctx context.Context, path string) (map[string]string, error)
}

// NewTimeoutAwareProvider wraps inner and registers a "provider_timeouts_total" counter.
func NewTimeoutAwareProvider(inner SecretProvider, reg *Registry) *TimeoutAwareProvider {
	return &TimeoutAwareProvider{
		inner:    inner,
		timeouts: reg.Counter("provider_timeouts_total"),
	}
}

func (t *TimeoutAwareProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	start := time.Now()
	_ = start
	val, err := t.inner.GetSecret(ctx, path, key)
	if err != nil && isTimeout(err) {
		t.timeouts.Inc()
	}
	return val, err
}

func (t *TimeoutAwareProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := t.inner.GetSecretsByPath(ctx, path)
	if err != nil && isTimeout(err) {
		t.timeouts.Inc()
	}
	return secrets, err
}

func isTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}
