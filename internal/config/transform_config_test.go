package config_test

import (
	"testing"

	"github.com/your-org/vaultenv/internal/config"
)

func TestValidateTransform_KnownNames(t *testing.T) {
	for _, name := range []string{"trim", "upper", "none", "TRIM", "Upper"} {
		if err := config.ValidateTransform(name); err != nil {
			t.Errorf("expected %q to be valid, got: %v", name, err)
		}
	}
}

func TestValidateTransform_UnknownName(t *testing.T) {
	if err := config.ValidateTransform("base64"); err == nil {
		t.Error("expected error for unknown transform 'base64'")
	}
}

func TestValidateTransform_EmptyString(t *testing.T) {
	if err := config.ValidateTransform(""); err == nil {
		t.Error("expected error for empty transform name")
	}
}

func TestDefaultTransformConfig_IsNone(t *testing.T) {
	cfg := config.DefaultTransformConfig()
	if cfg.Name != "none" {
		t.Errorf("expected default transform 'none', got %q", cfg.Name)
	}
}

func TestValidateTransform_WithWhitespace(t *testing.T) {
	if err := config.ValidateTransform("  trim  "); err != nil {
		t.Errorf("expected whitespace-padded 'trim' to be valid, got: %v", err)
	}
}
