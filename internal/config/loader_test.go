package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnv_ExplicitPath(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "myconfig.yaml")

	content := []byte(`provider: vault
vault:
  address: http://vault:8200
  token: tok
secrets:
  - path: app/secret
    key: val
    env_var: SECRET_VAL
`)
	if err := os.WriteFile(cfgPath, content, 0o600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}

	t.Setenv("VAULTENV_CONFIG", cfgPath)

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != ProviderVault {
		t.Errorf("expected vault provider, got %q", cfg.Provider)
	}
	if len(cfg.Secrets) != 1 {
		t.Errorf("expected 1 secret, got %d", len(cfg.Secrets))
	}
}

func TestLoadDefault_NoneFound(t *testing.T) {
	// Change to a temp dir so no default config files exist.
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	_, err := LoadDefault()
	if err == nil {
		t.Fatal("expected error when no config file present, got nil")
	}
}

func TestLoadDefault_PicksFirstCandidate(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	content := []byte(`provider: ssm
ssm:
  region: eu-west-1
secrets:
  - path: /prod/db
    env_var: DB_URL
`)
	if err := os.WriteFile(".vaultenv.yaml", content, 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	cfg, err := LoadDefault()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != ProviderSSM {
		t.Errorf("expected ssm, got %q", cfg.Provider)
	}
}
