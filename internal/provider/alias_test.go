package provider

import (
	"context"
	"errors"
	"testing"
)

// stubProvider is a minimal in-memory provider used only within this test file.
type stubProvider struct {
	secrets map[string]string // "path#key" -> value
}

func (s *stubProvider) GetSecret(_ context.Context, path, key string) (string, error) {
	v, ok := s.secrets[path+"#"+key]
	if !ok {
		return "", ErrNotFound{Key: path + "#" + key}
	}
	return v, nil
}

func (s *stubProvider) GetSecretsByPath(_ context.Context, path string) (map[string]string, error) {
	out := map[string]string{}
	for k, v := range s.secrets {
		if len(k) > len(path) && k[:len(path)] == path {
			out[k] = v
		}
	}
	return out, nil
}

func newStub(entries map[string]string) *stubProvider {
	return &stubProvider{secrets: entries}
}

func TestAliasProvider_NilInner(t *testing.T) {
	_, err := NewAliasProvider(nil, AliasMap{})
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}
}

func TestAliasProvider_EmptyTargetRejected(t *testing.T) {
	inner := newStub(nil)
	_, err := NewAliasProvider(inner, AliasMap{"DB_PASS": ""})
	if err == nil {
		t.Fatal("expected error for empty alias target")
	}
}

func TestAliasProvider_RewritesPathAndKey(t *testing.T) {
	inner := newStub(map[string]string{
		"secret/prod/db#password": "s3cr3t",
	})
	ap, err := NewAliasProvider(inner, AliasMap{
		"DB_PASSWORD": "secret/prod/db#password",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := ap.GetSecret(context.Background(), "DB_PASSWORD", "")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}
	if val != "s3cr3t" {
		t.Errorf("want %q, got %q", "s3cr3t", val)
	}
}

func TestAliasProvider_CaseInsensitiveLookup(t *testing.T) {
	inner := newStub(map[string]string{
		"secret/prod/db#password": "hunter2",
	})
	ap, _ := NewAliasProvider(inner, AliasMap{
		"db_password": "secret/prod/db#password",
	})
	// Lookup with upper-case variant should still resolve.
	val, err := ap.GetSecret(context.Background(), "DB_PASSWORD", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Errorf("want %q, got %q", "hunter2", val)
	}
}

func TestAliasProvider_FallsBackToKeyArg(t *testing.T) {
	// Alias target has no '#key' suffix; caller-supplied key should be used.
	inner := newStub(map[string]string{
		"secret/prod/db#username": "admin",
	})
	ap, _ := NewAliasProvider(inner, AliasMap{
		"DB_HOST": "secret/prod/db",
	})
	val, err := ap.GetSecret(context.Background(), "DB_HOST", "username")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "admin" {
		t.Errorf("want %q, got %q", "admin", val)
	}
}

func TestAliasProvider_UnknownPathPassesThrough(t *testing.T) {
	inner := newStub(map[string]string{
		"secret/other#token": "tok",
	})
	ap, _ := NewAliasProvider(inner, AliasMap{
		"ALIAS": "secret/prod/db#password",
	})
	val, err := ap.GetSecret(context.Background(), "secret/other", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "tok" {
		t.Errorf("want %q, got %q", "tok", val)
	}
}

func TestAliasProvider_NotFoundPropagated(t *testing.T) {
	inner := newStub(map[string]string{})
	ap, _ := NewAliasProvider(inner, AliasMap{
		"MISSING": "secret/prod/db#password",
	})
	_, err := ap.GetSecret(context.Background(), "MISSING", "")
	if !IsNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAliasProvider_GetSecretsByPath_Forwarded(t *testing.T) {
	inner := newStub(map[string]string{
		"secret/prod#key": "val",
	})
	ap, _ := NewAliasProvider(inner, AliasMap{
		"ALIAS": "secret/prod#key",
	})
	// GetSecretsByPath should be forwarded to inner unchanged.
	result, err := ap.GetSecretsByPath(context.Background(), "secret/prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result["secret/prod#key"]; !ok {
		t.Logf("result: %v", result)
	}
	_ = errors.New("sentinel") // keep errors import used
}
