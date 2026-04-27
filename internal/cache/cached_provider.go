package cache

import (
	"context"
	"fmt"
	"time"
)

// SecretProvider is the interface that wraps the basic secret retrieval methods.
type SecretProvider interface {
	GetSecret(ctx context.Context, path, key string) (string, error)
}

// CachedProvider wraps a SecretProvider with an in-memory TTL cache to
// avoid redundant calls to the upstream secrets backend.
type CachedProvider struct {
	provider SecretProvider
	cache    *Cache
}

// NewCachedProvider wraps provider with a cache using the given TTL.
// A zero or negative TTL disables caching (every call hits the provider).
func NewCachedProvider(provider SecretProvider, ttl time.Duration) *CachedProvider {
	return &CachedProvider{
		provider: provider,
		cache:    New(ttl),
	}
}

// GetSecret returns the secret value for path/key. Results are cached for the
// configured TTL duration. On a cache miss the underlying provider is called.
func (c *CachedProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	cacheKey := cacheKey(path, key)

	if val, ok := c.cache.Get(cacheKey); ok {
		return val, nil
	}

	val, err := c.provider.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}

	c.cache.Set(cacheKey, val)
	return val, nil
}

// Invalidate removes a specific path/key combination from the cache.
func (c *CachedProvider) Invalidate(path, key string) {
	c.cache.Delete(cacheKey(path, key))
}

// Flush clears all cached secrets, forcing the next request to hit the provider.
func (c *CachedProvider) Flush() {
	c.cache.Flush()
}

func cacheKey(path, key string) string {
	return fmt.Sprintf("%s#%s", path, key)
}
