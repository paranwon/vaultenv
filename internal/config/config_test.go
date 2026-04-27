package config

import (
	"testing"
)

const validYAML = `
provider: vault
vault:
  address: http://127.0.0.1:8200
  token: root
  mount_path: secret
secrets:
  - path: myapp/db
    key: password
    env_var: DB_PASSWORD
  - path: myapp/api
    key: key
    env_var: API_KEY
`

func TestLoad_ValidConfig(t *testing.T) {
	cfg, err := parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != ProviderVault {
		t.Errorf("expected provider vault, got %q", cfg.Provider)
	}
	if len(cfg.Secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d", len(cfg.Secrets))
	}
	if cfg.Secrets[0].EnvVar != "DB_PASSWORD" {
		t.Errorf("expected DB_PASSWORD, got %q", cfg.Secrets[0].EnvVar)
	}
}

func TestLoad_InvalidProvider(t *testing.T) {
	yml := `provider: unknown
secrets:
  - path: foo
    env_var: FOO
`
	_, err := parse([]byte(yml))
	if err == nil {
		t.Fatal("expected error for invalid provider, got nil")
	}
}

func TestLoad_MissingPath(t *testing.T) {
	yml := `provider: ssm
secrets:
  - path: ""
    env_var: FOO
`
	_, err := parse([]byte(yml))
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestLoad_MissingEnvVar(t *testing.T) {
	yml := `provider: ssm
secrets:
  - path: /myapp/secret
    env_var: ""
`
	_, err := parse([]byte(yml))
	if err == nil {
		t.Fatal("expected error for empty env_var, got nil")
	}
}

func TestLoad_SSMProvider(t *testing.T) {
	yml := `provider: ssm
ssm:
  region: us-east-1
secrets:
  - path: /myapp/token
    env_var: APP_TOKEN
`
	cfg, err := parse([]byte(yml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Provider != ProviderSSM {
		t.Errorf("expected ssm provider, got %q", cfg.Provider)
	}
	if cfg.SSM.Region != "us-east-1" {
		t.Errorf("expected region us-east-1, got %q", cfg.SSM.Region)
	}
}
