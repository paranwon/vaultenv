package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newFakeVaultServer(t *testing.T, path string, payload map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/"+path {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(payload)
	}))
}

func TestVaultProvider_GetSecret_KVv2(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"data": map[string]interface{}{
				"DB_PASSWORD": "s3cr3t",
				"API_KEY":     "abc123",
			},
		},
	}

	srv := newFakeVaultServer(t, "secret/data/myapp", payload)
	defer srv.Close()

	p, err := NewVaultProvider(VaultConfig{Address: srv.URL, Token: "test-token"})
	if err != nil {
		t.Fatalf("NewVaultProvider: %v", err)
	}

	secrets, err := p.GetSecret(context.Background(), "secret/data/myapp")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}

	if secrets["DB_PASSWORD"] != "s3cr3t" {
		t.Errorf("expected DB_PASSWORD=s3cr3t, got %q", secrets["DB_PASSWORD"])
	}
	if secrets["API_KEY"] != "abc123" {
		t.Errorf("expected API_KEY=abc123, got %q", secrets["API_KEY"])
	}
}

func TestVaultProvider_GetSecret_KeyFilter(t *testing.T) {
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"data": map[string]interface{}{
				"DB_PASSWORD": "s3cr3t",
				"API_KEY":     "abc123",
			},
		},
	}

	srv := newFakeVaultServer(t, "secret/data/myapp", payload)
	defer srv.Close()

	p, _ := NewVaultProvider(VaultConfig{Address: srv.URL, Token: "test-token"})

	secrets, err := p.GetSecret(context.Background(), "secret/data/myapp#DB_PASSWORD")
	if err != nil {
		t.Fatalf("GetSecret with filter: %v", err)
	}
	if len(secrets) != 1 {
		t.Errorf("expected 1 key, got %d", len(secrets))
	}
	if secrets["DB_PASSWORD"] != "s3cr3t" {
		t.Errorf("expected DB_PASSWORD=s3cr3t, got %q", secrets["DB_PASSWORD"])
	}
}

func TestSplitPathKey(t *testing.T) {
	path, key := splitPathKey("secret/data/app#MY_KEY")
	if path != "secret/data/app" || key != "MY_KEY" {
		t.Errorf("unexpected split: path=%q key=%q", path, key)
	}

	path, key = splitPathKey("secret/data/app")
	if path != "secret/data/app" || key != "" {
		t.Errorf("unexpected split: path=%q key=%q", path, key)
	}
}
