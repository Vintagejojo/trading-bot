package safety

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	maxRequests int           // Maximum requests allowed
	interval    time.Duration // Time window
	tokens      int           // Current available tokens
	lastRefill  time.Time     // Last time tokens were refilled
	mu          sync.Mutex    // Protects token count
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxRequests int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		interval:    interval,
		tokens:      maxRequests,
		lastRefill:  time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// Wait blocks until a request can be made
func (rl *RateLimiter) Wait() error {
	for !rl.Allow() {
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// TryAllow attempts to allow a request and returns error if denied
func (rl *RateLimiter) TryAllow() error {
	if !rl.Allow() {
		return fmt.Errorf("rate limit exceeded: %d requests per %v", rl.maxRequests, rl.interval)
	}
	return nil
}

// refill adds tokens based on elapsed time
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.interval {
		rl.tokens = rl.maxRequests
		rl.lastRefill = now
	}
}

// GetAvailableTokens returns the current number of available tokens
func (rl *RateLimiter) GetAvailableTokens() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refill()
	return rl.tokens
}

// Reset resets the rate limiter to full capacity
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.tokens = rl.maxRequests
	rl.lastRefill = time.Now()
}
