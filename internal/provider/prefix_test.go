package provider

import (
	"context"
	"testing"
)

// stubProvider records the last path it received so tests can assert on it.
type stubProvider struct {
	lastPath      string
	secretValue   string
	secretsByPath map[string]string
	err           error
}

func (s *stubProvider) GetSecret(_ context.Context, path string) (string, error) {
	s.lastPath = path
	return s.secretValue, s.err
}

func (s *stubProvider) GetSecretsByPath(_ context.Context, path string) (map[string]string, error) {
	s.lastPath = path
	return s.secretsByPath, s.err
}

func TestPrefixProvider_GetSecret_PrependsPrefixOnce(t *testing.T) {
	stub := &stubProvider{secretValue: "hunter2"}
	p := NewPrefixProvider(stub, "prod/myapp")

	val, err := p.GetSecret(context.Background(), "db/password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hunter2" {
		t.Errorf("expected hunter2, got %q", val)
	}
	if stub.lastPath != "prod/myapp/db/password" {
		t.Errorf("expected prod/myapp/db/password, got %q", stub.lastPath)
	}
}

func TestPrefixProvider_GetSecret_LeadingSlashInPath(t *testing.T) {
	stub := &stubProvider{secretValue: "s3cr3t"}
	p := NewPrefixProvider(stub, "prod")

	_, _ = p.GetSecret(context.Background(), "/token")
	if stub.lastPath != "prod/token" {
		t.Errorf("expected prod/token, got %q", stub.lastPath)
	}
}

func TestPrefixProvider_GetSecret_TrailingSlashInPrefix(t *testing.T) {
	stub := &stubProvider{secretValue: "abc"}
	p := NewPrefixProvider(stub, "prod/myapp/")

	_, _ = p.GetSecret(context.Background(), "key")
	if stub.lastPath != "prod/myapp/key" {
		t.Errorf("expected prod/myapp/key, got %q", stub.lastPath)
	}
}

func TestPrefixProvider_GetSecretsByPath_PrependsPrefixOnce(t *testing.T) {
	expected := map[string]string{"host": "localhost", "port": "5432"}
	stub := &stubProvider{secretsByPath: expected}
	p := NewPrefixProvider(stub, "staging")

	got, err := p.GetSecretsByPath(context.Background(), "db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.lastPath != "staging/db" {
		t.Errorf("expected staging/db, got %q", stub.lastPath)
	}
	if len(got) != len(expected) {
		t.Errorf("expected %d secrets, got %d", len(expected), len(got))
	}
}
