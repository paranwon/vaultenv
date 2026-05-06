package provider_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestStaticProvider_GetSecret_Found(t *testing.T) {
	p := provider.NewStaticProvider(map[string]string{
		"myapp/db/password": "s3cr3t",
	})

	val, err := p.GetSecret(context.Background(), "myapp/db", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("expected %q, got %q", "s3cr3t", val)
	}
}

func TestStaticProvider_GetSecret_NotFound(t *testing.T) {
	p := provider.NewStaticProvider(map[string]string{})

	_, err := p.GetSecret(context.Background(), "myapp/db", "password")
	if err == nil {
		t.Fatal("expected ErrNotFound, got nil")
	}
	if !provider.IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true, got false for error: %v", err)
	}
}

func TestStaticProvider_GetSecret_CaseInsensitive(t *testing.T) {
	p := provider.NewStaticProvider(map[string]string{
		"MyApp/DB/Password": "hunter2",
	})

	val, err := p.GetSecret(context.Background(), "myapp/db", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Errorf("expected %q, got %q", "hunter2", val)
	}
}

func TestStaticProvider_GetSecretsByPath_Found(t *testing.T) {
	p := provider.NewStaticProvider(map[string]string{
		"myapp/db/host":     "localhost",
		"myapp/db/port":     "5432",
		"myapp/cache/host":  "redis",
	})

	secrets, err := p.GetSecretsByPath(context.Background(), "myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(secrets))
	}
}

func TestStaticProvider_GetSecretsByPath_NotFound(t *testing.T) {
	p := provider.NewStaticProvider(map[string]string{
		"myapp/db/host": "localhost",
	})

	_, err := p.GetSecretsByPath(context.Background(), "myapp/cache")
	if err == nil {
		t.Fatal("expected ErrNotFound, got nil")
	}
	if !provider.IsNotFound(err) {
		t.Errorf("expected IsNotFound, got: %v", err)
	}
}

func TestStaticProvider_GetSecretsByPath_TrailingSlash(t *testing.T) {
	p := provider.NewStaticProvider(map[string]string{
		"svc/tls/cert": "CERT",
		"svc/tls/key":  "KEY",
	})

	// Path with and without trailing slash should behave identically.
	secrets, err := p.GetSecretsByPath(context.Background(), "svc/tls/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(secrets) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(secrets))
	}
}
