package env

import (
	"strings"
	"testing"
)

func TestMerge_AddsSecrets(t *testing.T) {
	inj := &Injector{baseEnv: []string{"PATH=/usr/bin", "HOME=/root"}}

	result := inj.Merge(map[string]string{
		"DB_PASSWORD": "s3cr3t",
		"API_KEY":     "abc123",
	})

	if !containsEntry(result, "DB_PASSWORD=s3cr3t") {
		t.Error("expected DB_PASSWORD to be injected")
	}
	if !containsEntry(result, "API_KEY=abc123") {
		t.Error("expected API_KEY to be injected")
	}
	if !containsEntry(result, "PATH=/usr/bin") {
		t.Error("expected PATH to be preserved")
	}
}

func TestMerge_OverridesExisting(t *testing.T) {
	inj := &Injector{baseEnv: []string{"DB_PASSWORD=old", "HOME=/root"}}

	result := inj.Merge(map[string]string{"DB_PASSWORD": "new"})

	var count int
	for _, e := range result {
		if strings.HasPrefix(e, "DB_PASSWORD=") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 DB_PASSWORD entry, got %d", count)
	}
	if !containsEntry(result, "DB_PASSWORD=new") {
		t.Error("expected DB_PASSWORD to be overridden to 'new'")
	}
}

func TestMerge_EmptySecrets(t *testing.T) {
	base := []string{"FOO=bar", "BAZ=qux"}
	inj := &Injector{baseEnv: base}

	result := inj.Merge(nil)

	if len(result) != len(base) {
		t.Errorf("expected %d entries, got %d", len(base), len(result))
	}
}

func TestExec_EmptyArgv(t *testing.T) {
	inj := NewInjector()
	err := inj.Exec(nil, []string{})
	if err == nil {
		t.Fatal("expected error for empty argv")
	}
}

func containsEntry(env []string, entry string) bool {
	for _, e := range env {
		if e == entry {
			return true
		}
	}
	return false
}
