package provider

import (
	"context"
	"time"
)

// SecretEvent describes a single secret fetch observation.
type SecretEvent struct {
	Provider  string
	Path      string
	Key       string
	Duration  time.Duration
	CacheHit  bool
	Err       error
}

// ObserveFunc is called after every GetSecret or GetSecretsByPath call.
type ObserveFunc func(SecretEvent)

// ObservingProvider wraps a Provider and calls an ObserveFunc after every
// secret fetch, enabling lightweight tracing or debugging hooks without
// coupling to a specific metrics or logging backend.
type ObservingProvider struct {
	inner   Provider
	name    string
	observe ObserveFunc
}

// NewObservingProvider returns a Provider that calls observe after every
// fetch. name labels the provider in emitted events.
func NewObservingProvider(inner Provider, name string, observe ObserveFunc) (*ObservingProvider, error) {
	if inner == nil {
		return nil, ErrNotFound{Key: "inner provider must not be nil"}
	}
	if observe == nil {
		observe = func(SecretEvent) {}
	}
	return &ObservingProvider{inner: inner, name: name, observe: observe}, nil
}

func (o *ObservingProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	start := time.Now()
	val, err := o.inner.GetSecret(ctx, path, key)
	o.observe(SecretEvent{
		Provider: o.name,
		Path:     path,
		Key:      key,
		Duration: time.Since(start),
		Err:      err,
	})
	return val, err
}

func (o *ObservingProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	start := time.Now()
	vals, err := o.inner.GetSecretsByPath(ctx, path)
	o.observe(SecretEvent{
		Provider: o.name,
		Path:     path,
		Duration: time.Since(start),
		Err:      err,
	})
	return vals, err
}
