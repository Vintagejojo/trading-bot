package safety

import (
	"fmt"
	"sync"
	"time"
)

// CircuitState represents the current state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota // Normal operation
	StateOpen                        // Circuit is open, rejecting requests
	StateHalfOpen                    // Testing if system recovered
)

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures
type CircuitBreaker struct {
	maxFailures    int           // Max failures before opening circuit
	resetTimeout   time.Duration // Time to wait before attempting recovery
	failureCount   int           // Current failure count
	lastFailTime   time.Time     // Time of last failure
	state          CircuitState  // Current circuit state
	mu             sync.RWMutex  // Protects circuit state
	onStateChange  func(CircuitState)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// SetOnStateChange sets a callback for state changes
func (cb *CircuitBreaker) SetOnStateChange(fn func(CircuitState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.onStateChange = fn
}

// Call executes the given function if the circuit allows it
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check if we should attempt recovery
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.setState(StateHalfOpen)
		} else {
			cb.mu.Unlock()
			return fmt.Errorf("circuit breaker is open")
		}
	}

	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onFailure handles a failed call
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.state == StateHalfOpen {
		// Failed during recovery attempt
		cb.setState(StateOpen)
		return
	}

	if cb.failureCount >= cb.maxFailures {
		cb.setState(StateOpen)
	}
}

// onSuccess handles a successful call
func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		// Recovery successful
		cb.setState(StateClosed)
	}
	cb.failureCount = 0
}

// setState changes the circuit state and triggers callback
func (cb *CircuitBreaker) setState(newState CircuitState) {
	if cb.state != newState {
		cb.state = newState
		if cb.onStateChange != nil {
			go cb.onStateChange(newState)
		}
	}
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// IsOpen returns true if circuit is open
func (cb *CircuitBreaker) IsOpen() bool {
	return cb.GetState() == StateOpen
}

// Reset manually resets the circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.setState(StateClosed)
}

// String returns the string representation of circuit state
func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}
