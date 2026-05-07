package provider

import (
	"context"
	"fmt"
	"strings"
)

// AliasMap maps logical alias names to concrete secret paths.
// For example: {"DB_PASSWORD": "secret/prod/database#password"}
type AliasMap map[string]string

// aliasProvider rewrites GetSecret calls by substituting a logical alias
// with the concrete path+key defined in the alias map, then delegates to
// the inner provider.
type aliasProvider struct {
	inner   Provider
	aliases AliasMap
}

// NewAliasProvider wraps inner so that any path looked up via GetSecret is
// checked against aliases first. If the path (case-insensitive) matches an
// alias key the request is transparently rewritten to the mapped value.
// GetSecretsByPath is always forwarded unchanged.
func NewAliasProvider(inner Provider, aliases AliasMap) (Provider, error) {
	if inner == nil {
		return nil, fmt.Errorf("alias: inner provider must not be nil")
	}
	if aliases == nil {
		aliases = AliasMap{}
	}
	// Normalise keys to upper-case so lookups are case-insensitive.
	norm := make(AliasMap, len(aliases))
	for k, v := range aliases {
		if strings.TrimSpace(v) == "" {
			return nil, fmt.Errorf("alias: target for key %q must not be empty", k)
		}
		norm[strings.ToUpper(k)] = v
	}
	return &aliasProvider{inner: inner, aliases: norm}, nil
}

func (a *aliasProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	resolvedPath, resolvedKey := a.resolve(path, key)
	return a.inner.GetSecret(ctx, resolvedPath, resolvedKey)
}

func (a *aliasProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	return a.inner.GetSecretsByPath(ctx, path)
}

// resolve checks whether path (upper-cased) is an alias. If so it splits the
// mapped value into path and key using splitPathKey; the caller-supplied key
// is used as a fallback when the alias target contains no '#' separator.
func (a *aliasProvider) resolve(path, key string) (string, string) {
	if target, ok := a.aliases[strings.ToUpper(path)]; ok {
		p, k := splitPathKey(target)
		if k == "" {
			k = key
		}
		return p, k
	}
	return path, key
}
