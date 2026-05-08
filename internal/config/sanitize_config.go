package config

import "fmt"

// SanitizeMode names a built-in sanitize function.
type SanitizeMode string

const (
	SanitizeModeNone         SanitizeMode = "none"
	SanitizeModeTrimSpace    SanitizeMode = "trim_space"
	SanitizeModeStripControl SanitizeMode = "strip_control"
)

// SanitizeConfig controls the sanitize provider middleware.
type SanitizeConfig struct {
	Enabled bool         `yaml:"enabled"`
	Mode    SanitizeMode `yaml:"mode"`
}

// DefaultSanitizeConfig returns a disabled SanitizeConfig.
func DefaultSanitizeConfig() SanitizeConfig {
	return SanitizeConfig{
		Enabled: false,
		Mode:    SanitizeModeNone,
	}
}

// Validate returns an error if the configuration is inconsistent.
func (c SanitizeConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	switch c.Mode {
	case SanitizeModeTrimSpace, SanitizeModeStripControl:
		return nil
	case SanitizeModeNone, "":
		return fmt.Errorf("sanitize: mode must be set when enabled")
	default:
		return fmt.Errorf("sanitize: unknown mode %q", c.Mode)
	}
}
