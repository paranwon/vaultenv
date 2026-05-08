package provider_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
)

// slowProvider delays every GetSecret call by the given duration.
type slowProvider struct {
	delay   time.Duration
	value   string
	err     error
	callCnt atomic.Int64
}

func (s *slowProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	s.callCnt.Add(1)
	select {
	case <-time.After(s.delay):
		return s.value, s.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (s *slowProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	return map[string]string{}, nil
}

func TestHedgeProvider_NilInner(t *testing.T) {
	_, err := provider.NewHedgeProvider(nil, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestHedgeProvider_ZeroDelay(t *testing.T) {
	inner, _ := provider.NewStaticProvider(map[string]string{"k": "v"})
	_, err := provider.NewHedgeProvider(inner, 0)
	if err == nil {
		t.Fatal("expected error for zero delay")
	}
}

func TestHedgeProvider_FastInner_NoHedge(t *testing.T) {
	inner, _ := provider.NewStaticProvider(map[string]string{"secret/db#pass": "hunter2"})
	h, err := provider.NewHedgeProvider(inner, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, err := h.GetSecret(context.Background(), "secret/db", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "hunter2" {
		t.Errorf("got %q, want %q", v, "hunter2")
	}
}

func TestHedgeProvider_SlowInner_IssuesHedge(t *testing.T) {
	slow := &slowProvider{delay: 80 * time.Millisecond, value: "secret"}
	h, err := provider.NewHedgeProvider(slow, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v, err := h.GetSecret(context.Background(), "path", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "secret" {
		t.Errorf("got %q, want %q", v, "secret")
	}
	if slow.callCnt.Load() < 2 {
		t.Errorf("expected hedge call, got %d total calls", slow.callCnt.Load())
	}
}

func TestHedgeProvider_ContextCancelled(t *testing.T) {
	slow := &slowProvider{delay: 500 * time.Millisecond, value: "x"}
	h, _ := provider.NewHedgeProvider(slow, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := h.GetSecret(ctx, "p", "k")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestHedgeProvider_GetSecretsByPath_Delegates(t *testing.T) {
	inner, _ := provider.NewStaticProvider(map[string]string{"a/b": "1"})
	h, _ := provider.NewHedgeProvider(inner, 10*time.Millisecond)

	got, err := h.GetSecretsByPath(context.Background(), "a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Error("expected non-nil map")
	}
}
