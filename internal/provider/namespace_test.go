package provider_test

import (
	"context"
	"testing"

	"github.com/nicholasgasior/vaultenv/internal/provider"
)

func TestNamespaceProvider_NilInner(t *testing.T) {
	_, err := provider.NewNamespaceProvider(nil, "prod")
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}
}

func TestNamespaceProvider_EmptyNamespace(t *testing.T) {
	stub := provider.NewStaticProvider(nil)
	_, err := provider.NewNamespaceProvider(stub, "")
	if err == nil {
		t.Fatal("expected error for empty namespace")
	}
}

func TestNamespaceProvider_GetSecret_PrependNamespace(t *testing.T) {
	data := map[string]string{
		"prod/db/password": "s3cr3t",
	}
	stub := provider.NewStaticProvider(data)

	np, err := provider.NewNamespaceProvider(stub, "prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := np.GetSecret(context.Background(), "db/password", "")
	if err != nil {
		t.Fatalf("GetSecret error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %q", val)
	}
}

func TestNamespaceProvider_GetSecret_LeadingSlashNormalised(t *testing.T) {
	data := map[string]string{
		"ns/key": "val",
	}
	stub := provider.NewStaticProvider(data)
	np, _ := provider.NewNamespaceProvider(stub, "ns")

	val, err := np.GetSecret(context.Background(), "/key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "val" {
		t.Errorf("expected val, got %q", val)
	}
}

func TestNamespaceProvider_GetSecretsByPath_StripsNamespace(t *testing.T) {
	data := map[string]string{
		"staging/svc/API_KEY": "abc",
		"staging/svc/DB_PASS": "xyz",
	}
	stub := provider.NewStaticProvider(data)
	np, _ := provider.NewNamespaceProvider(stub, "staging")

	results, err := np.GetSecretsByPath(context.Background(), "svc")
	if err != nil {
		t.Fatalf("GetSecretsByPath error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for k, v := range results {
		if k == "staging/svc/API_KEY" || k == "staging/svc/DB_PASS" {
			t.Errorf("namespace not stripped from key %q", k)
		}
		_ = v
	}
}

func TestNamespaceProvider_GetSecret_NotFound(t *testing.T) {
	stub := provider.NewStaticProvider(nil)
	np, _ := provider.NewNamespaceProvider(stub, "prod")

	_, err := np.GetSecret(context.Background(), "missing", "")
	if !provider.IsNotFound(err) {
		t.Errorf("expected not-found error, got %v", err)
	}
}
