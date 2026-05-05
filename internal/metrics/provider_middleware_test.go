package metrics_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultenv/internal/metrics"
)

// stubProvider is a minimal SecretProvider for testing.
type stubProvider struct {
	val string
	err error
}

func (s *stubProvider) GetSecret(_ context.Context, _, _ string) (string, error) {
	return s.val, s.err
}

func TestInstrumentedProvider_SuccessIncrements(t *testing.T) {
	reg := metrics.New()
	stub := &stubProvider{val: "s3cr3t"}
	ip := metrics.NewInstrumentedProvider(stub, reg, "vault")

	val, err := ip.GetSecret(context.Background(), "secret/app", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Fatalf("unexpected value: %s", val)
	}
	snap := reg.Snapshot()
	if snap["vault_fetch_success_total"] != 1 {
		t.Fatalf("expected success=1, got %d", snap["vault_fetch_success_total"])
	}
	if snap["vault_fetch_error_total"] != 0 {
		t.Fatalf("expected error=0, got %d", snap["vault_fetch_error_total"])
	}
}

func TestInstrumentedProvider_ErrorIncrements(t *testing.T) {
	reg := metrics.New()
	stub := &stubProvider{err: errors.New("not found")}
	ip := metrics.NewInstrumentedProvider(stub, reg, "ssm")

	_, err := ip.GetSecret(context.Background(), "/prod/db", "pass")
	if err == nil {
		t.Fatal("expected error")
	}
	snap := reg.Snapshot()
	if snap["ssm_fetch_error_total"] != 1 {
		t.Fatalf("expected error=1, got %d", snap["ssm_fetch_error_total"])
	}
	if snap["ssm_fetch_success_total"] != 0 {
		t.Fatalf("expected success=0, got %d", snap["ssm_fetch_success_total"])
	}
}

func TestInstrumentedProvider_LatencyRecorded(t *testing.T) {
	reg := metrics.New()
	stub := &stubProvider{val: "ok"}
	ip := metrics.NewInstrumentedProvider(stub, reg, "vault")

	for i := 0; i < 3; i++ {
		_, _ = ip.GetSecret(context.Background(), "path", "key")
	}
	snap := reg.Histogram("vault_fetch_latency_ms").Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 latency samples, got %d", len(snap))
	}
}
