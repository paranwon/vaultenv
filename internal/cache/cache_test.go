package cache

import (
	"testing"
	"time"
)

func fixedTime(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestCache_SetAndGet(t *testing.T) {
	c := New(5 * time.Minute)
	c.Set("secret/foo", "bar")

	val, ok := c.Get("secret/foo")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if val != "bar" {
		t.Fatalf("expected 'bar', got %q", val)
	}
}

func TestCache_Miss(t *testing.T) {
	c := New(5 * time.Minute)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCache_Expiry(t *testing.T) {
	now := time.Now()
	c := New(1 * time.Minute)
	c.nowFunc = fixedTime(now)
	c.Set("secret/foo", "bar")

	// Advance time past TTL
	c.nowFunc = fixedTime(now.Add(2 * time.Minute))

	_, ok := c.Get("secret/foo")
	if ok {
		t.Fatal("expected cache miss after expiry")
	}
}

func TestCache_NotExpiredBeforeTTL(t *testing.T) {
	now := time.Now()
	c := New(5 * time.Minute)
	c.nowFunc = fixedTime(now)
	c.Set("secret/foo", "bar")

	c.nowFunc = fixedTime(now.Add(4 * time.Minute))

	val, ok := c.Get("secret/foo")
	if !ok {
		t.Fatal("expected cache hit before expiry")
	}
	if val != "bar" {
		t.Fatalf("expected 'bar', got %q", val)
	}
}

func TestCache_Delete(t *testing.T) {
	c := New(5 * time.Minute)
	c.Set("secret/foo", "bar")
	c.Delete("secret/foo")

	_, ok := c.Get("secret/foo")
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}

func TestCache_Flush(t *testing.T) {
	c := New(5 * time.Minute)
	c.Set("a", "1")
	c.Set("b", "2")
	c.Flush()

	if c.Len() != 0 {
		t.Fatalf("expected empty cache after flush, got %d entries", c.Len())
	}
}

func TestCache_Overwrite(t *testing.T) {
	c := New(5 * time.Minute)
	c.Set("key", "old")
	c.Set("key", "new")

	val, ok := c.Get("key")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if val != "new" {
		t.Fatalf("expected 'new', got %q", val)
	}
}
