package provider_test

import (
	"context"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
	"github.com/your-org/vaultenv/internal/ratelimit"
)

type stubProvider struct {
	calls int
	value string
	err   error
}

func (s *stubProvider) GetSecret(_ context.Context, _, _ string) (string, error) {
	s.calls++
	return s.value, s.err
}

func (s *stubProvider) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return map[string]string{"k": s.value}, nil
}

func TestRateLimitedProvider_InvalidPolicy(t *testing.T) {
	_, err := provider.NewRateLimitedProvider(&stubProvider{}, ratelimit.Policy{})
	if err == nil {
		t.Fatal("expected error for invalid policy")
	}
}

func TestRateLimitedProvider_GetSecret_Passes(t *testing.T) {
	stub := &stubProvider{value: "s3cr3t"}
	p, err := provider.NewRateLimitedProvider(stub, ratelimit.Policy{RequestsPerSecond: 100, Burst: 5})
	if err != nil {
		t.Fatal(err)
	}
	val, err := p.GetSecret(context.Background(), "path", "key")
	if err != nil {
		t.Fatal(err)
	}
	if val != "s3cr3t" {
		t.Fatalf("expected s3cr3t, got %q", val)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls)
	}
}

func TestRateLimitedProvider_GetSecretsByPath_Passes(t *testing.T) {
	stub := &stubProvider{value: "val"}
	p, err := provider.NewRateLimitedProvider(stub, ratelimit.Policy{RequestsPerSecond: 100, Burst: 5})
	if err != nil {
		t.Fatal(err)
	}
	secrets, err := p.GetSecretsByPath(context.Background(), "path")
	if err != nil {
		t.Fatal(err)
	}
	if secrets["k"] != "val" {
		t.Fatalf("unexpected secrets: %v", secrets)
	}
}

func TestRateLimitedProvider_ContextCancelled(t *testing.T) {
	stub := &stubProvider{value: "v"}
	// Burst=1 means after first call the limiter needs ~1s to refill.
	p, err := provider.NewRateLimitedProvider(stub, ratelimit.Policy{RequestsPerSecond: 1, Burst: 1})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	// Consume the single burst token.
	_, _ = p.GetSecret(ctx, "p", "k")

	ctxShort, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, err = p.GetSecret(ctxShort, "p", "k")
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}
