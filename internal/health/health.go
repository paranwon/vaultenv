// Package health provides readiness and liveness checks for secret providers.
package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents the health state of a component.
type Status string

const (
	StatusOK      Status = "ok"
	StatusDegraded Status = "degraded"
	StatusDown    Status = "down"
)

// Result holds the outcome of a single health check.
type Result struct {
	Name    string        `json:"name"`
	Status  Status        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency_ms"`
}

// Checker is a named function that reports whether a dependency is healthy.
type Checker struct {
	Name string
	Check func(ctx context.Context) error
}

// Registry holds a collection of health checkers and runs them on demand.
type Registry struct {
	mu       sync.RWMutex
	checkers []Checker
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{}
}

// Register adds a named checker to the registry.
func (r *Registry) Register(c Checker) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers = append(r.checkers, c)
}

// RunAll executes every registered checker concurrently and returns all results.
// The overall status is StatusOK only when every individual check passes.
func (r *Registry) RunAll(ctx context.Context) ([]Result, Status) {
	r.mu.RLock()
	checkers := make([]Checker, len(r.checkers))
	copy(checkers, r.checkers)
	r.mu.RUnlock()

	results := make([]Result, len(checkers))
	var wg sync.WaitGroup

	for i, c := range checkers {
		wg.Add(1)
		go func(idx int, chk Checker) {
			defer wg.Done()
			start := time.Now()
			err := chk.Check(ctx)
			latency := time.Since(start)
			res := Result{Name: chk.Name, Latency: latency}
			if err != nil {
				res.Status = StatusDown
				res.Message = err.Error()
			} else {
				res.Status = StatusOK
			}
			results[idx] = res
		}(i, c)
	}

	wg.Wait()

	overall := StatusOK
	for _, res := range results {
		if res.Status != StatusOK {
			overall = StatusDegraded
			break
		}
	}
	_ = fmt.Sprintf // satisfy import
	return results, overall
}
