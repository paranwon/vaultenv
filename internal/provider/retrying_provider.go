package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/your-org/vaultenv/internal/retry"
)

// HTTPStatusError is returned by providers when the upstream API responds
// with a non-2xx status code.
type HTTPStatusError struct {
	StatusCode int
	Message    string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// isRetryableProviderError decides whether a provider error warrants a retry.
func isRetryableProviderError(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *HTTPStatusError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode == http.StatusTooManyRequests {
			return true
		}
		if httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			return false
		}
	}
	return true
}

// RetryingProvider wraps any SecretProvider with automatic retry logic.
type RetryingProvider struct {
	inner  SecretProvider
	policy retry.Policy
}

// NewRetryingProvider wraps inner with the supplied retry policy.
func NewRetryingProvider(inner SecretProvider, policy retry.Policy) *RetryingProvider {
	return &RetryingProvider{inner: inner, policy: policy}
}

// GetSecret calls the underlying provider, retrying on transient errors.
func (r *RetryingProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	var result string
	err := retry.Do(ctx, r.policy, isRetryableProviderError, func() error {
		var e error
		result, e = r.inner.GetSecret(ctx, path, key)
		return e
	})
	return result, err
}
