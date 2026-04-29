package template_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultenv/internal/template"
)

type fakeResolver struct {
	secrets map[string]string
}

func (f *fakeResolver) GetSecret(_ context.Context, provider, path string) (string, error) {
	key := provider + ":" + path
	v, ok := f.secrets[key]
	if !ok {
		return "", errors.New("not found: " + key)
	}
	return v, nil
}

func TestExpand_SingleRef(t *testing.T) {
	r := &fakeResolver{secrets: map[string]string{"vault:secret/data/app#password": "s3cr3t"}}
	got, err := template.Expand(context.Background(), "pass={{ vault:secret/data/app#password }}", r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "pass=s3cr3t" {
		t.Errorf("got %q, want %q", got, "pass=s3cr3t")
	}
}

func TestExpand_MultipleRefs(t *testing.T) {
	r := &fakeResolver{secrets: map[string]string{
		"ssm:/app/user": "admin",
		"ssm:/app/pass": "hunter2",
	}}
	input := "{{ ssm:/app/user }}:{{ ssm:/app/pass }}"
	got, err := template.Expand(context.Background(), input, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "admin:hunter2" {
		t.Errorf("got %q", got)
	}
}

func TestExpand_NoRefs(t *testing.T) {
	r := &fakeResolver{}
	got, err := template.Expand(context.Background(), "plain string", r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "plain string" {
		t.Errorf("got %q", got)
	}
}

func TestExpand_MissingSecret_ReturnsError(t *testing.T) {
	r := &fakeResolver{secrets: map[string]string{}}
	_, err := template.Expand(context.Background(), "x={{ vault:missing#key }}", r)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHasRefs(t *testing.T) {
	if !template.HasRefs("{{ vault:secret/data/x#k }}") {
		t.Error("expected HasRefs true")
	}
	if template.HasRefs("no refs here") {
		t.Error("expected HasRefs false")
	}
}

func TestListRefs_Deduplicates(t *testing.T) {
	s := "{{ ssm:/a }} and {{ ssm:/a }} and {{ ssm:/b }}"
	refs := template.ListRefs(s)
	if len(refs) != 2 {
		t.Errorf("expected 2 unique refs, got %d: %v", len(refs), refs)
	}
}
