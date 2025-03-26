package analysis

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter controls the rate of API requests to external services
type RateLimiter struct {
	maxRequests int
	requests    []time.Time
	mu          sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the specified request limit per minute
func NewRateLimiter(maxRequestsPerMinute int) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequestsPerMinute,
		requests:    make([]time.Time, 0, maxRequestsPerMinute),
	}
}

// Acquire tries to acquire permission to make a request, blocking if necessary
func (r *RateLimiter) Acquire(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := r.tryAcquire(); err == nil {
				return nil
			}
			// Wait before retrying
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				// Continue and try again
			}
		}
	}
}

// tryAcquire attempts to acquire a rate limit token without blocking
func (r *RateLimiter) tryAcquire() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	// Clean old requests
	r.requests = filterOldRequests(r.requests, cutoff)

	// Check if we're at limit
	if len(r.requests) >= r.maxRequests {
		return fmt.Errorf("rate limit exceeded")
	}

	r.requests = append(r.requests, now)
	return nil
}

// filterOldRequests removes requests older than the cutoff time
func filterOldRequests(requests []time.Time, cutoff time.Time) []time.Time {
	filtered := make([]time.Time, 0, len(requests))
	for _, req := range requests {
		if req.After(cutoff) {
			filtered = append(filtered, req)
		}
	}
	return filtered
} 