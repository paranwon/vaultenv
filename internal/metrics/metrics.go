// Package metrics provides lightweight in-process counters and histograms
// for tracking secret fetch latency, cache hit/miss rates, and provider errors.
package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Counter is a monotonically increasing integer counter.
type Counter struct{ n uint64 }

func (c *Counter) Inc()          { atomic.AddUint64(&c.n, 1) }
func (c *Counter) Value() uint64 { return atomic.LoadUint64(&c.n) }

// Histogram records a distribution of float64 durations (milliseconds).
type Histogram struct {
	mu      sync.Mutex
	samples []float64
}

func (h *Histogram) Observe(d time.Duration) {
	h.mu.Lock()
	h.samples = append(h.samples, float64(d.Milliseconds()))
	h.mu.Unlock()
}

func (h *Histogram) Snapshot() []float64 {
	h.mu.Lock()
	out := make([]float64, len(h.samples))
	copy(out, h.samples)
	h.mu.Unlock()
	return out
}

// Registry holds all named metrics for the process.
type Registry struct {
	mu         sync.RWMutex
	counters    map[string]*Counter
	histograms  map[string]*Histogram
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{
		counters:   make(map[string]*Counter),
		histograms: make(map[string]*Histogram),
	}
}

// Counter returns (or creates) the named counter.
func (r *Registry) Counter(name string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.counters[name]; ok {
		return c
	}
	c := &Counter{}
	r.counters[name] = c
	return c
}

// Histogram returns (or creates) the named histogram.
func (r *Registry) Histogram(name string) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.histograms[name]; ok {
		return h
	}
	h := &Histogram{}
	r.histograms[name] = h
	return h
}

// Snapshot returns a point-in-time copy of all counter values.
func (r *Registry) Snapshot() map[string]uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]uint64, len(r.counters))
	for k, c := range r.counters {
		out[k] = c.Value()
	}
	return out
}
