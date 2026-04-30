package health_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/health"
)

func TestRunAll_AllHealthy(t *testing.T) {
	reg := health.New()
	reg.Register(health.Checker{
		Name:  "vault",
		Check: func(_ context.Context) error { return nil },
	})
	reg.Register(health.Checker{
		Name:  "ssm",
		Check: func(_ context.Context) error { return nil },
	})

	results, overall := reg.RunAll(context.Background())

	if overall != health.StatusOK {
		t.Fatalf("expected overall ok, got %s", overall)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != health.StatusOK {
			t.Errorf("checker %q: expected ok, got %s: %s", r.Name, r.Status, r.Message)
		}
	}
}

func TestRunAll_OneFails(t *testing.T) {
	reg := health.New()
	reg.Register(health.Checker{
		Name:  "vault",
		Check: func(_ context.Context) error { return nil },
	})
	reg.Register(health.Checker{
		Name:  "ssm",
		Check: func(_ context.Context) error { return errors.New("connection refused") },
	})

	results, overall := reg.RunAll(context.Background())

	if overall != health.StatusDegraded {
		t.Fatalf("expected degraded, got %s", overall)
	}

	var failed *health.Result
	for i := range results {
		if results[i].Name == "ssm" {
			failed = &results[i]
		}
	}
	if failed == nil {
		t.Fatal("ssm result not found")
	}
	if failed.Status != health.StatusDown {
		t.Errorf("expected down, got %s", failed.Status)
	}
	if failed.Message == "" {
		t.Error("expected non-empty message for failed check")
	}
}

func TestRunAll_EmptyRegistry(t *testing.T) {
	reg := health.New()
	results, overall := reg.RunAll(context.Background())

	if overall != health.StatusOK {
		t.Fatalf("empty registry should be ok, got %s", overall)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestRunAll_LatencyRecorded(t *testing.T) {
	reg := health.New()
	reg.Register(health.Checker{
		Name: "slow",
		Check: func(_ context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	})

	results, _ := reg.RunAll(context.Background())
	if results[0].Latency < 10*time.Millisecond {
		t.Errorf("expected latency >= 10ms, got %s", results[0].Latency)
	}
}
