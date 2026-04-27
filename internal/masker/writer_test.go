package masker

import (
	"bytes"
	"testing"
)

func TestWriter_MasksSecret(t *testing.T) {
	m := New()
	m.Add("s3cr3t")

	var buf bytes.Buffer
	w := NewWriter(&buf, m)

	n, err := w.Write([]byte("the password is s3cr3t ok"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len("the password is s3cr3t ok") {
		t.Fatalf("expected n=%d, got %d", len("the password is s3cr3t ok"), n)
	}

	got := buf.String()
	if got != "the password is ****** ok" {
		t.Fatalf("expected masked output, got %q", got)
	}
}

func TestWriter_NoSecrets_PassThrough(t *testing.T) {
	m := New()

	var buf bytes.Buffer
	w := NewWriter(&buf, m)

	input := "nothing secret here"
	n, err := w.Write([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(input) {
		t.Fatalf("expected n=%d, got %d", len(input), n)
	}
	if buf.String() != input {
		t.Fatalf("expected pass-through, got %q", buf.String())
	}
}

func TestWriter_MultipleSecrets(t *testing.T) {
	m := New()
	m.Add("alpha")
	m.Add("beta")

	var buf bytes.Buffer
	w := NewWriter(&buf, m)

	_, err := w.Write([]byte("alpha and beta are both masked"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if got != "***** and **** are both masked" {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestWriter_ReportsOriginalLength(t *testing.T) {
	m := New()
	m.Add("short")

	var buf bytes.Buffer
	w := NewWriter(&buf, m)

	// "short" -> "*****" same length, but test that n == original len regardless.
	input := []byte("short")
	n, err := w.Write(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(input) {
		t.Fatalf("expected n=%d, got %d", len(input), n)
	}
}
