package config

import (
	"fmt"
	"strings"
)

// TransformConfig holds the configuration for value transformation.
type TransformConfig struct {
	// Name is the identifier of the transform to apply (e.g. "trim", "upper").
	Name string `yaml:"name" json:"name"`
}

// KnownTransforms lists the built-in transform names accepted in config files.
var KnownTransforms = []string{"trim", "upper", "none"}

// ValidateTransform returns an error if the transform name is not recognised.
func ValidateTransform(name string) error {
	norm := strings.ToLower(strings.TrimSpace(name))
	for _, k := range KnownTransforms {
		if norm == k {
			return nil
		}
	}
	return fmt.Errorf("unknown transform %q: must be one of %s",
		name, strings.Join(KnownTransforms, ", "))
}

// DefaultTransformConfig returns the default (no-op) transform configuration.
func DefaultTransformConfig() TransformConfig {
	return TransformConfig{Name: "none"}
}

// IsNoOp reports whether the transform is a no-op (i.e. "none" or empty).
// This can be used to skip unnecessary processing when applying transforms.
func (t TransformConfig) IsNoOp() bool {
	norm := strings.ToLower(strings.TrimSpace(t.Name))
	return norm == "none" || norm == ""
}
