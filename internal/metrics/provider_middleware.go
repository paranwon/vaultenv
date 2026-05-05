package metrics

import (
	"context"
	"time"
)

// SecretProvider is the interface satisfied by all vaultenv providers.
type SecretProvider interface {
	GetSecret(ctx context.Context, path, key string) (string, error)
}

// InstrumentedProvider wraps any SecretProvider and records fetch
// latency, success, and error counts into a Registry.
type InstrumentedProvider struct {
	inner    SecretProvider
	latency  *Histogram
	success  *Counter
	errors   *Counter
}

// NewInstrumentedProvider returns a SecretProvider that emits metrics
// under the given prefix (e.g. "vault" or "ssm").
func NewInstrumentedProvider(inner SecretProvider, reg *Registry, prefix string) *InstrumentedProvider {
	return &InstrumentedProvider{
		inner:   inner,
		latency: reg.Histogram(prefix + "_fetch_latency_ms"),
		success: reg.Counter(prefix + "_fetch_success_total"),
		errors:  reg.Counter(prefix + "_fetch_error_total"),
	}
}

// GetSecret delegates to the inner provider and records metrics.
func (p *InstrumentedProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	start := time.Now()
	val, err := p.inner.GetSecret(ctx, path, key)
	p.latency.Observe(time.Since(start))
	if err != nil {
		p.errors.Inc()
		return "", err
	}
	p.success.Inc()
	return val, nil
}
