package provider_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestTTLProvider_NilInner(t *testing.T) {
	_, err := provider.NewTTLProvider(nil, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestTTLProvider_ZeroTTL(t *testing.T) {
	stub := &stubProvider{secrets: map[string]string{"k": "v"}}
	_, err := provider.NewTTLProvider(stub, 0)
	if err == nil {
		t.Fatal("expected error for zero TTL, got nil")
	}
}

func TestTTLProvider_NegativeTTL(t *testing.T) {
	stub := &stubProvider{secrets: map[string]string{"k": "v"}}
	_, err := provider.NewTTLProvider(stub, -5*time.Second)
	if err == nil {
		t.Fatal("expected error for negative TTL, got nil")
	}
}

func TestTTLProvider_GetSecret_Delegates(t *testing.T) {
	stub := &stubProvider{secrets: map[string]string{"key": "value"}}
	p, err := provider.NewTTLProvider(stub, 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := p.GetSecret(context.Background(), "key")
	if err != nil {
		t.Fatalf("GetSecret error: %v", err)
	}
	if val != "value" {
		t.Errorf("got %q, want %q", val, "value")
	}
}

func TestTTLProvider_GetSecretsByPath_InjectsTTLKey(t *testing.T) {
	stub := &stubProvider{secrets: map[string]string{"alpha": "1", "beta": "2"}}
	p, err := provider.NewTTLProvider(stub, 2*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := p.GetSecretsByPath(context.Background(), "some/path")
	if err != nil {
		t.Fatalf("GetSecretsByPath error: %v", err)
	}
	ttlVal, ok := result[provider.TTLMetaKey]
	if !ok {
		t.Fatalf("expected %q key in result", provider.TTLMetaKey)
	}
	if ttlVal != (2 * time.Minute).String() {
		t.Errorf("got TTL %q, want %q", ttlVal, (2*time.Minute).String())
	}
}

func TestTTLProvider_GetSecretsByPath_PropagatesError(t *testing.T) {
	stub := &stubProvider{err: errors.New("backend down")}
	p, err := provider.NewTTLProvider(stub, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecretsByPath(context.Background(), "path")
	if err == nil || err.Error() != "backend down" {
		t.Errorf("expected 'backend down', got %v", err)
	}
}

func TestTTLProvider_GetSecretsByPath_NilMapHandled(t *testing.T) {
	stub := &stubProvider{secrets: nil}
	p, err := provider.NewTTLProvider(stub, 10*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	result, err := p.GetSecretsByPath(context.Background(), "path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result[provider.TTLMetaKey]; !ok {
		t.Error("TTLMetaKey missing when inner returned nil map")
	}
}
