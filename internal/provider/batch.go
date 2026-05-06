package provider

import (
	"context"
	"fmt"
	"sync"
)

// BatchProvider fetches multiple secrets concurrently, bounding parallelism
// to a configurable worker count.
type BatchProvider struct {
	inner      Provider
	workers    int
}

// BatchResult holds the outcome of a single secret fetch.
type BatchResult struct {
	Key    string
	Value  string
	Err    error
}

// NewBatchProvider wraps inner with concurrent batch-fetch capability.
// workers controls the maximum number of in-flight requests; it must be >= 1.
func NewBatchProvider(inner Provider, workers int) (*BatchProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("batch: inner provider must not be nil")
	}
	if workers < 1 {
		return nil, fmt.Errorf("batch: workers must be >= 1, got %d", workers)
	}
	return &BatchProvider{inner: inner, workers: workers}, nil
}

// GetSecret delegates to the inner provider.
func (b *BatchProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	return b.inner.GetSecret(ctx, path, key)
}

// GetSecretsByPath delegates to the inner provider.
func (b *BatchProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	return b.inner.GetSecretsByPath(ctx, path)
}

// FetchAll concurrently fetches every (path, key) pair in keys using up to
// b.workers goroutines. Results are returned in the same order as keys.
func (b *BatchProvider) FetchAll(ctx context.Context, keys []SecretRef) []BatchResult {
	results := make([]BatchResult, len(keys))

	type work struct {
		idx int
		ref SecretRef
	}

	ch := make(chan work, len(keys))
	for i, k := range keys {
		ch <- work{idx: i, ref: k}
	}
	close(ch)

	var wg sync.WaitGroup
	for w := 0; w < b.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				v, err := b.inner.GetSecret(ctx, item.ref.Path, item.ref.Key)
				results[item.idx] = BatchResult{Key: item.ref.Key, Value: v, Err: err}
			}
		}()
	}
	wg.Wait()
	return results
}

// SecretRef identifies a single secret by its path and key.
type SecretRef struct {
	Path string
	Key  string
}
