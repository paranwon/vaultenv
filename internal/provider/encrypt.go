package provider

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// DecryptFunc is a function that decrypts a ciphertext and returns plaintext.
type DecryptFunc func(ciphertext string) (string, error)

// encryptedProvider wraps a Provider and decrypts secret values on retrieval.
type encryptedProvider struct {
	inner   Provider
	decrypt DecryptFunc
}

// NewEncryptedProvider returns a Provider that decrypts values returned by inner
// using the supplied DecryptFunc. This is useful when secrets are stored
// encrypted at rest in Vault/SSM and need client-side decryption.
func NewEncryptedProvider(inner Provider, decrypt DecryptFunc) (Provider, error) {
	if inner == nil {
		return nil, fmt.Errorf("encrypted provider: inner provider must not be nil")
	}
	if decrypt == nil {
		return nil, fmt.Errorf("encrypted provider: decrypt func must not be nil")
	}
	return &encryptedProvider{inner: inner, decrypt: decrypt}, nil
}

func (p *encryptedProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	val, err := p.inner.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}
	plain, err := p.decrypt(val)
	if err != nil {
		return "", fmt.Errorf("encrypted provider: decrypt secret at %s/%s: %w", path, key, err)
	}
	return plain, nil
}

func (p *encryptedProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := p.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		plain, err := p.decrypt(v)
		if err != nil {
			return nil, fmt.Errorf("encrypted provider: decrypt key %s at path %s: %w", k, path, err)
		}
		out[k] = plain
	}
	return out, nil
}

// NewAESGCMDecryptFunc returns a DecryptFunc that decrypts AES-256-GCM
// base64-encoded ciphertexts produced by standard Go crypto/cipher.
// The ciphertext must be base64(nonce || ciphertext).
func NewAESGCMDecryptFunc(keyBytes []byte) (DecryptFunc, error) {
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm: create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aes-gcm: create GCM: %w", err)
	}
	return func(ciphertext string) (string, error) {
		data, err := base64.StdEncoding.DecodeString(ciphertext)
		if err != nil {
			return "", fmt.Errorf("aes-gcm: base64 decode: %w", err)
		}
		ns := gcm.NonceSize()
		if len(data) < ns {
			return "", fmt.Errorf("aes-gcm: ciphertext too short")
		}
		plain, err := gcm.Open(nil, data[:ns], data[ns:], nil)
		if err != nil {
			return "", fmt.Errorf("aes-gcm: decrypt: %w", err)
		}
		return string(plain), nil
	}, nil
}
