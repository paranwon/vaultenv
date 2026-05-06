package provider_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/your-org/vaultenv/internal/provider"
)

type countingProvider struct {
	calls atomic.Int64
	data  map[string]string
}

func (c *countingProvider) GetSecret(_ context.Context, path, key string) (string, error) {
	c.calls.Add(1)
	if v, ok := c.data[path+"#"+key]; ok {
		return v, nil
	}
	return "", provider.ErrNotFound{Key: key}
}

func (c *countingProvider) GetSecretsByPath(_ context.Context, _ string) (map[string]string, error) {
	return nil, errors.New("not implemented")
}

func TestBatchProvider_InvalidArgs(t *testing.T) {
	_, err := provider.NewBatchProvider(nil, 4)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}

	p := &countingProvider{data: map[string]string{}}
	_, err = provider.NewBatchProvider(p, 0)
	if err == nil {
		t.Fatal("expected error for zero workers")
	}
}

func TestBatchProvider_FetchAll_AllFound(t *testing.T) {
	data := map[string]string{
		"secret/app#db_pass": "hunter2",
		"secret/app#api_key": "abc123",
		"secret/app#token":   "tok999",
	}
	cp := &countingProvider{data: data}
	bp, err := provider.NewBatchProvider(cp, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	refs := []provider.SecretRef{
		{Path: "secret/app", Key: "db_pass"},
		{Path: "secret/app", Key: "api_key"},
		{Path: "secret/app", Key: "token"},
	}

	results := bp.FetchAll(context.Background(), refs)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] unexpected error: %v", i, r.Err)
		}
		want := data[refs[i].Path+"#"+refs[i].Key]
		if r.Value != want {
			t.Errorf("result[%d] value = %q, want %q", i, r.Value, want)
		}
	}
	if cp.calls.Load() != 3 {
		t.Errorf("expected 3 provider calls, got %d", cp.calls.Load())
	}
}

func TestBatchProvider_FetchAll_PartialMiss(t *testing.T) {
	cp := &countingProvider{data: map[string]string{
		"secret/svc#key": "val",
	}}
	bp, _ := provider.NewBatchProvider(cp, 2)

	refs := []provider.SecretRef{
		{Path: "secret/svc", Key: "key"},
		{Path: "secret/svc", Key: "missing"},
	}
	results := bp.FetchAll(context.Background(), refs)
	if results[0].Err != nil {
		t.Errorf("expected no error for first ref, got %v", results[0].Err)
	}
	if !provider.IsNotFound(results[1].Err) {
		t.Errorf("expected not-found for second ref, got %v", results[1].Err)
	}
}

func TestBatchProvider_FetchAll_WorkerBound(t *testing.T) {
	const total = 20
	data := make(map[string]string, total)
	refs := make([]provider.SecretRef, total)
	for i := 0; i < total; i++ {
		k := fmt.Sprintf("k%d", i)
		data["p#"+k] = "v"
		refs[i] = provider.SecretRef{Path: "p", Key: k}
	}
	cp := &countingProvider{data: data}
	bp, _ := provider.NewBatchProvider(cp, 4)
	results := bp.FetchAll(context.Background(), refs)
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] error: %v", i, r.Err)
		}
	}
	if cp.calls.Load() != total {
		t.Errorf("expected %d calls, got %d", total, cp.calls.Load())
	}
}
