package provider

import (
	"context"
	"errors"
	"fmt"
)

// Ensure interface compliance.
var _ Provider = (*validatingProvider)(nil)

type validatingProvider struct {
	inner Provider
	rules []ValidationRule
}

// NewValidatingProvider returns a Provider that applies the given rules to every
// secret returned by inner. If any rule returns an error the fetch fails.
func NewValidatingProvider(inner Provider, rules ...ValidationRule) (Provider, error) {
	if inner == nil {
		return nil, errors.New("validating provider: inner provider must not be nil")
	}
	if len(rules) == 0 {
		return inner, nil
	}
	return &validatingProvider{inner: inner, rules: rules}, nil
}

func (v *validatingProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	val, err := v.inner.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}
	if err := v.validate(key, val); err != nil {
		return "", err
	}
	return val, nil
}

func (v *validatingProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := v.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	for k, val := range secrets {
		if err := v.validate(k, val); err != nil {
			return nil, fmt.Errorf("validating provider: path %q: %w", path, err)
		}
	}
	return secrets, nil
}

func (v *validatingProvider) validate(key, value string) error {
	for _, rule := range v.rules {
		if err := rule(key, value); err != nil {
			return fmt.Errorf("validating provider: %w", err)
		}
	}
	return nil
}
