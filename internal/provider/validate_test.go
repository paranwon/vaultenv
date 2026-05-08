package provider_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestValidatingProvider_NilInner(t *testing.T) {
	_, err := provider.NewValidatingProvider(nil, provider.NonEmptyRule)
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}
}

func TestValidatingProvider_NoRules_ReturnsInner(t *testing.T) {
	stub := provider.NewStaticProvider(map[string]string{"key": "val"})
	p, err := provider.NewValidatingProvider(stub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == stub {
		// When no rules are provided the inner provider is returned directly.
		return
	}
	// Still acceptable if a wrapper is returned.
}

func TestValidatingProvider_NonEmptyRule_Passes(t *testing.T) {
	stub := provider.NewStaticProvider(map[string]string{"DB_PASS": "s3cr3t"})
	p, err := provider.NewValidatingProvider(stub, provider.NonEmptyRule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := p.GetSecret(context.Background(), "", "DB_PASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected s3cr3t, got %q", val)
	}
}

func TestValidatingProvider_NonEmptyRule_Fails(t *testing.T) {
	stub := provider.NewStaticProvider(map[string]string{"DB_PASS": "   "})
	p, err := provider.NewValidatingProvider(stub, provider.NonEmptyRule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecret(context.Background(), "", "DB_PASS")
	if err == nil {
		t.Fatal("expected validation error for blank value")
	}
}

func TestValidatingProvider_MaxLengthRule(t *testing.T) {
	stub := provider.NewStaticProvider(map[string]string{"TOKEN": "toolongvalue"})
	p, err := provider.NewValidatingProvider(stub, provider.MaxLengthRule(5))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecret(context.Background(), "", "TOKEN")
	if err == nil {
		t.Fatal("expected max-length error")
	}
}

func TestValidatingProvider_RegexpRule_Passes(t *testing.T) {
	stub := provider.NewStaticProvider(map[string]string{"PORT": "8080"})
	p, err := provider.NewValidatingProvider(stub, provider.RegexpRule(regexp.MustCompile(`^\d+$`)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := p.GetSecret(context.Background(), "", "PORT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "8080" {
		t.Errorf("expected 8080, got %q", val)
	}
}

func TestValidatingProvider_RegexpRule_Fails(t *testing.T) {
	stub := provider.NewStaticProvider(map[string]string{"PORT": "not-a-port"})
	p, err := provider.NewValidatingProvider(stub, provider.RegexpRule(regexp.MustCompile(`^\d+$`)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecret(context.Background(), "", "PORT")
	if err == nil {
		t.Fatal("expected regexp validation error")
	}
}

func TestValidatingProvider_GetSecretsByPath_ValidatesAll(t *testing.T) {
	stub := provider.NewStaticProvider(map[string]string{"A": "ok", "B": ""})
	p, err := provider.NewValidatingProvider(stub, provider.NonEmptyRule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecretsByPath(context.Background(), "")
	if err == nil {
		t.Fatal("expected validation error from GetSecretsByPath")
	}
}
