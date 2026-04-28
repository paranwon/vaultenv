package provider_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
	"github.com/your-org/vaultenv/internal/retry"
)

type stubProvider struct {
	calls   int
	errors  []error
	value   string
}

func (s *stubProvider) GetSecret(_ context.Context, _, _ string) (string, error) {
	idx := s.calls
	s.calls++
	if idx < len(s.errors) {
		return "", s.errors[idx]
	}
	return s.value, nil
}

func fastRetryPolicy() retry.Policy {
	return retry.Policy{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
	}
}

func TestRetryingProvider_SuccessOnFirstCall(t *testing.T) {
	stub := &stubProvider{value: "s3cr3t"}
	p := provider.NewRetryingProvider(stub, fastRetryPolicy())

	val, err := p.GetSecret(context.Background(), "secret/app", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %q", val)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls)
	}
}

func TestRetryingProvider_RetriesTransientError(t *testing.T) {
	stub := &stubProvider{
		value:  "s3cr3t",
		errors: []error{errors.New("connection reset"), errors.New("timeout")},
	}
	p := provider.NewRetryingProvider(stub, fastRetryPolicy())

	val, err := p.GetSecret(context.Background(), "secret/app", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %q", val)
	}
	if stub.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", stub.calls)
	}
}

func TestRetryingProvider_StopsOn403(t *testing.T) {
	stub := &stubProvider{
		errors: []error{
			&provider.HTTPStatusError{StatusCode: 403, Message: "Forbidden"},
		},
	}
	p := provider.NewRetryingProvider(stub, fastRetryPolicy())

	_, err := p.GetSecret(context.Background(), "secret/app", "password")
	var httpErr *provider.HTTPStatusError
	if !errors.As(err, &httpErr) || httpErr.StatusCode != 403 {
		t.Fatalf("expected 403 HTTPStatusError, got %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call (no retry), got %d", stub.calls)
	}
}

func TestRetryingProvider_Retries429(t *testing.T) {
	stub := &stubProvider{
		value: "token",
		errors: []error{
			&provider.HTTPStatusError{StatusCode: 429, Message: "Too Many Requests"},
		},
	}
	p := provider.NewRetryingProvider(stub, fastRetryPolicy())

	val, err := p.GetSecret(context.Background(), "secret/app", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "token" {
		t.Fatalf("expected token, got %q", val)
	}
}
