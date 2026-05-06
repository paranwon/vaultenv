package provider_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
)

type countingProvider struct {
	calls int
	err   error
	val   string
}

func (c *countingProvider) GetSecret(_ context.Context, _, _ string) (string, error) {
	c.calls++
	return c.val, c.err
}

func (c *countingProvider) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	return map[string]string{"k": c.val}, nil
}

func TestCircuitBreaker_NilInner(t *testing.T) {
	_, err := provider.NewCircuitBreakerProvider(nil, provider.CircuitBreakerConfig{MaxFailures: 2, ResetTimeout: time.Second})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestCircuitBreaker_InvalidConfig(t *testing.T) {
	p := &countingProvider{}
	_, err := provider.NewCircuitBreakerProvider(p, provider.CircuitBreakerConfig{MaxFailures: 0, ResetTimeout: time.Second})
	if err == nil {
		t.Fatal("expected error for MaxFailures=0")
	}
	_, err = provider.NewCircuitBreakerProvider(p, provider.CircuitBreakerConfig{MaxFailures: 1, ResetTimeout: 0})
	if err == nil {
		t.Fatal("expected error for zero ResetTimeout")
	}
}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	inner := &countingProvider{val: "secret"}
	cp, _ := provider.NewCircuitBreakerProvider(inner, provider.CircuitBreakerConfig{MaxFailures: 2, ResetTimeout: 50 * time.Millisecond})
	val, err := cp.GetSecret(context.Background(), "path", "key")
	if err != nil || val != "secret" {
		t.Fatalf("unexpected err=%v val=%s", err, val)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	inner := &countingProvider{err: errors.New("backend down")}
	cp, _ := provider.NewCircuitBreakerProvider(inner, provider.CircuitBreakerConfig{MaxFailures: 3, ResetTimeout: 100 * time.Millisecond})
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		cp.GetSecret(ctx, "p", "k") //nolint:errcheck
	}
	// Circuit should now be open — next call must fail fast.
	_, err := cp.GetSecret(ctx, "p", "k")
	if !errors.Is(err, provider.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if inner.calls != 3 {
		t.Fatalf("inner should have been called exactly 3 times, got %d", inner.calls)
	}
}

func TestCircuitBreaker_ResetsAfterTimeout(t *testing.T) {
	inner := &countingProvider{err: errors.New("down")}
	cp, _ := provider.NewCircuitBreakerProvider(inner, provider.CircuitBreakerConfig{MaxFailures: 1, ResetTimeout: 30 * time.Millisecond})
	ctx := context.Background()
	cp.GetSecret(ctx, "p", "k") //nolint:errcheck
	// Confirm open.
	_, err := cp.GetSecret(ctx, "p", "k")
	if !errors.Is(err, provider.ErrCircuitOpen) {
		t.Fatalf("expected open circuit, got %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	// After timeout, half-open — should attempt inner again.
	inner.err = nil
	inner.val = "recovered"
	val, err := cp.GetSecret(ctx, "p", "k")
	if err != nil || val != "recovered" {
		t.Fatalf("expected recovery, got err=%v val=%s", err, val)
	}
}

func TestCircuitBreaker_NotFoundDoesNotTrip(t *testing.T) {
	inner := &countingProvider{err: &provider.ErrNotFound{Key: "k"}}
	cp, _ := provider.NewCircuitBreakerProvider(inner, provider.CircuitBreakerConfig{MaxFailures: 2, ResetTimeout: time.Second})
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, err := cp.GetSecret(ctx, "p", "k")
		if errors.Is(err, provider.ErrCircuitOpen) {
			t.Fatal("circuit should not open on ErrNotFound")
		}
	}
}
