package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

type stubChainProvider struct {
	secrets map[string]string
	hardErr error
}

func (s *stubChainProvider) GetSecret(_ context.Context, path, key string) (string, error) {
	if s.hardErr != nil {
		return "", s.hardErr
	}
	v, ok := s.secrets[path+"#"+key]
	if !ok {
		return "", provider.ErrNotFound(path, key)
	}
	return v, nil
}

func (s *stubChainProvider) GetSecretsByPath(_ context.Context, path string) (map[string]string, error) {
	if s.hardErr != nil {
		return nil, s.hardErr
	}
	out := map[string]string{}
	for k, v := range s.secrets {
		out[k] = v
	}
	if len(out) == 0 {
		return nil, provider.ErrNotFound(path, "")
	}
	return out, nil
}

func TestChainProvider_RequiresAtLeastOne(t *testing.T) {
	_, err := provider.NewChainProvider()
	if err == nil {
		t.Fatal("expected error for empty chain")
	}
}

func TestChainProvider_FirstSucceeds(t *testing.T) {
	p1 := &stubChainProvider{secrets: map[string]string{"sec/db#pass": "hunter2"}}
	p2 := &stubChainProvider{secrets: map[string]string{"sec/db#pass": "wrong"}}
	cp, _ := provider.NewChainProvider(p1, p2)

	val, err := cp.GetSecret(context.Background(), "sec/db", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Fatalf("got %q, want %q", val, "hunter2")
	}
}

func TestChainProvider_FallsBackOnNotFound(t *testing.T) {
	p1 := &stubChainProvider{secrets: map[string]string{}}
	p2 := &stubChainProvider{secrets: map[string]string{"sec/db#pass": "fallback"}}
	cp, _ := provider.NewChainProvider(p1, p2)

	val, err := cp.GetSecret(context.Background(), "sec/db", "pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "fallback" {
		t.Fatalf("got %q, want %q", val, "fallback")
	}
}

func TestChainProvider_HardErrorStopsChain(t *testing.T) {
	hardErr := errors.New("connection refused")
	p1 := &stubChainProvider{hardErr: hardErr}
	p2 := &stubChainProvider{secrets: map[string]string{"sec/db#pass": "never"}}
	cp, _ := provider.NewChainProvider(p1, p2)

	_, err := cp.GetSecret(context.Background(), "sec/db", "pass")
	if err == nil {
		t.Fatal("expected hard error to propagate")
	}
	if !errors.Is(err, hardErr) {
		t.Fatalf("expected wrapped hard error, got: %v", err)
	}
}

func TestChainProvider_AllNotFound(t *testing.T) {
	p1 := &stubChainProvider{secrets: map[string]string{}}
	p2 := &stubChainProvider{secrets: map[string]string{}}
	cp, _ := provider.NewChainProvider(p1, p2)

	_, err := cp.GetSecret(context.Background(), "sec/db", "pass")
	if !provider.IsNotFound(err) {
		t.Fatalf("expected not-found error, got: %v", err)
	}
}

func TestChainProvider_GetSecretsByPath_FallsBack(t *testing.T) {
	p1 := &stubChainProvider{secrets: map[string]string{}}
	p2 := &stubChainProvider{secrets: map[string]string{"k": "v"}}
	cp, _ := provider.NewChainProvider(p1, p2)

	secrets, err := cp.GetSecretsByPath(context.Background(), "any/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets["k"] != "v" {
		t.Fatalf("expected fallback secrets, got %v", secrets)
	}
}
