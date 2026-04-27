package env

import (
	"fmt"
	"strings"
)

// Mapping describes how a secret path/key maps to an environment variable name.
type Mapping struct {
	// EnvVar is the target environment variable name (e.g. "DB_PASSWORD").
	EnvVar string
	// Path is the secret path in the provider (e.g. "secret/myapp").
	Path string
	// Key is the field within the secret (e.g. "password"). Optional.
	Key string
}

// ParseMapping parses a mapping string of the form:
//
//	ENV_VAR=provider:path[#key]
//
// Examples:
//
//	DB_PASS=vault:secret/myapp#password
//	API_KEY=ssm:/myapp/api_key
func ParseMapping(s string) (Mapping, error) {
	eqIdx := strings.IndexByte(s, '=')
	if eqIdx <= 0 {
		return Mapping{}, fmt.Errorf("env: invalid mapping %q: missing '='" , s)
	}

	envVar := s[:eqIdx]
	rest := s[eqIdx+1:]

	if rest == "" {
		return Mapping{}, fmt.Errorf("env: invalid mapping %q: empty secret reference", s)
	}

	var path, key string
	if hashIdx := strings.LastIndexByte(rest, '#'); hashIdx >= 0 {
		path = rest[:hashIdx]
		key = rest[hashIdx+1:]
	} else {
		path = rest
	}

	if path == "" {
		return Mapping{}, fmt.Errorf("env: invalid mapping %q: empty path", s)
	}

	return Mapping{EnvVar: envVar, Path: path, Key: key}, nil
}

// ParseMappings parses multiple mapping strings and returns them all.
func ParseMappings(entries []string) ([]Mapping, error) {
	mappings := make([]Mapping, 0, len(entries))
	for _, e := range entries {
		m, err := ParseMapping(e)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}
	return mappings, nil
}
