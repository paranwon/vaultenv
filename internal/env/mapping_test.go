package env

import (
	"testing"
)

func TestParseMapping_Valid(t *testing.T) {
	m, err := ParseMapping("MY_SECRET=secret/data/app#password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.EnvVar != "MY_SECRET" {
		t.Errorf("expected EnvVar=MY_SECRET, got %q", m.EnvVar)
	}
	if m.Path != "secret/data/app" {
		t.Errorf("expected Path=secret/data/app, got %q", m.Path)
	}
	if m.Key != "password" {
		t.Errorf("expected Key=password, got %q", m.Key)
	}
}

func TestParseMapping_NoKey(t *testing.T) {
	m, err := ParseMapping("MY_PARAM=/ssm/param")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.EnvVar != "MY_PARAM" {
		t.Errorf("expected EnvVar=MY_PARAM, got %q", m.EnvVar)
	}
	if m.Path != "/ssm/param" {
		t.Errorf("expected Path=/ssm/param, got %q", m.Path)
	}
	if m.Key != "" {
		t.Errorf("expected empty Key, got %q", m.Key)
	}
}

func TestParseMapping_MissingEquals(t *testing.T) {
	_, err := ParseMapping("INVALID")
	if err == nil {
		t.Fatal("expected error for missing '=', got nil")
	}
}

func TestParseMapping_EmptyEnvVar(t *testing.T) {
	_, err := ParseMapping("=secret/data/app#key")
	if err == nil {
		t.Fatal("expected error for empty env var, got nil")
	}
}

func TestParseMapping_EmptyPath(t *testing.T) {
	_, err := ParseMapping("MY_VAR=")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestParseMappings_Multiple(t *testing.T) {
	inputs := []string{
		"DB_PASS=secret/data/db#password",
		"API_KEY=secret/data/app#api_key",
		"PARAM=/ssm/myapp/param",
	}
	mappings, err := ParseMappings(inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mappings) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(mappings))
	}
	if mappings[0].EnvVar != "DB_PASS" {
		t.Errorf("expected DB_PASS, got %q", mappings[0].EnvVar)
	}
	if mappings[2].Key != "" {
		t.Errorf("expected empty key for SSM param, got %q", mappings[2].Key)
	}
}

func TestParseMappings_InvalidEntry(t *testing.T) {
	inputs := []string{
		"GOOD=secret/data/app#key",
		"BAD_NO_EQUALS",
	}
	_, err := ParseMappings(inputs)
	if err == nil {
		t.Fatal("expected error for invalid entry, got nil")
	}
}
