package provider

import (
	"context"
	"sync"
	"time"
)

// Snapshot holds a point-in-time copy of all secrets fetched from an inner
// provider. After the first successful bulk fetch the snapshot is served
// from memory, making subsequent lookups allocation-free and offline-safe.
type snapshotProvider struct {
	inner    Provider
	ttl      time.Duration

	mu        sync.RWMutex
	data      map[string]string // path -> value
	loadedAt  time.Time
	paths     []string
}

// NewSnapshotProvider wraps inner and eagerly fetches every path in paths
// when the first secret is requested (or the snapshot has expired).
// ttl controls how long the snapshot is considered fresh; zero means forever.
func NewSnapshotProvider(inner Provider, paths []string, ttl time.Duration) Provider {
	return &snapshotProvider{
		inner: inner,
		ttl:   ttl,
		paths: paths,
	}
}

func (s *snapshotProvider) refresh(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-checked locking: another goroutine may have refreshed already.
	if s.isFresh() {
		return nil
	}

	newData := make(map[string]string)
	for _, path := range s.paths {
		secrets, err := s.inner.GetSecretsByPath(ctx, path)
		if err != nil {
			return err
		}
		for k, v := range secrets {
			newData[k] = v
		}
	}
	s.data = newData
	s.loadedAt = time.Now()
	return nil
}

// isFresh must be called with at least a read lock held.
func (s *snapshotProvider) isFresh() bool {
	if s.data == nil {
		return false
	}
	if s.ttl == 0 {
		return true
	}
	return time.Since(s.loadedAt) < s.ttl
}

func (s *snapshotProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	s.mu.RLock()
	fresh := s.isFresh()
	s.mu.RUnlock()

	if !fresh {
		if err := s.refresh(ctx); err != nil {
			return "", err
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	combined := path + "/" + key
	if v, ok := s.data[combined]; ok {
		return v, nil
	}
	if v, ok := s.data[key]; ok {
		return v, nil
	}
	return "", ErrNotFound{Key: key}
}

func (s *snapshotProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	s.mu.RLock()
	fresh := s.isFresh()
	s.mu.RUnlock()

	if !fresh {
		if err := s.refresh(ctx); err != nil {
			return nil, err
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make(map[string]string)
	for k, v := range s.data {
		out[k] = v
	}
	return out, nil
}
