// Package masker provides utilities for redacting secret values
// from log output and error messages to prevent accidental disclosure.
package masker

import "strings"

// Masker replaces known secret values with a redaction placeholder.
type Masker struct {
	secrets []string
	placeholder string
}

// New returns a Masker that will replace any of the given secret values
// with placeholder in all processed strings.
func New(placeholder string, secrets ...string) *Masker {
	// Filter out empty strings to avoid replacing everything.
	filtered := make([]string, 0, len(secrets))
	for _, s := range secrets {
		if s != "" {
			filtered = append(filtered, s)
		}
	}
	return &Masker{
		secrets:     filtered,
		placeholder: placeholder,
	}
}

// Mask replaces all known secret values in s with the placeholder.
func (m *Masker) Mask(s string) string {
	for _, secret := range m.secrets {
		s = strings.ReplaceAll(s, secret, m.placeholder)
	}
	return s
}

// Add registers additional secret values to be masked.
func (m *Masker) Add(secrets ...string) {
	for _, s := range secrets {
		if s != "" {
			m.secrets = append(m.secrets, s)
		}
	}
}

// MaskEnv masks secret values found inside KEY=VALUE environment variable
// strings, returning the entry with the value replaced by the placeholder.
func (m *Masker) MaskEnv(entry string) string {
	return m.Mask(entry)
}
