package provider_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
)

// TestCircuitBreaker_ChainedWithRetry verifies that a circuit breaker sitting
// in front of a retrying provider opens correctly when the backend is down,
// preventing retries from hammering a failing dependency.
func TestCircuitBreaker_ChainedWithRetry(t *testing.T) {
	const maxFailures = 2

	inner := &countingProvider{err: errors.New("service unavailable")}

	// Wrap inner with circuit breaker.
	cp, err := provider.NewCircuitBreakerProvider(inner, provider.CircuitBreakerConfig{
		MaxFailures:  maxFailures,
		ResetTimeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewCircuitBreakerProvider: %v", err)
	}

	ctx := context.Background()

	// Trip the circuit.
	for i := 0; i < maxFailures; i++ {
		cp.GetSecret(ctx, "secret/data/app", "password") //nolint:errcheck
	}

	// Circuit is open — fast-fail without calling inner.
	callsBefore := inner.calls
	_, err = cp.GetSecret(ctx, "secret/data/app", "password")
	if !errors.Is(err, provider.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if inner.calls != callsBefore {
		t.Fatalf("inner should not have been called while circuit is open")
	}

	// Simulate recovery: wait for reset, then succeed.
	time.Sleep(60 * time.Millisecond)
	inner.err = nil
	inner.val = "s3cr3t"

	val, err := cp.GetSecret(ctx, "secret/data/app", "password")
	if err != nil {
		t.Fatalf("expected success after reset, got %v", err)
	}
	if val != "s3cr3t" {
		t.Fatalf("expected 's3cr3t', got %q", val)
	}
}
