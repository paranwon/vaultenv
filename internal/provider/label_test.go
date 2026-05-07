package provider_test

import (
	"context"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestLabelProvider_NilInner(t *testing.T) {
	_, err := provider.NewLabelProvider(nil, map[string]string{"env": "prod"})
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestLabelProvider_EmptyLabels(t *testing.T) {
	stub := newStub(map[string]string{"k": "v"}, nil)
	_, err := provider.NewLabelProvider(stub, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty labels, got nil")
	}
}

func TestLabelProvider_GetSecret_Delegates(t *testing.T) {
	stub := newStub(map[string]string{"token": "abc123"}, nil)
	p, err := provider.NewLabelProvider(stub, map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := p.GetSecret(context.Background(), "secret/app", "token")
	if err != nil {
		t.Fatalf("GetSecret returned error: %v", err)
	}
	if val != "abc123" {
		t.Errorf("expected abc123, got %q", val)
	}
}

func TestLabelProvider_GetSecret_PropagatesError(t *testing.T) {
	stub := newStub(nil, provider.ErrNotFound{Key: "missing"})
	p, _ := provider.NewLabelProvider(stub, map[string]string{"env": "prod"})

	_, err := p.GetSecret(context.Background(), "secret/app", "missing")
	if !provider.IsNotFound(err) {
		t.Errorf("expected not-found error, got %v", err)
	}
}

func TestLabelProvider_GetSecretsByPath_InjectsLabels(t *testing.T) {
	stub := newStub(map[string]string{"db_pass": "hunter2"}, nil)
	p, err := provider.NewLabelProvider(stub, map[string]string{"env": "staging", "team": "platform"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secrets, err := p.GetSecretsByPath(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("GetSecretsByPath returned error: %v", err)
	}

	if secrets["db_pass"] != "hunter2" {
		t.Errorf("original secret missing or wrong: %v", secrets["db_pass"])
	}
	if secrets["__label__.env"] != "staging" {
		t.Errorf("expected label env=staging, got %q", secrets["__label__.env"])
	}
	if secrets["__label__.team"] != "platform" {
		t.Errorf("expected label team=platform, got %q", secrets["__label__.team"])
	}
}

func TestLabelProvider_Labels_ReturnsCopy(t *testing.T) {
	stub := newStub(nil, nil)
	p, _ := provider.NewLabelProvider(stub, map[string]string{"env": "prod"})

	labels := p.Labels()
	labels["env"] = "mutated"

	if p.Labels()["env"] != "prod" {
		t.Error("Labels() should return a defensive copy")
	}
}

func TestLabelProvider_GetSecretsByPath_PropagatesError(t *testing.T) {
	stub := newStub(nil, provider.ErrNotFound{Key: "path"})
	p, _ := provider.NewLabelProvider(stub, map[string]string{"env": "prod"})

	_, err := p.GetSecretsByPath(context.Background(), "secret/missing")
	if err == nil {
		t.Fatal("expected error from inner provider, got nil")
	}
}
