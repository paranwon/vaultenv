package masker_test

import (
	"testing"

	"github.com/your-org/vaultenv/internal/masker"
)

const redacted = "***"

func TestMask_ReplacesKnownSecret(t *testing.T) {
	m := masker.New(redacted, "s3cr3t")
	got := m.Mask("the password is s3cr3t, keep it safe")
	want := "the password is ***, keep it safe"
	if got != want {
		t.Errorf("Mask() = %q, want %q", got, want)
	}
}

func TestMask_MultipleSecrets(t *testing.T) {
	m := masker.New(redacted, "alpha", "beta")
	got := m.Mask("alpha and beta should both be hidden")
	want := "*** and *** should both be hidden"
	if got != want {
		t.Errorf("Mask() = %q, want %q", got, want)
	}
}

func TestMask_NoMatch_ReturnsOriginal(t *testing.T) {
	m := masker.New(redacted, "hidden")
	input := "nothing sensitive here"
	got := m.Mask(input)
	if got != input {
		t.Errorf("Mask() = %q, want original %q", got, input)
	}
}

func TestMask_EmptySecretIgnored(t *testing.T) {
	m := masker.New(redacted, "", "real")
	got := m.Mask("real value")
	want := "*** value"
	if got != want {
		t.Errorf("Mask() = %q, want %q", got, want)
	}
}

func TestAdd_RegistersNewSecrets(t *testing.T) {
	m := masker.New(redacted, "first")
	m.Add("second")
	got := m.Mask("first and second")
	want := "*** and ***"
	if got != want {
		t.Errorf("after Add, Mask() = %q, want %q", got, want)
	}
}

func TestMaskEnv_RedactsValue(t *testing.T) {
	m := masker.New(redacted, "topsecret")
	got := m.MaskEnv("MY_TOKEN=topsecret")
	want := "MY_TOKEN=***"
	if got != want {
		t.Errorf("MaskEnv() = %q, want %q", got, want)
	}
}

func TestNew_EmptySecretsList(t *testing.T) {
	m := masker.New(redacted)
	got := m.Mask("anything")
	if got != "anything" {
		t.Errorf("empty masker should not alter input, got %q", got)
	}
}
