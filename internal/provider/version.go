package provider

import (
	"context"
	"fmt"
	"strings"
)

// VersionedProvider wraps an inner Provider and appends a version suffix to
// every secret path before delegating, enabling callers to pin secrets to a
// specific version label (e.g. "v2", "prod").
//
// The version string is appended as a path segment:
//
//	"secret/myapp/db" + version "v2" → "secret/myapp/db/v2"
//
// GetSecretsByPath strips the version suffix from returned keys so that the
// caller sees the original key names.
type versionedProvider struct {
	inner   Provider
	version string
}

// NewVersionedProvider returns a Provider that pins all lookups to version.
// version must be non-empty and must not contain a forward slash.
func NewVersionedProvider(inner Provider, version string) (Provider, error) {
	if inner == nil {
		return nil, fmt.Errorf("versionedProvider: inner provider must not be nil")
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return nil, fmt.Errorf("versionedProvider: version must not be empty")
	}
	if strings.Contains(version, "/") {
		return nil, fmt.Errorf("versionedProvider: version must not contain '/'")
	}
	return &versionedProvider{inner: inner, version: version}, nil
}

func (v *versionedProvider) GetSecret(ctx context.Context, path string) (string, error) {
	versionedPath := v.addVersion(path)
	return v.inner.GetSecret(ctx, versionedPath)
}

func (v *versionedProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	versionedPath := v.addVersion(path)
	results, err := v.inner.GetSecretsByPath(ctx, versionedPath)
	if err != nil {
		return nil, err
	}

	// Strip the version suffix from each key so callers see clean names.
	suffix := "/" + v.version
	stripped := make(map[string]string, len(results))
	for k, val := range results {
		stripped[strings.TrimSuffix(k, suffix)] = val
	}
	return stripped, nil
}

func (v *versionedProvider) addVersion(path string) string {
	path = strings.TrimRight(path, "/")
	return path + "/" + v.version
}
