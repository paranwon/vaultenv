package metrics_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultenv/internal/metrics"
)

type stubSecretProvider struct {
	getErr  error
	pathErr error
}

func (s *stubSecretProvider) GetSecret(_ context.Context, _, _ string) (string, error) {
	if s.getErr != nil {
		return "", s.getErr
	}
	return "secret-value", nil
}

func (s *stubSecretProvider) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	if s.pathErr != nil {
		return nil, s.pathErr
	}
	return map[string]string{"key": "val"}, nil
}

func TestTimeoutAwareProvider_NoError_NoIncrement(t *testing.T) {
	reg := metrics.New()
	p := metrics.NewTimeoutAwareProvider(&stubSecretProvider{}, reg)
	_, err := p.GetSecret(context.Background(), "path", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := reg.Snapshot()
	if snap["provider_timeouts_total"] != 0 {
		t.Fatalf("expected 0 timeouts, got %d", snap["provider_timeouts_total"])
	}
}

func TestTimeoutAwareProvider_TimeoutError_Increments(t *testing.T) {
	reg := metrics.New()
	p := metrics.NewTimeoutAwareProvider(&stubSecretProvider{getErr: context.DeadlineExceeded}, reg)
	_, err := p.GetSecret(context.Background(), "path", "key")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	snap := reg.Snapshot()
	if snap["provider_timeouts_total"] != 1 {
		t.Fatalf("expected 1 timeout, got %d", snap["provider_timeouts_total"])
	}
}

func TestTimeoutAwareProvider_NonTimeoutError_NoIncrement(t *testing.T) {
	reg := metrics.New()
	p := metrics.NewTimeoutAwareProvider(&stubSecretProvider{getErr: errors.New("network error")}, reg)
	_, _ = p.GetSecret(context.Background(), "path", "key")
	snap := reg.Snapshot()
	if snap["provider_timeouts_total"] != 0 {
		t.Fatalf("expected 0 timeouts, got %d", snap["provider_timeouts_total"])
	}
}

func TestTimeoutAwareProvider_GetSecretsByPath_Timeout(t *testing.T) {
	reg := metrics.New()
	p := metrics.NewTimeoutAwareProvider(&stubSecretProvider{pathErr: context.DeadlineExceeded}, reg)
	_, err := p.GetSecretsByPath(context.Background(), "path")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
	snap := reg.Snapshot()
	if snap["provider_timeouts_total"] != 1 {
		t.Fatalf("expected 1 timeout, got %d", snap["provider_timeouts_total"])
	}
}
