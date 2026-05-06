package provider_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestFilterProvider_NilInner_ReturnsError(t *testing.T) {
	_, err := provider.NewFilterProvider(nil, provider.PrefixFilter("APP_"))
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}
}

func TestFilterProvider_NilFilter_ReturnsInner(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"key": "val"})
	p, err := provider.NewFilterProvider(inner, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != inner {
		t.Error("expected the inner provider to be returned unchanged")
	}
}

func TestFilterProvider_GetSecret_Allowed(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"APP_TOKEN": "secret"})
	p, _ := provider.NewFilterProvider(inner, provider.PrefixFilter("APP_"))

	val, err := p.GetSecret(context.Background(), "myapp", "APP_TOKEN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret" {
		t.Errorf("got %q, want %q", val, "secret")
	}
}

func TestFilterProvider_GetSecret_Blocked(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"DB_PASS": "hunter2"})
	p, _ := provider.NewFilterProvider(inner, provider.PrefixFilter("APP_"))

	_, err := p.GetSecret(context.Background(), "myapp", "DB_PASS")
	if !provider.IsNotFound(err) {
		t.Fatalf("expected not-found, got %v", err)
	}
}

func TestFilterProvider_GetSecretsByPath_FiltersKeys(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{
		"APP_ID":  "1",
		"APP_KEY": "abc",
		"DB_URL":  "postgres://",
	})
	p, _ := provider.NewFilterProvider(inner, provider.PrefixFilter("APP_"))

	secrets, err := p.GetSecretsByPath(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d: %v", len(secrets), secrets)
	}
	if _, ok := secrets["DB_URL"]; ok {
		t.Error("DB_URL should have been filtered out")
	}
}

func TestSuffixFilter_AllowsMatchingKeys(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{
		"DB_PASSWORD": "s3cr3t",
		"DB_HOST":     "localhost",
		"API_PASSWORD": "p4ss",
	})
	p, _ := provider.NewFilterProvider(inner, provider.SuffixFilter("_PASSWORD"))

	secrets, err := p.GetSecretsByPath(context.Background(), "cfg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(secrets))
	}
	if _, ok := secrets["DB_HOST"]; ok {
		t.Error("DB_HOST should have been filtered out")
	}
}
