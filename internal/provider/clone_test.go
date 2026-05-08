package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestCloneProvider_NilInner(t *testing.T) {
	_, err := provider.NewCloneProvider(nil)
	if err == nil {
		t.Fatal("expected error for nil inner provider")
	}
}

func TestCloneProvider_GetSecret_ReturnsValue(t *testing.T) {
	inner, _ := provider.NewStaticProvider(map[string]string{
		"secret/app#password": "hunter2",
	})
	cp, err := provider.NewCloneProvider(inner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := cp.GetSecret(context.Background(), "secret/app", "password")
	if err != nil {
		t.Fatalf("GetSecret error: %v", err)
	}
	if got != "hunter2" {
		t.Errorf("got %q, want %q", got, "hunter2")
	}
}

func TestCloneProvider_GetSecret_PropagatesError(t *testing.T) {
	inner, _ := provider.NewStaticProvider(map[string]string{})
	cp, _ := provider.NewCloneProvider(inner)

	_, err := cp.GetSecret(context.Background(), "secret/app", "missing")
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if !provider.IsNotFound(err) {
		t.Errorf("expected IsNotFound, got: %v", err)
	}
}

func TestCloneProvider_GetSecretsByPath_ReturnsCopy(t *testing.T) {
	inner, _ := provider.NewStaticProvider(map[string]string{
		"cfg/db#host": "localhost",
		"cfg/db#port": "5432",
	})
	cp, _ := provider.NewCloneProvider(inner)

	got, err := cp.GetSecretsByPath(context.Background(), "cfg/db")
	if err != nil {
		t.Fatalf("GetSecretsByPath error: %v", err)
	}
	if got["host"] != "localhost" {
		t.Errorf("host: got %q, want %q", got["host"], "localhost")
	}
	if got["port"] != "5432" {
		t.Errorf("port: got %q, want %q", got["port"], "5432")
	}
}

func TestCloneProvider_GetSecretsByPath_PropagatesError(t *testing.T) {
	sentinel := errors.New("backend unavailable")
	errProvider := &errStub{err: sentinel}
	cp, _ := provider.NewCloneProvider(errProvider)

	_, err := cp.GetSecretsByPath(context.Background(), "any/path")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}

// errStub is a minimal Provider that always returns a fixed error.
type errStub struct{ err error }

func (e *errStub) GetSecret(_ context.Context, _, _ string) (string, error) {
	return "", e.err
}
func (e *errStub) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	return nil, e.err
}
