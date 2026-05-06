package provider_test

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

func aesGCMEncrypt(t *testing.T, key []byte, plaintext string) string {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("create cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("create gcm: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Fatalf("read nonce: %v", err)
	}
	ct := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ct)
}

func TestEncryptedProvider_GetSecret_Decrypts(t *testing.T) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("rand key: %v", err)
	}
	encrypted := aesGCMEncrypt(t, key, "super-secret")

	inner := provider.NewStaticProvider(map[string]string{"secret/db:password": encrypted})
	decFn, err := provider.NewAESGCMDecryptFunc(key)
	if err != nil {
		t.Fatalf("decrypt func: %v", err)
	}
	p, err := provider.NewEncryptedProvider(inner, decFn)
	if err != nil {
		t.Fatalf("new encrypted provider: %v", err)
	}

	got, err := p.GetSecret(context.Background(), "secret/db", "password")
	if err != nil {
		t.Fatalf("GetSecret: %v", err)
	}
	if got != "super-secret" {
		t.Errorf("want %q got %q", "super-secret", got)
	}
}

func TestEncryptedProvider_GetSecretsByPath_Decrypts(t *testing.T) {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key) //nolint:errcheck

	secrets := map[string]string{
		"secret/app:TOKEN": aesGCMEncrypt(t, key, "tok-abc"),
		"secret/app:API_KEY": aesGCMEncrypt(t, key, "key-xyz"),
	}
	inner := provider.NewStaticProvider(secrets)
	decFn, _ := provider.NewAESGCMDecryptFunc(key)
	p, _ := provider.NewEncryptedProvider(inner, decFn)

	got, err := p.GetSecretsByPath(context.Background(), "secret/app")
	if err != nil {
		t.Fatalf("GetSecretsByPath: %v", err)
	}
	if got["TOKEN"] != "tok-abc" {
		t.Errorf("TOKEN: want tok-abc got %q", got["TOKEN"])
	}
	if got["API_KEY"] != "key-xyz" {
		t.Errorf("API_KEY: want key-xyz got %q", got["API_KEY"])
	}
}

func TestEncryptedProvider_NilInner_ReturnsError(t *testing.T) {
	_, err := provider.NewEncryptedProvider(nil, func(s string) (string, error) { return s, nil })
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestEncryptedProvider_BadCiphertext_ReturnsError(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{"sec/db:pw": "not-base64!!!"})
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key) //nolint:errcheck
	decFn, _ := provider.NewAESGCMDecryptFunc(key)
	p, _ := provider.NewEncryptedProvider(inner, decFn)

	_, err := p.GetSecret(context.Background(), "sec/db", "pw")
	if err == nil {
		t.Fatal("expected decrypt error")
	}
}

func TestEncryptedProvider_PropagatesInnerError(t *testing.T) {
	inner := provider.NewStaticProvider(map[string]string{})
	decFn := func(s string) (string, error) { return s, nil }
	p, _ := provider.NewEncryptedProvider(inner, decFn)

	_, err := p.GetSecret(context.Background(), "missing", "key")
	if !errors.Is(err, provider.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}
