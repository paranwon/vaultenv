package audit

import (
	"errors"
	"testing"
)

// TestNoopLogger verifies that NoopLogger never returns an error and
// never panics regardless of inputs.
func TestNoopLogger_NeverErrors(t *testing.T) {
	l := NoopLogger

	if err := l.SecretFetched("vault", "secret/app#key", "API_KEY"); err != nil {
		t.Errorf("SecretFetched: unexpected error: %v", err)
	}
	if err := l.SecretError("ssm", "/app/db", "DB_PASS", errors.New("boom")); err != nil {
		t.Errorf("SecretError: unexpected error: %v", err)
	}
	if err := l.ProcessExec([]string{"bash", "-c", "echo hi"}); err != nil {
		t.Errorf("ProcessExec: unexpected error: %v", err)
	}
	if err := l.Log(Event{Action: "custom"}); err != nil {
		t.Errorf("Log: unexpected error: %v", err)
	}
}

// TestNoopLogger_ImplementsAuditor ensures compile-time interface satisfaction
// is correct (the var _ check in noop.go covers this, but an explicit test
// makes the intent clear in test output).
func TestNoopLogger_ImplementsAuditor(t *testing.T) {
	var a Auditor = NoopLogger
	if a == nil {
		t.Fatal("expected non-nil Auditor")
	}
}
