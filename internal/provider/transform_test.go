package provider_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

type stubProvider struct {
	secret  string
	secrets map[string]string
	err     error
}

func (s *stubProvider) GetSecret(_ context.Context, _, _ string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.secret, nil
}

func (s *stubProvider) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.secrets, nil
}

func TestTransformProvider_GetSecret_AppliesTransform(t *testing.T) {
	inner := &stubProvider{secret: "  hello  "}
	p := provider.NewTransformProvider(inner, provider.TrimSpaceTransform)

	val, err := p.GetSecret(context.Background(), "path", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected 'hello', got %q", val)
	}
}

func TestTransformProvider_GetSecretsByPath_AppliesTransform(t *testing.T) {
	inner := &stubProvider{secrets: map[string]string{"A": "  foo  ", "B": " bar"}}
	p := provider.NewTransformProvider(inner, provider.TrimSpaceTransform)

	secrets, err := p.GetSecretsByPath(context.Background(), "path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secrets["A"] != "foo" || secrets["B"] != "bar" {
		t.Errorf("unexpected secrets: %v", secrets)
	}
}

func TestTransformProvider_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("backend down")
	inner := &stubProvider{err: sentinel}
	p := provider.NewTransformProvider(inner, provider.TrimSpaceTransform)

	_, err := p.GetSecret(context.Background(), "path", "key")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestTransformProvider_PropagatesTransformError(t *testing.T) {
	inner := &stubProvider{secret: "value"}
	boom := errors.New("transform failed")
	p := provider.NewTransformProvider(inner, func(string) (string, error) { return "", boom })

	_, err := p.GetSecret(context.Background(), "path", "key")
	if !errors.Is(err, boom) {
		t.Errorf("expected transform error, got %v", err)
	}
}

func TestTransformProvider_UpperCase(t *testing.T) {
	inner := &stubProvider{secret: "mysecret"}
	p := provider.NewTransformProvider(inner, provider.UpperCaseTransform)

	val, err := p.GetSecret(context.Background(), "path", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != strings.ToUpper("mysecret") {
		t.Errorf("expected upper-cased value, got %q", val)
	}
}

func TestNewTransformProvider_PanicsOnNilInner(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil inner")
		}
	}()
	provider.NewTransformProvider(nil, provider.TrimSpaceTransform)
}

func TestNewTransformProvider_PanicsOnNilFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil func")
		}
	}()
	provider.NewTransformProvider(&stubProvider{}, nil)
}
