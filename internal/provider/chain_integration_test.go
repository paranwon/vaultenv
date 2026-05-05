package provider_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

// TestChainProvider_WithPrefixProviders verifies that ChainProvider composes
// correctly with PrefixProvider, a realistic layering pattern.
func TestChainProvider_WithPrefixProviders(t *testing.T) {
	base1 := &stubChainProvider{
		secrets: map[string]string{
			"prod/app/db#password": "prod-secret",
		},
	}
	base2 := &stubChainProvider{
		secrets: map[string]string{
			"staging/app/db#password": "staging-secret",
		},
	}

	prod, err := provider.NewPrefixProvider(base1, "prod")
	if err != nil {
		t.Fatalf("NewPrefixProvider: %v", err)
	}
	staging, err := provider.NewPrefixProvider(base2, "staging")
	if err != nil {
		t.Fatalf("NewPrefixProvider: %v", err)
	}

	chain, err := provider.NewChainProvider(prod, staging)
	if err != nil {
		t.Fatalf("NewChainProvider: %v", err)
	}

	// The prod prefix provider should match first.
	val, err := chain.GetSecret(context.Background(), "app/db", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "prod-secret" {
		t.Fatalf("got %q, want %q", val, "prod-secret")
	}
}
