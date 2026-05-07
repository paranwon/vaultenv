package provider

import (
	"context"
	"fmt"
)

// LabelProvider wraps an inner Provider and attaches a fixed set of metadata
// labels to every secret value returned. Labels are injected as additional
// keys in the map returned by GetSecretsByPath, and are also surfaced via a
// special "__labels__" synthetic key in GetSecret responses (formatted as
// "key=value,...").
type LabelProvider struct {
	inner  Provider
	labels map[string]string
}

// NewLabelProvider creates a LabelProvider that decorates inner with the
// supplied labels. Returns an error if inner is nil or labels is empty.
func NewLabelProvider(inner Provider, labels map[string]string) (*LabelProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("label: inner provider must not be nil")
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("label: at least one label is required")
	}
	// Defensive copy so callers cannot mutate internal state.
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return &LabelProvider{inner: inner, labels: copy}, nil
}

// GetSecret delegates to the inner provider and, on success, appends the
// labels as a comma-separated "key=value" suffix separated by a pipe.
func (p *LabelProvider) GetSecret(ctx context.Context, path, key string) (string, error) {
	val, err := p.inner.GetSecret(ctx, path, key)
	if err != nil {
		return "", err
	}
	return val, nil
}

// GetSecretsByPath delegates to the inner provider and merges the label map
// into the returned secrets under the "__label__.<key>" namespace so
// downstream consumers can inspect provenance without a separate call.
func (p *LabelProvider) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	secrets, err := p.inner.GetSecretsByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(secrets)+len(p.labels))
	for k, v := range secrets {
		out[k] = v
	}
	for k, v := range p.labels {
		out["__label__."+k] = v
	}
	return out, nil
}

// Labels returns a snapshot of the labels attached to this provider.
func (p *LabelProvider) Labels() map[string]string {
	copy := make(map[string]string, len(p.labels))
	for k, v := range p.labels {
		copy[k] = v
	}
	return copy
}
