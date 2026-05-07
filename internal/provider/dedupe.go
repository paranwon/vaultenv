package provider

import (
	"context"
	"sync"
)

// dedupeFlight tracks an in-flight or completed call.
type dedupeFlight struct {
	wg  sync.WaitGroup
	val map[string]string
	err error
}

// DedupeProvider coalesces concurrent identical GetSecret calls into a single
// upstream request. Duplicate callers block until the first resolves and then
// share its result. GetSecretsByPath calls are not deduplicated.
type DedupeProvider struct {
	inner  Provider
	mu     sync.Mutex
	flights map[string]*dedupeFlight
}

// NewDedupeProvider wraps inner with request deduplication.
func NewDedupeProvider(inner Provider) (*DedupeProvider, error) {
	if inner == nil {
		return nil, ErrNotFound.WithKey("inner provider is nil")
	}
	return &DedupeProvider{
		inner:   inner,
		flights: make(map[string]*dedupeFlight),
	}, nil
}

// GetSecret returns the secret at path, deduplicating concurrent calls for the
// same path.
func (d *DedupeProvider) GetSecret(ctx context.Context, path string) (map[string]string, error) {
	d.mu.Lock()
	if f, ok := d.flights[path]; ok {
		d.mu.Unlock()
		// Wait for the in-flight request, but respect context cancellation.
		doneCh := make(chan struct{})
		go func() { f.wg.Wait(); close(doneCh) }()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-doneCh:
			return f.val, f.err
		}
	}

	f := &dedupeFlight{}
	f.wg.Add(1)
	d.flights[path] = f
	d.mu.Unlock()

	f.val, f.err = d.inner.GetSecret(ctx, path)
	f.wg.Done()

	d.mu.Lock()
	delete(d.flights, path)
	d.mu.Unlock()

	return f.val, f.err
}

// GetSecretsByPath delegates directly to the inner provider without deduplication.
func (d *DedupeProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	return d.inner.GetSecretsByPath(ctx, path)
}
