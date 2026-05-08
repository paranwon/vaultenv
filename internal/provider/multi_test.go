package provider

import (
	"context"
	"errors"
	"testing"
)

// stubProvider is a simple in-memory Provider used for testing.
type stubProvider struct {
	secrets map[string]string
	err     error
}

func (s *stubProvider) GetSecret(_ context.Context, path, key string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	v, ok := s.secrets[path+"#"+key]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (s *stubProvider) GetSecretsByPath(_ context.Context, path string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := map[string]string{}
	for k, v := range s.secrets {
		out[k] = v
	}
	return out, nil
}

func TestMultiProvider_FirstSucceeds(t *testing.T) {
	p1 := &stubProvider{secrets: map[string]string{"sec/a#token": "abc123"}}
	p2 := &stubProvider{secrets: map[string]string{"sec/a#token": "other"}}
	mp, err := NewMultiProvider(p1, p2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := mp.GetSecret(context.Background(), "sec/a", "token")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if val != "abc123" {
		t.Errorf("expected abc123, got %q", val)
	}
}

func TestMultiProvider_FallsBackToSecond(t *testing.T) {
	p1 := &stubProvider{err: errors.New("vault unavailable")}
	p2 := &stubProvider{secrets: map[string]string{"sec/a#token": "fallback"}}
	mp, _ := NewMultiProvider(p1, p2)
	val, err := mp.GetSecret(context.Background(), "sec/a", "token")
	if err != nil {
		t.Fatalf("expected fallback success, got: %v", err)
	}
	if val != "fallback" {
		t.Errorf("expected fallback, got %q", val)
	}
}

func TestMultiProvider_AllFail(t *testing.T) {
	p1 := &stubProvider{err: errors.New("err1")}
	p2 := &stubProvider{err: errors.New("err2")}
	mp, _ := NewMultiProvider(p1, p2)
	_, err := mp.GetSecret(context.Background(), "sec/a", "token")
	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestMultiProvider_NoProviders(t *testing.T) {
	_, err := NewMultiProvider()
	if err == nil {
		t.Fatal("expected error for empty provider list")
	}
}

func TestMultiProvider_GetSecretsByPath_Fallback(t *testing.T) {
	p1 := &stubProvider{err: errors.New("unavailable")}
	p2 := &stubProvider{secrets: map[string]string{"k": "v"}}
	mp, _ := NewMultiProvider(p1, p2)
	vals, err := mp.GetSecretsByPath(context.Background(), "any/path")
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if vals["k"] != "v" {
		t.Errorf("unexpected values: %v", vals)
	}
}

func TestMultiProvider_GetSecretsByPath_AllFail(t *testing.T) {
	p1 := &stubProvider{err: errors.New("err1")}
	p2 := &stubProvider{err: errors.New("err2")}
	mp, _ := NewMultiProvider(p1, p2)
	_, err := mp.GetSecretsByPath(context.Background(), "any/path")
	if err == nil {
		t.Fatal("expected error when all providers fail for GetSecretsByPath")
	}
}
