package provider_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/nicholasgasior/vaultenv/internal/provider"
)

func TestSanitizeProvider_NilInner(t *testing.T) {
	_, err := provider.NewSanitizeProvider(nil, provider.TrimWhitespaceSanitize)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestSanitizeProvider_NilFunc(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"secret/key": "value"})
	_, err := provider.NewSanitizeProvider(inner, nil)
	if err == nil {
		t.Fatal("expected error for nil sanitize func")
	}
}

func TestSanitizeProvider_GetSecret_TrimWhitespace(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"secret/api": "  token123  "})
	p, err := provider.NewSanitizeProvider(inner, provider.TrimWhitespaceSanitize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := p.GetSecret(context.Background(), "secret", "api")
	if err != nil {
		t.Fatalf("GetSecret error: %v", err)
	}
	if got != "token123" {
		t.Errorf("got %q, want %q", got, "token123")
	}
}

func TestSanitizeProvider_GetSecret_StripControlChars(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"secret/key": "val\x01ue\x00"})
	p, err := provider.NewSanitizeProvider(inner, provider.StripControlChars)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := p.GetSecret(context.Background(), "secret", "key")
	if err != nil {
		t.Fatalf("GetSecret error: %v", err)
	}
	if got != "value" {
		t.Errorf("got %q, want %q", got, "value")
	}
}

func TestSanitizeProvider_GetSecretsByPath_AppliesSanitize(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{
		"cfg/db": "  postgres  ",
		"cfg/user": "  admin  ",
	})
	p, err := provider.NewSanitizeProvider(inner, provider.TrimWhitespaceSanitize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	secrets, err := p.GetSecretsByPath(context.Background(), "cfg")
	if err != nil {
		t.Fatalf("GetSecretsByPath error: %v", err)
	}
	for k, v := range secrets {
		if strings.TrimSpace(v) != v {
			t.Errorf("key %s: value %q still has whitespace", k, v)
		}
	}
}

func TestSanitizeProvider_SanitizeFuncError_Propagates(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"sec/key": "bad"})
	failFn := func(string) (string, error) { return "", errors.New("rejected") }
	p, err := provider.NewSanitizeProvider(inner, failFn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecret(context.Background(), "sec", "key")
	if err == nil {
		t.Fatal("expected error from sanitize func")
	}
	if !strings.Contains(err.Error(), "sanitize") {
		t.Errorf("error %q missing 'sanitize' prefix", err.Error())
	}
}

func TestSanitizeProvider_InnerNotFound_Propagates(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{})
	p, err := provider.NewSanitizeProvider(inner, provider.TrimWhitespaceSanitize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = p.GetSecret(context.Background(), "missing", "key")
	if !provider.IsNotFound(err) {
		t.Errorf("expected not-found, got %v", err)
	}
}
