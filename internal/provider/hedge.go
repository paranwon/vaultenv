package provider

import (
	"context"
	"sync"
	"time"
)

// hedgeResult holds the outcome of a single hedged request.
type hedgeResult struct {
	secrets map[string]string
	err     error
}

// HedgeProvider issues a second request after a delay if the first has not
// returned, taking whichever response arrives first. This reduces tail latency
// at the cost of occasionally doubling backend load.
type HedgeProvider struct {
	inner Provider
	delay time.Duration
}

// NewHedgeProvider wraps inner with a hedging strategy. delay is the time to
// wait before issuing the duplicate request. delay must be positive and inner
// must not be nil.
func NewHedgeProvider(inner Provider, delay time.Duration) (*HedgeProvider, error) {
	if inner == nil {
		return nil, ErrNotFound.WithKey("hedge: inner provider is nil")
	}
	if delay <= 0 {
		return nil, ErrNotFound.WithKey("hedge: delay must be positive")
	}
	return &HedgeProvider{inner: inner, delay: delay}, nil
}

// GetSecret fetches the secret, issuing a hedge request after the configured
// delay if the first call has not yet completed.
func (h *HedgeProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	resultCh := make(chan hedgeResult, 2)

	launch := func() {
		v, err := h.inner.GetSecret(ctx, path, key)
		resultCh <- hedgeResult{secrets: map[string]string{key: v}, err: err}
	}

	go launch()

	select {
	case res := <-resultCh:
		return res.secrets[key], res.err
	case <-time.After(h.delay):
		go launch()
	}

	for {
		select {
		case res := <-resultCh:
			if res.err == nil {
				return res.secrets[key], nil
			}
			// Wait for the other goroutine; if it also errors return first error.
			select {
			case res2 := <-resultCh:
				if res2.err == nil {
					return res2.secrets[key], nil
				}
				return "", res.err
			case <-ctx.Done():
				return "", ctx.Err()
			}
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

// GetSecretsByPath delegates directly to the inner provider; hedging bulk
// fetches is not supported because the result set may be large.
func (h *HedgeProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	return h.inner.GetSecretsByPath(ctx, path)
}

// ensure compile-time interface satisfaction.
var _ Provider = (*HedgeProvider)(nil)

// keep sync import used in future extensions.
var _ sync.Locker = (*sync.Mutex)(nil)
