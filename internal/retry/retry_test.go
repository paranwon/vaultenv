package retry_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/retry"
)

func fastPolicy() retry.Policy {
	return retry.Policy{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
	}
}

func TestDo_SucceedsFirstAttempt(t *testing.T) {
	calls := 0
	err := retry.Do(context.Background(), fastPolicy(), retry.DefaultIsRetryable, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesTransientError(t *testing.T) {
	calls := 0
	err := retry.Do(context.Background(), fastPolicy(), retry.DefaultIsRetryable, func() error {
		calls++
		if calls < 3 {
			return fmt.Errorf("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_StopsOnPermanentError(t *testing.T) {
	calls := 0
	perm := fmt.Errorf("access denied: %w", retry.ErrPermanent)
	err := retry.Do(context.Background(), fastPolicy(), retry.DefaultIsRetryable, func() error {
		calls++
		return perm
	})
	if !errors.Is(err, retry.ErrPermanent) {
		t.Fatalf("expected permanent error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_ExhaustsMaxAttempts(t *testing.T) {
	calls := 0
	sentinel := errors.New("always fails")
	err := retry.Do(context.Background(), fastPolicy(), retry.DefaultIsRetryable, func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_RespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := retry.Do(ctx, fastPolicy(), retry.DefaultIsRetryable, func() error {
		calls++
		cancel()
		return errors.New("transient")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
