package safety

import (
	"fmt"
	"log"
	"time"
)

// RecoveryStrategy defines how to recover from failures
type RecoveryStrategy int

const (
	RecoveryImmediate RecoveryStrategy = iota // Retry immediately
	RecoveryLinear                             // Linear backoff
	RecoveryExponential                        // Exponential backoff
)

// RecoveryManager handles automatic recovery from errors
type RecoveryManager struct {
	strategy      RecoveryStrategy
	maxRetries    int
	baseDelay     time.Duration
	maxDelay      time.Duration
	onRecovery    func(attempt int, err error)
	onMaxRetries  func(err error)
}

// RecoveryConfig holds configuration for recovery manager
type RecoveryConfig struct {
	Strategy      string        `yaml:"strategy"`       // "immediate", "linear", "exponential"
	MaxRetries    int           `yaml:"max_retries"`
	BaseDelay     time.Duration `yaml:"base_delay"`
	MaxDelay      time.Duration `yaml:"max_delay"`
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(config RecoveryConfig) *RecoveryManager {
	strategy := RecoveryExponential
	switch config.Strategy {
	case "immediate":
		strategy = RecoveryImmediate
	case "linear":
		strategy = RecoveryLinear
	case "exponential":
		strategy = RecoveryExponential
	}

	return &RecoveryManager{
		strategy:   strategy,
		maxRetries: config.MaxRetries,
		baseDelay:  config.BaseDelay,
		maxDelay:   config.MaxDelay,
	}
}

// SetOnRecovery sets callback for recovery attempts
func (rm *RecoveryManager) SetOnRecovery(fn func(attempt int, err error)) {
	rm.onRecovery = fn
}

// SetOnMaxRetries sets callback for when max retries is reached
func (rm *RecoveryManager) SetOnMaxRetries(fn func(err error)) {
	rm.onMaxRetries = fn
}

// Retry attempts to execute a function with retry logic
func (rm *RecoveryManager) Retry(fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rm.maxRetries; attempt++ {
		// Try the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Max retries reached
		if attempt == rm.maxRetries {
			if rm.onMaxRetries != nil {
				rm.onMaxRetries(err)
			}
			return fmt.Errorf("max retries (%d) exceeded: %w", rm.maxRetries, err)
		}

		// Calculate delay
		delay := rm.calculateDelay(attempt)

		// Notify recovery attempt
		if rm.onRecovery != nil {
			rm.onRecovery(attempt+1, err)
		}

		log.Printf("🔄 Recovery attempt %d/%d after error: %v (waiting %v)",
			attempt+1, rm.maxRetries, err, delay)

		// Wait before retry
		time.Sleep(delay)
	}

	return lastErr
}

// calculateDelay determines the delay before next retry
func (rm *RecoveryManager) calculateDelay(attempt int) time.Duration {
	var delay time.Duration

	switch rm.strategy {
	case RecoveryImmediate:
		delay = 0

	case RecoveryLinear:
		delay = rm.baseDelay * time.Duration(attempt+1)

	case RecoveryExponential:
		delay = rm.baseDelay * (1 << attempt) // 2^attempt
	}

	// Cap at maximum delay
	if delay > rm.maxDelay {
		delay = rm.maxDelay
	}

	return delay
}

// RetryWithContext retries with context cancellation support
func (rm *RecoveryManager) RetryWithContext(fn func() error, stopChan <-chan struct{}) error {
	var lastErr error

	for attempt := 0; attempt <= rm.maxRetries; attempt++ {
		// Check if context cancelled
		select {
		case <-stopChan:
			return fmt.Errorf("recovery cancelled")
		default:
		}

		// Try the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Max retries reached
		if attempt == rm.maxRetries {
			if rm.onMaxRetries != nil {
				rm.onMaxRetries(err)
			}
			return fmt.Errorf("max retries (%d) exceeded: %w", rm.maxRetries, err)
		}

		// Calculate delay
		delay := rm.calculateDelay(attempt)

		// Notify recovery attempt
		if rm.onRecovery != nil {
			rm.onRecovery(attempt+1, err)
		}

		log.Printf("🔄 Recovery attempt %d/%d after error: %v (waiting %v)",
			attempt+1, rm.maxRetries, err, delay)

		// Wait before retry with cancellation support
		select {
		case <-time.After(delay):
		case <-stopChan:
			return fmt.Errorf("recovery cancelled during backoff")
		}
	}

	return lastErr
}
