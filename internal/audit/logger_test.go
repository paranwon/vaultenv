package audit

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestLogger_Disabled(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, false)
	_ = l.SecretFetched("vault", "secret/app#key", "API_KEY")
	if buf.Len() != 0 {
		t.Fatalf("expected no output when disabled, got %q", buf.String())
	}
}

func TestLogger_SecretFetched(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, true)
	if err := l.SecretFetched("vault", "secret/app#key", "API_KEY"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var e Event
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.Action != "secret_fetched" {
		t.Errorf("expected action secret_fetched, got %q", e.Action)
	}
	if e.Provider != "vault" {
		t.Errorf("expected provider vault, got %q", e.Provider)
	}
	if e.Level != LevelInfo {
		t.Errorf("expected level INFO, got %q", e.Level)
	}
	if e.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestLogger_SecretError(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, true)
	if err := l.SecretError("ssm", "/app/secret", "DB_PASS", errors.New("not found")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var e Event
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.Level != LevelError {
		t.Errorf("expected level ERROR, got %q", e.Level)
	}
	if !strings.Contains(e.Error, "not found") {
		t.Errorf("expected error to contain 'not found', got %q", e.Error)
	}
}

func TestLogger_ProcessExec(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, true)
	if err := l.ProcessExec([]string{"/usr/bin/env", "bash"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var e Event
	if err := json.Unmarshal(buf.Bytes(), &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.Action != "process_exec" {
		t.Errorf("expected action process_exec, got %q", e.Action)
	}
	if e.Message != "/usr/bin/env" {
		t.Errorf("expected message /usr/bin/env, got %q", e.Message)
	}
}

func TestLogger_MultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, true)
	_ = l.SecretFetched("vault", "secret/a#key", "A")
	_ = l.SecretFetched("vault", "secret/b#key", "B")
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}
	for _, line := range lines {
		var e Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Errorf("invalid JSON line %q: %v", line, err)
		}
	}
}
