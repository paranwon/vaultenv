package provider_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
)

// Verify that TimeoutProvider composes correctly with RetryingProvider:
// the outer timeout should abort retries mid-flight.
func TestTimeoutProvider_AbortsRetries(t *testing.T) {
	calls := 0
	stub := &callCountProvider{
		fn: func() (string, error) {
			calls++
			// simulate a transient error that would normally be retried
			return "", errors.New("connection reset")
		},
		delay: 30 * time.Millisecond,
	}

	retrying, err := provider.NewRetryingProvider(stub, fastRetryPolicy())
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// timeout shorter than one retry cycle
	withTimeout, err := provider.NewTimeoutProvider(retrying, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err = withTimeout.GetSecret(context.Background(), "secret/data/app", "key")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	if calls > 2 {
		t.Fatalf("timeout should have aborted retries early, but got %d calls", calls)
	}
}

// callCountProvider records call count and optionally delays.
type callCountProvider struct {
	fn    func() (string, error)
	delay time.Duration
}

func (c *callCountProvider) GetSecret(ctx context.Context, _, _ string) (string, error) {
	select {
	case <-time.After(c.delay):
	case <-ctx.Done():
		return "", ctx.Err()
	}
	return c.fn()
}

func (c *callCountProvider) GetSecretsByPath(ctx context.Context, _ string) (map[string]string, error) {
	select {
	case <-time.After(c.delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	_, err := c.fn()
	return nil, err
}
