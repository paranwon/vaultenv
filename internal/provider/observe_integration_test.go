package provider_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/nicholasgasior/vaultenv/internal/provider"
)

// TestObservingProvider_ChainedWithPrefix verifies that ObservingProvider
// correctly wraps a PrefixProvider and emits events with the rewritten path.
func TestObservingProvider_ChainedWithPrefix(t *testing.T) {
	static, _ := provider.NewStaticProvider(map[string]string{
		"prod/secret/db:password": "s3cr3t",
	})
	prefixed, err := provider.NewPrefixProvider(static, "prod/")
	if err != nil {
		t.Fatalf("NewPrefixProvider: %v", err)
	}

	var calls int64
	op, err := provider.NewObservingProvider(prefixed, "prefixed-static", func(e provider.SecretEvent) {
		atomic.AddInt64(&calls, 1)
		if e.Provider != "prefixed-static" {
			t.Errorf("Provider label: got %q", e.Provider)
		}
	})
	if err != nil {
		t.Fatalf("NewObservingProvider: %v", err)
	}

	val, err := op.GetSecret(context.Background(), "secret/db", "password")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("value: got %q, want %q", val, "s3cr3t")
	}
	if atomic.LoadInt64(&calls) != 1 {
		t.Errorf("expected 1 observe call, got %d", calls)
	}
}
