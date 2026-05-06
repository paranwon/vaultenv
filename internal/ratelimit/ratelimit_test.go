package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/ratelimit"
)

func TestNew_InvalidPolicy(t *testing.T) {
	_, err := ratelimit.New(ratelimit.Policy{RequestsPerSecond: 0, Burst: 1})
	if err == nil {
		t.Fatal("expected error for zero RPS")
	}
	_, err = ratelimit.New(ratelimit.Policy{RequestsPerSecond: 10, Burst: 0})
	if err == nil {
		t.Fatal("expected error for zero burst")
	}
}

func TestWait_ContextCancelled(t *testing.T) {
	// 1 RPS with burst 1: after first token consumed, next must wait ~1s.
	l, err := ratelimit.New(ratelimit.Policy{RequestsPerSecond: 1, Burst: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	// Consume the burst token.
	if err := l.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	// Now cancel immediately — should not block.
	ctxCancel, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	err = l.Wait(ctxCancel)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestWait_HighBurst_AllowsMultiple(t *testing.T) {
	l, err := ratelimit.New(ratelimit.Policy{RequestsPerSecond: 100, Burst: 10})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}
}

func TestWait_RespectsSustainedRate(t *testing.T) {
	l, err := ratelimit.New(ratelimit.Policy{RequestsPerSecond: 1000, Burst: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	start := time.Now()
	const calls = 5
	for i := 0; i < calls; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatal(err)
		}
	}
	// 5 calls at 1000 RPS should complete in well under 50ms.
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Fatalf("rate limiter too slow: %v for %d calls", elapsed, calls)
	}
}

func TestWait_AlreadyCancelledContext(t *testing.T) {
	// Verify that Wait returns immediately when given an already-cancelled context,
	// regardless of available burst tokens.
	l, err := ratelimit.New(ratelimit.Policy{RequestsPerSecond: 100, Burst: 10})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before calling Wait
	if err := l.Wait(ctx); err == nil {
		t.Fatal("expected error for already-cancelled context")
	}
}
