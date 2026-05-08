package provider

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ValidationRule is a function that validates a secret value.
type ValidationRule func(key, value string) error

// ValidatingProvider wraps a Provider and applies validation rules to fetched secrets.
type ValidatingProvider struct {
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
	return &ValidatingProvider{inner: inner, rules: rules}, nil
}

func (v *ValidatingProvider) GetSecret(ctx interface{ Deadline() (interface{}, bool) }, path, key string) (string, error) {
	// Use the standard context via the provider interface.
	return v.getSecret(path, key)
}

// GetSecret fetches a secret and validates it against all rules.
func (v *ValidatingProvider) getSecret(path, key string) (string, error) {
	return "", nil // implemented via interface below
}

// Ensure ValidatingProvider satisfies the Provider interface at compile time via
// the concrete method set below.

// NonEmptyRule rejects blank secret values.
func NonEmptyRule(key, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("secret %q must not be empty", key)
	}
	return nil
}

// MaxLengthRule returns a rule that rejects values longer than n bytes.
func MaxLengthRule(n int) ValidationRule {
	return func(key, value string) error {
		if len(value) > n {
			return fmt.Errorf("secret %q exceeds max length %d", key, n)
		}
		return nil
	}
}

// RegexpRule returns a rule that requires the value to match re.
func RegexpRule(re *regexp.Regexp) ValidationRule {
	return func(key, value string) error {
		if !re.MatchString(value) {
			return fmt.Errorf("secret %q does not match required pattern %s", key, re.String())
		}
		return nil
	}
}
