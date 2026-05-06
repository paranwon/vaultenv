package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nicholasgasior/vaultenv/internal/provider"
)

func TestObservingProvider_NilInner(t *testing.T) {
	_, err := provider.NewObservingProvider(nil, "test", nil)
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestObservingProvider_GetSecret_EmitsEvent(t *testing.T) {
	static, _ := provider.NewStaticProvider(map[string]string{"secret/foo:bar": "baz"})

	var got provider.SecretEvent
	op, err := provider.NewObservingProvider(static, "static", func(e provider.SecretEvent) {
		got = e
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := op.GetSecret(context.Background(), "secret/foo", "bar")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if val != "baz" {
		t.Errorf("value: got %q, want %q", val, "baz")
	}
	if got.Provider != "static" {
		t.Errorf("Provider: got %q, want %q", got.Provider, "static")
	}
	if got.Path != "secret/foo" {
		t.Errorf("Path: got %q, want %q", got.Path, "secret/foo")
	}
	if got.Key != "bar" {
		t.Errorf("Key: got %q, want %q", got.Key, "bar")
	}
	if got.Err != nil {
		t.Errorf("Err: got %v, want nil", got.Err)
	}
	if got.Duration <= 0 {
		t.Error("Duration should be positive")
	}
}

func TestObservingProvider_GetSecret_ErrorPropagated(t *testing.T) {
	static, _ := provider.NewStaticProvider(map[string]string{})

	var gotErr error
	op, _ := provider.NewObservingProvider(static, "static", func(e provider.SecretEvent) {
		gotErr = e.Err
	})

	_, err := op.GetSecret(context.Background(), "secret/missing", "key")
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if !errors.Is(err, gotErr) {
		t.Errorf("observed error %v does not match returned error %v", gotErr, err)
	}
}

func TestObservingProvider_GetSecretsByPath_EmitsEvent(t *testing.T) {
	static, _ := provider.NewStaticProvider(map[string]string{"secret/svc:A": "1", "secret/svc:B": "2"})

	var got provider.SecretEvent
	op, _ := provider.NewObservingProvider(static, "s", func(e provider.SecretEvent) { got = e })

	vals, err := op.GetSecretsByPath(context.Background(), "secret/svc")
	if err != nil {
		t.Fatalf("GetSecretsByPath: %v", err)
	}
	if len(vals) != 2 {
		t.Errorf("expected 2 values, got %d", len(vals))
	}
	if got.Path != "secret/svc" {
		t.Errorf("Path: got %q", got.Path)
	}
}

func TestObservingProvider_NilObserve_IsNoOp(t *testing.T) {
	static, _ := provider.NewStaticProvider(map[string]string{"p:k": "v"})
	op, err := provider.NewObservingProvider(static, "s", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := op.GetSecret(context.Background(), "p", "k")
	if err != nil || val != "v" {
		t.Errorf("unexpected result: %q %v", val, err)
	}
}
