// Package ratelimit provides a token-bucket rate limiter for provider calls.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Policy configures the rate limiter.
type Policy struct {
	// RequestsPerSecond is the sustained rate of allowed requests.
	RequestsPerSecond float64
	// Burst is the maximum number of requests allowed in a single instant.
	Burst int
}

// Limiter is a token-bucket rate limiter.
type Limiter struct {
	mu     sync.Mutex
	tokens float64
	max    float64
	rate   float64 // tokens per nanosecond
	last   time.Time
	now    func() time.Time
}

// New creates a Limiter from the given Policy.
func New(p Policy) (*Limiter, error) {
	if p.RequestsPerSecond <= 0 {
		return nil, fmt.Errorf("ratelimit: RequestsPerSecond must be > 0, got %v", p.RequestsPerSecond)
	}
	if p.Burst <= 0 {
		return nil, fmt.Errorf("ratelimit: Burst must be > 0, got %v", p.Burst)
	}
	return &Limiter{
		tokens: float64(p.Burst),
		max:    float64(p.Burst),
		rate:   p.RequestsPerSecond / 1e9,
		last:   time.Now(),
		now:    time.Now,
	}, nil
}

// Wait blocks until a token is available or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := l.now()
		elapsed := now.Sub(l.last).Nanoseconds()
		l.tokens += float64(elapsed) * l.rate
		if l.tokens > l.max {
			l.tokens = l.max
		}
		l.last = now
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		// Calculate how long until next token.
		waitNs := time.Duration((1 - l.tokens) / l.rate)
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitNs):
		}
	}
}
