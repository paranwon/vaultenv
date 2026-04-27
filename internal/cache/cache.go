package cache

import (
	"sync"
	"time"
)

// Entry holds a cached secret value and its expiry time.
type Entry struct {
	Value     string
	ExpiresAt time.Time
}

// Cache is a simple in-memory TTL cache for secret values.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]Entry
	ttl     time.Duration
	nowFunc func() time.Time
}

// New creates a new Cache with the given TTL duration.
func New(ttl time.Duration) *Cache {
	return &Cache{
		items:   make(map[string]Entry),
		ttl:     ttl,
		nowFunc: time.Now,
	}
}

// Get retrieves a value from the cache. Returns the value and true if found
// and not expired, otherwise returns empty string and false.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[key]
	if !ok {
		return "", false
	}
	if c.nowFunc().After(entry.ExpiresAt) {
		return "", false
	}
	return entry.Value, true
}

// Set stores a value in the cache under the given key.
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Entry{
		Value:     value,
		ExpiresAt: c.nowFunc().Add(c.ttl),
	}
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Flush removes all entries from the cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]Entry)
}

// Len returns the number of entries currently in the cache (including expired).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}
