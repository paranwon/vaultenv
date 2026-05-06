package provider_test

import (
	"context"
	"crypto/rand"
	"io"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

// TestEncryptedProvider_ChainedWithPrefix verifies that an EncryptedProvider
// correctly decrypts values when composed with a PrefixProvider.
func TestEncryptedProvider_ChainedWithPrefix(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("rand key: %v", err)
	}

	encVal := aesGCMEncrypt(t, key, "plaintext-value")
	base := provider.NewStaticProvider(map[string]string{
		"prod/service/db:password": encVal,
	})

	prefixed, err := provider.NewPrefixProvider(base, "prod/")
	if err != nil {
		t.Fatalf("prefix provider: %v", err)
	}

	decFn, err := provider.NewAESGCMDecryptFunc(key)
	if err != nil {
		t.Fatalf("decrypt func: %v", err)
	}
	encrypted, err := provider.NewEncryptedProvider(prefixed, decFn)
	if err != nil {
		t.Fatalf("encrypted provider: %v", err)
	}

	got, err := encrypted.GetSecret(context.Background(), "service/db", "password")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if got != "plaintext-value" {
		t.Errorf("want %q got %q", "plaintext-value", got)
	}
}
