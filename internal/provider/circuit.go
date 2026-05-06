package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreakerConfig holds configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	// MaxFailures is the number of consecutive failures before opening.
	MaxFailures int
	// ResetTimeout is how long to wait before transitioning to half-open.
	ResetTimeout time.Duration
}

type circuitBreaker struct {
	inner      Provider
	cfg        CircuitBreakerConfig
	mu         sync.Mutex
	failures   int
	state      CircuitState
	openedAt   time.Time
}

// NewCircuitBreakerProvider wraps a Provider with a circuit breaker.
// When MaxFailures consecutive errors occur, the circuit opens and
// requests fail fast until ResetTimeout elapses.
func NewCircuitBreakerProvider(inner Provider, cfg CircuitBreakerConfig) (Provider, error) {
	if inner == nil {
		return nil, fmt.Errorf("circuit breaker: inner provider must not be nil")
	}
	if cfg.MaxFailures <= 0 {
		return nil, fmt.Errorf("circuit breaker: MaxFailures must be > 0")
	}
	if cfg.ResetTimeout <= 0 {
		return nil, fmt.Errorf("circuit breaker: ResetTimeout must be > 0")
	}
	return &circuitBreaker{inner: inner, cfg: cfg, state: StateClosed}, nil
}

func (c *circuitBreaker) GetSecret(ctx context.Context, path, key string) (string, error) {
	if err := c.allow(); err != nil {
		return "", err
	}
	val, err := c.inner.GetSecret(ctx, path, key)
	c.record(err)
	return val, err
}

func (c *circuitBreaker) GetSecretsByPath(ctx context.Context, path string) (map[string]string, error) {
	if err := c.allow(); err != nil {
		return nil, err
	}
	vals, err := c.inner.GetSecretsByPath(ctx, path)
	c.record(err)
	return vals, err
}

func (c *circuitBreaker) allow() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch c.state {
	case StateOpen:
		if time.Since(c.openedAt) >= c.cfg.ResetTimeout {
			c.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

func (c *circuitBreaker) record(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil && !IsNotFound(err) {
		c.failures++
		if c.failures >= c.cfg.MaxFailures {
			c.state = StateOpen
			c.openedAt = time.Now()
		}
		return
	}
	c.failures = 0
	c.state = StateClosed
}
