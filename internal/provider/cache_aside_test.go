package provider

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheAside_NilInner(t *testing.T) {
	_, err := NewCacheAsideProvider(nil, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestCacheAside_ZeroTTL_AlwaysCallsInner(t *testing.T) {
	var calls int32
	inner := &stubCacheProvider{fn: func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "val", nil
	}}
	p, _ := NewCacheAsideProvider(inner, 0)
	for i := 0; i < 3; i++ {
		p.GetSecret(context.Background(), "secret/foo", "bar") //nolint
	}
	if calls != 3 {
		t.Fatalf("expected 3 inner calls, got %d", calls)
	}
}

func TestCacheAside_CachesGetSecret(t *testing.T) {
	var calls int32
	inner := &stubCacheProvider{fn: func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "secret-value", nil
	}}
	now := time.Now()
	p, _ := NewCacheAsideProvider(inner, time.Minute)
	p.now = func() time.Time { return now }

	v1, _ := p.GetSecret(context.Background(), "secret/foo", "bar")
	v2, _ := p.GetSecret(context.Background(), "secret/foo", "bar")

	if v1 != "secret-value" || v2 != "secret-value" {
		t.Fatalf("unexpected values: %s, %s", v1, v2)
	}
	if calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", calls)
	}
}

func TestCacheAside_ExpiresAfterTTL(t *testing.T) {
	var calls int32
	inner := &stubCacheProvider{fn: func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "v", nil
	}}
	now := time.Now()
	p, _ := NewCacheAsideProvider(inner, time.Second)
	p.now = func() time.Time { return now }

	p.GetSecret(context.Background(), "p", "k") //nolint
	now = now.Add(2 * time.Second)               // advance past TTL
	p.GetSecret(context.Background(), "p", "k") //nolint

	if calls != 2 {
		t.Fatalf("expected 2 calls after expiry, got %d", calls)
	}
}

func TestCacheAside_PropagatesError(t *testing.T) {
	sentinel := errors.New("backend down")
	inner := &stubCacheProvider{fn: func() (string, error) { return "", sentinel }}
	p, _ := NewCacheAsideProvider(inner, time.Minute)
	_, err := p.GetSecret(context.Background(), "p", "k")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestCacheAside_Invalidate(t *testing.T) {
	var calls int32
	inner := &stubCacheProvider{fn: func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "v", nil
	}}
	p, _ := NewCacheAsideProvider(inner, time.Minute)
	p.GetSecret(context.Background(), "secret/foo", "bar") //nolint
	p.Invalidate("secret/foo")
	p.GetSecret(context.Background(), "secret/foo", "bar") //nolint
	if calls != 2 {
		t.Fatalf("expected 2 calls after invalidation, got %d", calls)
	}
}

func TestCacheAside_GetSecretsByPath_Cached(t *testing.T) {
	var calls int32
	inner := &stubCacheProvider{fn: func() (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", nil
	}}
	p, _ := NewCacheAsideProvider(inner, time.Minute)
	p.GetSecretsByPath(context.Background(), "secret/foo") //nolint
	p.GetSecretsByPath(context.Background(), "secret/foo") //nolint
	if calls != 1 {
		t.Fatalf("expected 1 path call, got %d", calls)
	}
}

// stubCacheProvider satisfies Provider using a simple callback.
type stubCacheProvider struct {
	fn func() (string, error)
}

func (s *stubCacheProvider) GetSecret(_ context.Context, _, _ string) (string, error) {
	return s.fn()
}
func (s *stubCacheProvider) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	_, err := s.fn()
	if err != nil {
		return nil, err
	}
	return map[string]string{"k": "v"}, nil
}
