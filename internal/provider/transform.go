package provider

import (
	"context"
	"strings"
)

// TransformFunc is a function that transforms a secret value.
type TransformFunc func(value string) (string, error)

// transformProvider wraps a Provider and applies a TransformFunc to every
// secret value returned by GetSecret and GetSecretsByPath.
type transformProvider struct {
	inner     Provider
	transform TransformFunc
}

// NewTransformProvider returns a Provider that applies fn to every secret value
// returned by inner. If fn returns an error the call is propagated as-is.
func NewTransformProvider(inner Provider, fn TransformFunc) Provider {
	if inner == nil {
		panic("transform: inner provider must not be nil")
	}
	if fn == nil {
		panic("transform: transform func must not be nil")
	}
	return &transformProvider{inner: inner, transform: fn}
}

func (t *transformProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	val, err := t.inner.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}
	return t.transform(val)
}

func (t *transformProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := t.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		transformed, err := t.transform(v)
		if err != nil {
			return nil, err
		}
		out[k] = transformed
	}
	return out, nil
}

// TrimSpaceTransform is a built-in TransformFunc that trims leading and
// trailing whitespace from secret values.
func TrimSpaceTransform(value string) (string, error) {
	return strings.TrimSpace(value), nil
}

// UpperCaseTransform is a built-in TransformFunc that upper-cases secret values.
func UpperCaseTransform(value string) (string, error) {
	return strings.ToUpper(value), nil
}
