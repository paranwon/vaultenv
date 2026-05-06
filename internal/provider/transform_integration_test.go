package provider_test

import (
	"context"
	"strings"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

// TestTransformProvider_ChainedWithPrefix verifies that TransformProvider
// composes correctly with PrefixProvider.
func TestTransformProvider_ChainedWithPrefix(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{
			"db_password": "  s3cr3t  ",
			"api_key":     "  abc123  ",
		},
	}

	// Wrap with prefix then trim whitespace.
	prefixed := provider.NewPrefixProvider(inner, "/prod")
	trimmed := provider.NewTransformProvider(prefixed, provider.TrimSpaceTransform)

	secrets, err := trimmed.GetSecretsByPath(context.Background(), "/secrets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for k, v := range secrets {
		if strings.TrimSpace(v) != v {
			t.Errorf("key %q still has whitespace: %q", k, v)
		}
	}
}
