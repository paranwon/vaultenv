package provider_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
)

type slowProvider struct {
	delay time.Duration
}

func (s *slowProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	select {
	case <-time.After(s.delay):
		return "value", nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (s *slowProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	select {
	case <-time.After(s.delay):
		return map[string]string{"k": "v"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestTimeoutProvider_InvalidArgs(t *testing.T) {
	_, err := provider.NewTimeoutProvider(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}

	_, err = provider.NewTimeoutProvider(&slowProvider{}, 0)
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestTimeoutProvider_GetSecret_Succeeds(t *testing.T) {
	p, _ := provider.NewTimeoutProvider(&slowProvider{delay: 1 * time.Millisecond}, 100*time.Millisecond)
	val, err := p.GetSecret(context.Background(), "secret/data/app", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}
}

func TestTimeoutProvider_GetSecret_TimesOut(t *testing.T) {
	p, _ := provider.NewTimeoutProvider(&slowProvider{delay: 200 * time.Millisecond}, 10*time.Millisecond)
	_, err := p.GetSecret(context.Background(), "secret/data/app", "token")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTimeoutProvider_GetSecretsByPath_TimesOut(t *testing.T) {
	p, _ := provider.NewTimeoutProvider(&slowProvider{delay: 200 * time.Millisecond}, 10*time.Millisecond)
	_, err := p.GetSecretsByPath(context.Background(), "secret/data/app")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTimeoutProvider_RespectsParentCancellation(t *testing.T) {
	p, _ := provider.NewTimeoutProvider(&slowProvider{delay: 500 * time.Millisecond}, 10*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.GetSecret(ctx, "secret/data/app", "token")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected Canceled, got %v", err)
	}
}
