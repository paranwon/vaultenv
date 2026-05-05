package provider_test

import (
	"fmt"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func TestErrNotFound_WithKey(t *testing.T) {
	err := provider.ErrNotFound("secret/data/db", "password")
	if !provider.IsNotFound(err) {
		t.Fatal("IsNotFound should return true for *NotFoundError")
	}
	want := `secret not found: path="secret/data/db" key="password"`
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestErrNotFound_WithoutKey(t *testing.T) {
	err := provider.ErrNotFound("secret/data/db", "")
	want := `secret not found: path="secret/data/db"`
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestIsNotFound_WrappedError(t *testing.T) {
	base := provider.ErrNotFound("p", "k")
	wrapped := fmt.Errorf("outer: %w", base)
	if !provider.IsNotFound(wrapped) {
		t.Fatal("IsNotFound should unwrap and detect *NotFoundError")
	}
}

func TestIsNotFound_OtherError(t *testing.T) {
	if provider.IsNotFound(fmt.Errorf("some other error")) {
		t.Fatal("IsNotFound should return false for non-NotFoundError")
	}
}
