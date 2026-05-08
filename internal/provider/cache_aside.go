package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CacheAsideProvider wraps an inner Provider with a simple in-process
// write-through / read-aside cache keyed on (path, key).
// Unlike the generic cache.CachedProvider it is self-contained and
// supports per-entry TTL eviction without an external cache.Cache.
type CacheAsideProvider struct {
	inner   Provider
	ttl     time.Duration
	mu      sync.Mutex
	entries map[string]cacheAsideEntry
	now     func() time.Time
}

type cacheAsideEntry struct {
	value     map[string]string
	expiresAt time.Time
}

// NewCacheAsideProvider creates a CacheAsideProvider with the given TTL.
// A zero or negative TTL disables caching (every call hits the inner provider).
func NewCacheAsideProvider(inner Provider, ttl time.Duration) (*CacheAsideProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("cache_aside: inner provider must not be nil")
	}
	return &CacheAsideProvider{
		inner:   inner,
		ttl:     ttl,
		entries: make(map[string]cacheAsideEntry),
		now:     time.Now,
	}, nil
}

func (c *CacheAsideProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	if c.ttl <= 0 {
		return c.inner.GetSecret(ctx, path, key)
	}
	ck := path + "\x00" + key
	c.mu.Lock()
	if e, ok := c.entries[ck]; ok && c.now().Before(e.expiresAt) {
		v := e.value[key]
		c.mu.Unlock()
		return v, nil
	}
	c.mu.Unlock()

	v, err := c.inner.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.entries[ck] = cacheAsideEntry{
		value:     map[string]string{key: v},
		expiresAt: c.now().Add(c.ttl),
	}
	c.mu.Unlock()
	return v, nil
}

func (c *CacheAsideProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	if c.ttl <= 0 {
		return c.inner.GetSecretsByPath(ctx, path)
	}
	ck := path + "\x00*"
	c.mu.Lock()
	if e, ok := c.entries[ck]; ok && c.now().Before(e.expiresAt) {
		out := make(map[string]string, len(e.value))
		for k, v := range e.value {
			out[k] = v
		}
		c.mu.Unlock()
		return out, nil
	}
	c.mu.Unlock()

	secrets, err := c.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.entries[ck] = cacheAsideEntry{
		value:     secrets,
		expiresAt: c.now().Add(c.ttl),
	}
	c.mu.Unlock()
	return secrets, nil
}

// Invalidate removes all cached entries for the given path.
func (c *CacheAsideProvider) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.entries {
		if len(k) >= len(path) && k[:len(path)] == path {
			delete(c.entries, k)
		}
	}
}
