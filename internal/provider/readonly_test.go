package provider_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestReadOnlyProvider_NilInner(t *testing.T) {
	_, err := provider.NewReadOnlyProvider(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}
}

func TestReadOnlyProvider_EmptyPathInAllowlist(t *testing.T) {
	inner := &stubProvider{secrets: map[string]string{}}
	_, err := provider.NewReadOnlyProvider(inner, []string{""})
	if err == nil {
		t.Fatal("expected error for empty string in allowedPaths")
	}
}

func TestReadOnlyProvider_NoFilter_AllowsAll(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{"secret/db:password": "s3cr3t"},
	}
	p, err := provider.NewReadOnlyProvider(inner, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := p.GetSecret(context.Background(), "secret/db", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("got %q, want %q", val, "s3cr3t")
	}
}

func TestReadOnlyProvider_AllowedPath_Passes(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{"secret/app:token": "abc123"},
	}
	p, err := provider.NewReadOnlyProvider(inner, []string{"secret/app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := p.GetSecret(context.Background(), "secret/app", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "abc123" {
		t.Errorf("got %q, want %q", val, "abc123")
	}
}

func TestReadOnlyProvider_BlockedPath_ReturnsNotFound(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{"secret/other:key": "value"},
	}
	p, err := provider.NewReadOnlyProvider(inner, []string{"secret/app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecret(context.Background(), "secret/other", "key")
	if err == nil {
		t.Fatal("expected not-found error for blocked path")
	}
	if !provider.IsNotFound(err) {
		t.Errorf("expected IsNotFound error, got: %v", err)
	}
}

func TestReadOnlyProvider_GetSecretsByPath_Blocked(t *testing.T) {
	inner := &stubProvider{
		pathSecrets: map[string]map[string]string{
			"secret/other": {"key": "val"},
		},
	}
	p, err := provider.NewReadOnlyProvider(inner, []string{"secret/app"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecretsByPath(context.Background(), "secret/other")
	if !provider.IsNotFound(err) {
		t.Errorf("expected IsNotFound error, got: %v", err)
	}
}
