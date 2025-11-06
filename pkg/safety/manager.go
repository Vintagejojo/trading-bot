package safety

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adshao/go-binance/v2"
)

// SafetyManager coordinates all safety mechanisms
type SafetyManager struct {
	circuitBreaker   *CircuitBreaker
	rateLimiter      *RateLimiter
	liquidityChecker *LiquidityChecker
	positionLimits   *PositionLimits
	recoveryManager  *RecoveryManager
	enabled          bool
}

// Config holds all safety configuration
type Config struct {
	Enabled          bool                 `yaml:"enabled"`
	CircuitBreaker   CircuitBreakerConfig `yaml:"circuit_breaker"`
	RateLimit        RateLimitConfig      `yaml:"rate_limit"`
	Liquidity        LiquidityConfig      `yaml:"liquidity"`
	PositionLimits   PositionLimitsConfig `yaml:"position_limits"`
	Recovery         RecoveryConfig       `yaml:"recovery"`
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	MaxFailures  int    `yaml:"max_failures"`
	ResetTimeout string `yaml:"reset_timeout"` // e.g., "5m"
}

// RateLimitConfig holds rate limiter configuration
type RateLimitConfig struct {
	MaxRequests int    `yaml:"max_requests"`
	Interval    string `yaml:"interval"` // e.g., "1m"
}

// NewSafetyManager creates a new safety manager
func NewSafetyManager(client *binance.Client, config Config) (*SafetyManager, error) {
	sm := &SafetyManager{
		enabled: config.Enabled,
	}

	if !config.Enabled {
		log.Println("⚠️  Safety features are DISABLED")
		return sm, nil
	}

	// Initialize circuit breaker
	sm.circuitBreaker = NewCircuitBreaker(
		config.CircuitBreaker.MaxFailures,
		parseDuration(config.CircuitBreaker.ResetTimeout, "5m"),
	)

	sm.circuitBreaker.SetOnStateChange(func(state CircuitState) {
		log.Printf("🔌 Circuit breaker state changed: %s", state)
	})

	// Initialize rate limiter
	sm.rateLimiter = NewRateLimiter(
		config.RateLimit.MaxRequests,
		parseDuration(config.RateLimit.Interval, "1m"),
	)

	// Initialize liquidity checker
	sm.liquidityChecker = NewLiquidityChecker(client, config.Liquidity)

	// Initialize position limits
	sm.positionLimits = NewPositionLimits(client, config.PositionLimits)

	// Initialize recovery manager
	sm.recoveryManager = NewRecoveryManager(config.Recovery)
	sm.recoveryManager.SetOnRecovery(func(attempt int, err error) {
		log.Printf("🔄 Recovery attempt %d: %v", attempt, err)
	})
	sm.recoveryManager.SetOnMaxRetries(func(err error) {
		log.Printf("❌ Max retries exceeded: %v", err)
	})

	log.Println("✅ Safety features enabled")
	return sm, nil
}

// CheckTradeAllowed verifies if a trade is allowed by all safety checks
func (sm *SafetyManager) CheckTradeAllowed(ctx context.Context, symbol string, quantity float64, price float64, side string) error {
	if !sm.enabled {
		return nil
	}

	// Check circuit breaker
	if sm.circuitBreaker.IsOpen() {
		return fmt.Errorf("circuit breaker is open - trading paused")
	}

	// Check rate limit
	if err := sm.rateLimiter.TryAllow(); err != nil {
		return err
	}

	// Check daily loss limit
	if sm.positionLimits.IsDailyLimitReached() {
		return fmt.Errorf("daily loss limit reached")
	}

	// Check position size limits
	if err := sm.positionLimits.CheckPositionSize(ctx, symbol, quantity, price); err != nil {
		return fmt.Errorf("position size check failed: %w", err)
	}

	// Check liquidity
	if err := sm.liquidityChecker.CheckLiquidity(ctx, symbol, quantity, side); err != nil {
		return fmt.Errorf("liquidity check failed: %w", err)
	}

	return nil
}

// ExecuteWithSafety executes a function with all safety mechanisms
func (sm *SafetyManager) ExecuteWithSafety(fn func() error) error {
	if !sm.enabled {
		return fn()
	}

	// Use circuit breaker
	return sm.circuitBreaker.Call(func() error {
		// Use recovery manager
		return sm.recoveryManager.Retry(fn)
	})
}

// RecordTrade records a completed trade for tracking
func (sm *SafetyManager) RecordTrade(profitLoss float64, isProfit bool) {
	if !sm.enabled {
		return
	}

	if isProfit {
		sm.positionLimits.RecordProfit(profitLoss)
	} else {
		sm.positionLimits.RecordLoss(profitLoss)
	}
}

// OpenPosition increments the open position counter
func (sm *SafetyManager) OpenPosition() {
	if sm.enabled {
		sm.positionLimits.IncrementPosition()
	}
}

// ClosePosition decrements the open position counter
func (sm *SafetyManager) ClosePosition() {
	if sm.enabled {
		sm.positionLimits.DecrementPosition()
	}
}

// ResetDailyLimits resets daily tracking (call at start of new day)
func (sm *SafetyManager) ResetDailyLimits() {
	if sm.enabled {
		sm.positionLimits.ResetDailyLoss()
		log.Println("🔄 Daily limits reset")
	}
}

// GetStatus returns current safety status
func (sm *SafetyManager) GetStatus() map[string]interface{} {
	if !sm.enabled {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	return map[string]interface{}{
		"enabled":           true,
		"circuit_breaker":   sm.circuitBreaker.GetState().String(),
		"rate_limit_tokens": sm.rateLimiter.GetAvailableTokens(),
		"daily_loss":        sm.positionLimits.GetCurrentDailyLoss(),
		"open_positions":    sm.positionLimits.GetOpenPositions(),
	}
}

// Helper function to parse duration strings
func parseDuration(s string, defaultValue string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		d, _ = time.ParseDuration(defaultValue)
	}
	return d
}
