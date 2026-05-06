package provider_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/vaultenv/internal/provider"
)

// countingProvider records how many times GetSecretsByPath is called.
type countingProvider struct {
	calls atomic.Int32
	data  map[string]string
}

func (c *countingProvider) GetSecret(_ context.Context, path, key string) (string, error) {
	v, ok := c.data[key]
	if !ok {
		return "", provider.ErrNotFound{Key: key}
	}
	return v, nil
}

func (c *countingProvider) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	c.calls.Add(1)
	out := make(map[string]string)
	for k, v := range c.data {
		out[k] = v
	}
	return out, nil
}

func TestSnapshotProvider_CachesResults(t *testing.T) {
	inner := &countingProvider{
		data: map[string]string{"token": "abc123"},
	}
	sp := provider.NewSnapshotProvider(inner, []string{"secret/app"}, 0)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		v, err := sp.GetSecret(ctx, "secret/app", "token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != "abc123" {
			t.Fatalf("want abc123, got %s", v)
		}
	}

	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 upstream call, got %d", got)
	}
}

func TestSnapshotProvider_RespectsExpiry(t *testing.T) {
	inner := &countingProvider{
		data: map[string]string{"key": "val"},
	}
	// Very short TTL so the snapshot expires immediately.
	sp := provider.NewSnapshotProvider(inner, []string{"secret/app"}, 1*time.Millisecond)
	ctx := context.Background()

	_, _ = sp.GetSecret(ctx, "secret/app", "key")
	time.Sleep(5 * time.Millisecond)
	_, _ = sp.GetSecret(ctx, "secret/app", "key")

	if got := inner.calls.Load(); got < 2 {
		t.Fatalf("expected at least 2 upstream calls after expiry, got %d", got)
	}
}

func TestSnapshotProvider_NotFound(t *testing.T) {
	inner := &countingProvider{
		data: map[string]string{},
	}
	sp := provider.NewSnapshotProvider(inner, []string{"secret/app"}, 0)
	ctx := context.Background()

	_, err := sp.GetSecret(ctx, "secret/app", "missing")
	if !provider.IsNotFound(err) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSnapshotProvider_GetSecretsByPath_ReturnsCopy(t *testing.T) {
	inner := &countingProvider{
		data: map[string]string{"a": "1", "b": "2"},
	}
	sp := provider.NewSnapshotProvider(inner, []string{"secret/app"}, 0)
	ctx := context.Background()

	secrets, err := sp.GetSecretsByPath(ctx, "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(secrets))
	}
	// Mutating the returned map must not affect the snapshot.
	delete(secrets, "a")

	secrets2, _ := sp.GetSecretsByPath(ctx, "secret/app")
	if _, ok := secrets2["a"]; !ok {
		t.Fatal("snapshot was mutated by caller")
	}
}
