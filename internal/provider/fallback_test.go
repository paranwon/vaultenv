package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

// stubProvider is a minimal in-memory provider used for testing.
type stubProvider struct {
	secrets map[string]string
	err     error
}

func (s *stubProvider) GetSecret(_ context.Context, path, key string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	if v, ok := s.secrets[path+"#"+key]; ok {
		return v, nil
	}
	return "", provider.ErrNotFound
}

func (s *stubProvider) GetSecretsByPath(_ context.Context, path string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := map[string]string{}
	for k, v := range s.secrets {
		if len(k) > len(path) && k[:len(path)] == path {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil, provider.ErrNotFound
	}
	return out, nil
}

func TestFallbackProvider_PrimarySucceeds(t *testing.T) {
	primary := &stubProvider{secrets: map[string]string{"secret/app#token": "primary-token"}}
	fallback := &stubProvider{secrets: map[string]string{"secret/app#token": "fallback-token"}}
	p := provider.NewFallbackProvider(primary, fallback)

	val, err := p.GetSecret(context.Background(), "secret/app", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "primary-token" {
		t.Errorf("expected primary-token, got %q", val)
	}
}

func TestFallbackProvider_FallsBackOnNotFound(t *testing.T) {
	primary := &stubProvider{secrets: map[string]string{}}
	fallback := &stubProvider{secrets: map[string]string{"secret/app#token": "fallback-token"}}
	p := provider.NewFallbackProvider(primary, fallback)

	val, err := p.GetSecret(context.Background(), "secret/app", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "fallback-token" {
		t.Errorf("expected fallback-token, got %q", val)
	}
}

func TestFallbackProvider_PropagatesPrimaryError(t *testing.T) {
	permanentErr := errors.New("permission denied")
	primary := &stubProvider{err: permanentErr}
	fallback := &stubProvider{secrets: map[string]string{"secret/app#token": "fallback-token"}}
	p := provider.NewFallbackProvider(primary, fallback)

	_, err := p.GetSecret(context.Background(), "secret/app", "token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, provider.ErrNotFound) {
		t.Error("should not return ErrNotFound for a non-404 primary error")
	}
}

func TestFallbackProvider_GetSecretsByPath_FallsBack(t *testing.T) {
	primary := &stubProvider{secrets: map[string]string{}}
	fallback := &stubProvider{secrets: map[string]string{"secret/app#key": "v"}}
	p := provider.NewFallbackProvider(primary, fallback)

	vals, err := p.GetSecretsByPath(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vals) == 0 {
		t.Error("expected at least one secret from fallback")
	}
}
