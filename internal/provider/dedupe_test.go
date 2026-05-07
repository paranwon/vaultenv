package provider_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/vaultenv/internal/provider"
)

func TestDedupeProvider_NilInner(t *testing.T) {
	_, err := provider.NewDedupeProvider(nil)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestDedupeProvider_GetSecret_SingleCall(t *testing.T) {
	var calls int32
	stub := &stubProvider{
		getSecret: func(_ context.Context, path string) (map[string]string, error) {
			atomic.AddInt32(&calls, 1)
			return map[string]string{"value": "s3cr3t"}, nil
		},
	}
	dp, err := provider.NewDedupeProvider(stub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	val, err := dp.GetSecret(context.Background(), "secret/foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val["value"] != "s3cr3t" {
		t.Errorf("unexpected value: %v", val)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 upstream call, got %d", calls)
	}
}

func TestDedupeProvider_GetSecret_CoalescesConcurrent(t *testing.T) {
	var calls int32
	block := make(chan struct{})
	stub := &stubProvider{
		getSecret: func(_ context.Context, path string) (map[string]string, error) {
			atomic.AddInt32(&calls, 1)
			<-block
			return map[string]string{"value": "shared"}, nil
		},
	}
	dp, _ := provider.NewDedupeProvider(stub)

	const n = 10
	var wg sync.WaitGroup
	results := make([]map[string]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			v, _ := dp.GetSecret(context.Background(), "secret/shared")
			results[idx] = v
		}(i)
	}
	time.Sleep(20 * time.Millisecond) // let goroutines queue
	close(block)
	wg.Wait()

	if c := atomic.LoadInt32(&calls); c != 1 {
		t.Errorf("expected exactly 1 upstream call, got %d", c)
	}
	for i, r := range results {
		if r["value"] != "shared" {
			t.Errorf("goroutine %d got unexpected result: %v", i, r)
		}
	}
}

func TestDedupeProvider_GetSecret_ContextCancelled(t *testing.T) {
	block := make(chan struct{})
	stub := &stubProvider{
		getSecret: func(_ context.Context, _ string) (map[string]string, error) {
			<-block
			return map[string]string{"value": "late"}, nil
		},
	}
	dp, _ := provider.NewDedupeProvider(stub)

	// Start a blocking call.
	go dp.GetSecret(context.Background(), "secret/slow") //nolint:errcheck
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := dp.GetSecret(ctx, "secret/slow")
	close(block)
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func TestDedupeProvider_GetSecretsByPath_Delegates(t *testing.T) {
	stub := &stubProvider{
		getSecretsByPath: func(_ context.Context, path string) (map[string]string, error) {
			return map[string]string{"k": "v"}, nil
		},
	}
	dp, _ := provider.NewDedupeProvider(stub)
	val, err := dp.GetSecretsByPath(context.Background(), "secret/")
	if err != nil || val["k"] != "v" {
		t.Errorf("unexpected result: %v %v", val, err)
	}
}
