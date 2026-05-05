package metrics_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultenv/internal/metrics"
)

func TestCounter_IncAndValue(t *testing.T) {
	r := metrics.New()
	c := r.Counter("fetches")
	if c.Value() != 0 {
		t.Fatalf("expected 0, got %d", c.Value())
	}
	c.Inc()
	c.Inc()
	if c.Value() != 2 {
		t.Fatalf("expected 2, got %d", c.Value())
	}
}

func TestCounter_SameName_ReturnsSameInstance(t *testing.T) {
	r := metrics.New()
	a := r.Counter("hits")
	b := r.Counter("hits")
	a.Inc()
	if b.Value() != 1 {
		t.Fatal("expected same counter instance")
	}
}

func TestHistogram_ObserveAndSnapshot(t *testing.T) {
	r := metrics.New()
	h := r.Histogram("latency_ms")
	h.Observe(10 * time.Millisecond)
	h.Observe(20 * time.Millisecond)
	snap := h.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(snap))
	}
	if snap[0] != 10 || snap[1] != 20 {
		t.Fatalf("unexpected samples: %v", snap)
	}
}

func TestHistogram_SnapshotIsCopy(t *testing.T) {
	r := metrics.New()
	h := r.Histogram("lat")
	h.Observe(5 * time.Millisecond)
	snap := h.Snapshot()
	snap[0] = 999
	snap2 := h.Snapshot()
	if snap2[0] == 999 {
		t.Fatal("snapshot should be an independent copy")
	}
}

func TestRegistry_Snapshot_CounterValues(t *testing.T) {
	r := metrics.New()
	r.Counter("a").Inc()
	r.Counter("a").Inc()
	r.Counter("b").Inc()
	snap := r.Snapshot()
	if snap["a"] != 2 {
		t.Fatalf("expected a=2, got %d", snap["a"])
	}
	if snap["b"] != 1 {
		t.Fatalf("expected b=1, got %d", snap["b"])
	}
}
