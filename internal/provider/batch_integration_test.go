package provider_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

// TestBatchProvider_ChainedWithPrefix verifies that BatchProvider composes
// correctly with PrefixProvider so that the prefix is applied transparently
// for every concurrent fetch.
func TestBatchProvider_ChainedWithPrefix(t *testing.T) {
	static, err := provider.NewStaticProvider(map[string]string{
		"prod/db#password": "s3cr3t",
		"prod/db#user":     "admin",
	})
	if err != nil {
		t.Fatalf("NewStaticProvider: %v", err)
	}

	prefixed, err := provider.NewPrefixProvider(static, "prod")
	if err != nil {
		t.Fatalf("NewPrefixProvider: %v", err)
	}

	bp, err := provider.NewBatchProvider(prefixed, 2)
	if err != nil {
		t.Fatalf("NewBatchProvider: %v", err)
	}

	refs := []provider.SecretRef{
		{Path: "db", Key: "password"},
		{Path: "db", Key: "user"},
	}

	results := bp.FetchAll(context.Background(), refs)

	want := map[string]string{
		"password": "s3cr3t",
		"user":     "admin",
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("key %q: unexpected error: %v", r.Key, r.Err)
			continue
		}
		if r.Value != want[r.Key] {
			t.Errorf("key %q: got %q, want %q", r.Key, r.Value, want[r.Key])
		}
	}
}
