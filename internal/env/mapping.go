package env

import (
	"fmt"
	"strings"
)

// Mapping represents a single environment variable → secret path binding.
// Format: ENV_VAR=path/to/secret[#key]
// The optional #key fragment selects a specific field within the secret.
type Mapping struct {
	EnvVar string // destination environment variable name
	Path   string // secret path in the provider (Vault or SSM)
	Key    string // optional field key within the secret
}

// ParseMapping parses a single mapping string of the form:
//
//	ENV_VAR=path/to/secret
//	ENV_VAR=path/to/secret#field
func ParseMapping(s string) (Mapping, error) {
	eqIdx := strings.IndexByte(s, '=')
	if eqIdx < 0 {
		return Mapping{}, fmt.Errorf("mapping %q: missing '=' separator", s)
	}

	envVar := s[:eqIdx]
	if envVar == "" {
		return Mapping{}, fmt.Errorf("mapping %q: environment variable name must not be empty", s)
	}

	rest := s[eqIdx+1:]
	if rest == "" {
		return Mapping{}, fmt.Errorf("mapping %q: secret path must not be empty", s)
	}

	path, key, _ := strings.Cut(rest, "#")

	return Mapping{
		EnvVar: envVar,
		Path:   path,
		Key:    key,
	}, nil
}

// ParseMappings parses a slice of mapping strings and returns all Mappings.
// It returns an error on the first invalid entry.
func ParseMappings(specs []string) ([]Mapping, error) {
	mappings := make([]Mapping, 0, len(specs))
	for _, spec := range specs {
		m, err := ParseMapping(spec)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}
	return mappings, nil
}
