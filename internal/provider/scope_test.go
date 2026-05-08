package provider_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestScopeProvider_NilInner(t *testing.T) {
	_, err := provider.NewScopeProvider(nil, []string{"prod/"})
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}
}

func TestScopeProvider_NoScopes(t *testing.T) {
	_, err := provider.NewScopeProvider(&stubProvider{}, nil)
	if err == nil {
		t.Fatal("expected error when no scopes provided")
	}
}

func TestScopeProvider_EmptyScopeString(t *testing.T) {
	_, err := provider.NewScopeProvider(&stubProvider{}, []string{""})
	if err == nil {
		t.Fatal("expected error for empty scope string")
	}
}

func TestScopeProvider_GetSecret_Allowed(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{"prod/db/password": "s3cr3t"},
	}
	p, err := provider.NewScopeProvider(inner, []string{"prod/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sec, err := p.GetSecret(context.Background(), "prod/db/password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sec.Value != "s3cr3t" {
		t.Errorf("got %q, want %q", sec.Value, "s3cr3t")
	}
}

func TestScopeProvider_GetSecret_Blocked(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{"staging/db/password": "other"},
	}
	p, err := provider.NewScopeProvider(inner, []string{"prod/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.GetSecret(context.Background(), "staging/db/password")
	if !provider.IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestScopeProvider_GetSecret_MultipleScopes(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{
			"prod/api/key":     "prod-key",
			"shared/tls/cert": "cert-data",
		},
	}
	p, err := provider.NewScopeProvider(inner, []string{"prod/", "shared/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, path := range []string{"prod/api/key", "shared/tls/cert"} {
		if _, err := p.GetSecret(context.Background(), path); err != nil {
			t.Errorf("path %q should be allowed, got: %v", path, err)
		}
	}

	_, err = p.GetSecret(context.Background(), "dev/secret")
	if !provider.IsNotFound(err) {
		t.Errorf("expected ErrNotFound for out-of-scope path, got %v", err)
	}
}

func TestScopeProvider_GetSecretsByPath_Blocked(t *testing.T) {
	inner := &stubProvider{}
	p, err := provider.NewScopeProvider(inner, []string{"prod/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.GetSecretsByPath(context.Background(), "staging/")
	if !provider.IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestScopeProvider_NormalisesTrailingSlash(t *testing.T) {
	inner := &stubProvider{
		secrets: map[string]string{"prod/db/pass": "val"},
	}
	// scope provided WITHOUT trailing slash — should still work
	p, err := provider.NewScopeProvider(inner, []string{"prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := p.GetSecret(context.Background(), "prod/db/pass"); err != nil {
		t.Errorf("expected allowed path to succeed, got: %v", err)
	}
}
