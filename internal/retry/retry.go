package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// Policy defines the retry behaviour for transient errors.
type Policy struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultPolicy is a sensible default for secret provider calls.
var DefaultPolicy = Policy{
	MaxAttempts:  3,
	InitialDelay: 200 * time.Millisecond,
	MaxDelay:     5 * time.Second,
	Multiplier:   2.0,
}

// IsRetryable is a predicate that callers can use to decide whether an error
// warrants another attempt.
type IsRetryable func(err error) bool

// DefaultIsRetryable treats all non-nil errors as retryable unless they wrap
// a sentinel ErrPermanent.
func DefaultIsRetryable(err error) bool {
	return err != nil && !errors.Is(err, ErrPermanent)
}

// ErrPermanent can be wrapped around an error to signal that retrying is
// pointless (e.g. 403 Forbidden).
var ErrPermanent = errors.New("permanent error")

// Do executes fn up to policy.MaxAttempts times, backing off exponentially
// between attempts. It returns the last error if all attempts fail.
func Do(ctx context.Context, policy Policy, retryable IsRetryable, fn func() error) error {
	var err error
	delay := policy.InitialDelay

	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			next := time.Duration(math.Min(
				float64(delay)*policy.Multiplier,
				float64(policy.MaxDelay),
			))
			delay = next
		}

		err = fn()
		if !retryable(err) {
			return err
		}
	}
	return err
}
