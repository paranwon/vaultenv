package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestVersionedProvider_NilInner(t *testing.T) {
	_, err := provider.NewVersionedProvider(nil, "v1")
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestVersionedProvider_EmptyVersion(t *testing.T) {
	stub := &stubProvider{secrets: map[string]string{}}
	_, err := provider.NewVersionedProvider(stub, "")
	if err == nil {
		t.Fatal("expected error for empty version, got nil")
	}
}

func TestVersionedProvider_SlashInVersion(t *testing.T) {
	stub := &stubProvider{secrets: map[string]string{}}
	_, err := provider.NewVersionedProvider(stub, "v1/bad")
	if err == nil {
		t.Fatal("expected error for version containing slash, got nil")
	}
}

func TestVersionedProvider_GetSecret_AppendsSuffix(t *testing.T) {
	stub := &stubProvider{
		secrets: map[string]string{
			"secret/db/v2": "hunter2",
		},
	}
	p, err := provider.NewVersionedProvider(stub, "v2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := p.GetSecret(context.Background(), "secret/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Errorf("expected 'hunter2', got %q", val)
	}
}

func TestVersionedProvider_GetSecret_NotFound(t *testing.T) {
	stub := &stubProvider{secrets: map[string]string{}}
	p, _ := provider.NewVersionedProvider(stub, "v1")

	_, err := p.GetSecret(context.Background(), "secret/missing")
	if !provider.IsNotFound(err) {
		t.Errorf("expected not-found error, got %v", err)
	}
}

func TestVersionedProvider_GetSecretsByPath_StripsVersion(t *testing.T) {
	stub := &stubProvider{
		byPath: map[string]map[string]string{
			"secret/app/prod": {
				"db_pass/prod": "s3cr3t",
				"api_key/prod": "abc123",
			},
		},
	}
	p, _ := provider.NewVersionedProvider(stub, "prod")

	results, err := p.GetSecretsByPath(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results["db_pass"] != "s3cr3t" {
		t.Errorf("expected db_pass='s3cr3t', got %q", results["db_pass"])
	}
	if results["api_key"] != "abc123" {
		t.Errorf("expected api_key='abc123', got %q", results["api_key"])
	}
}

func TestVersionedProvider_GetSecretsByPath_PropagatesError(t *testing.T) {
	sentinel := errors.New("backend unavailable")
	stub := &stubProvider{pathErr: sentinel}
	p, _ := provider.NewVersionedProvider(stub, "v1")

	_, err := p.GetSecretsByPath(context.Background(), "secret/app")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
