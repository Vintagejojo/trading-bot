package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"rsi-bot/pkg/safety"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("🧪 Testing Safety Features - Phase 7.5")
	log.Println("=========================================")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	if apiKey == "" || apiSecret == "" {
		log.Fatal("❌ BINANCE_API_KEY and BINANCE_API_SECRET must be set")
	}

	client := binance.NewClient(apiKey, apiSecret)
	client.BaseURL = "https://api.binance.us"

	// Test 1: Circuit Breaker
	testCircuitBreaker()

	// Test 2: Rate Limiter
	testRateLimiter()

	// Test 3: Liquidity Checker
	testLiquidityChecker(client)

	// Test 4: Position Limits
	testPositionLimits(client)

	// Test 5: Recovery Manager
	testRecoveryManager()

	// Test 6: Safety Manager (Integration)
	testSafetyManager(client)

	log.Println("\n✅ All safety feature tests completed!")
}

func testCircuitBreaker() {
	log.Println("\n--- Test 1: Circuit Breaker ---")

	cb := safety.NewCircuitBreaker(3, 2*time.Second)

	// Set state change callback
	cb.SetOnStateChange(func(state safety.CircuitState) {
		log.Printf("🔌 Circuit state changed: %s", state)
	})

	// Test normal operation (should succeed)
	err := cb.Call(func() error {
		log.Println("  ✓ Normal operation")
		return nil
	})
	if err != nil {
		log.Printf("  ❌ Unexpected error: %v", err)
	}

	// Trigger failures to open circuit
	log.Println("  Triggering failures to open circuit...")
	for i := 0; i < 4; i++ {
		err := cb.Call(func() error {
			return fmt.Errorf("simulated failure %d", i+1)
		})
		if err != nil {
			log.Printf("  ⚠️  Failure %d: %v", i+1, err)
		}
	}

	// Circuit should be open now
	if cb.IsOpen() {
		log.Println("  ✅ Circuit breaker opened as expected")
	} else {
		log.Println("  ❌ Circuit breaker should be open")
	}

	// Try to make a call (should be rejected)
	err = cb.Call(func() error {
		log.Println("  ❌ This should not execute")
		return nil
	})
	if err != nil {
		log.Printf("  ✅ Call rejected (circuit open): %v", err)
	}

	// Wait for reset timeout
	log.Println("  Waiting for circuit reset timeout (2s)...")
	time.Sleep(3 * time.Second)

	// Try recovery
	err = cb.Call(func() error {
		log.Println("  ✓ Recovery attempt")
		return nil
	})
	if err == nil {
		log.Println("  ✅ Circuit recovered successfully")
	} else {
		log.Printf("  ❌ Recovery failed: %v", err)
	}
}

func testRateLimiter() {
	log.Println("\n--- Test 2: Rate Limiter ---")

	rl := safety.NewRateLimiter(5, 2*time.Second)

	// Use all tokens
	log.Println("  Using all 5 tokens...")
	for i := 0; i < 5; i++ {
		if rl.Allow() {
			log.Printf("  ✓ Request %d allowed", i+1)
		}
	}

	// Next request should be denied
	if !rl.Allow() {
		log.Println("  ✅ 6th request denied (rate limit exceeded)")
	} else {
		log.Println("  ❌ Rate limit should have been exceeded")
	}

	// Check available tokens
	tokens := rl.GetAvailableTokens()
	log.Printf("  Available tokens: %d", tokens)

	// Wait for refill
	log.Println("  Waiting for token refill (2s)...")
	time.Sleep(3 * time.Second)

	tokens = rl.GetAvailableTokens()
	log.Printf("  ✅ Tokens refilled: %d", tokens)
}

func testLiquidityChecker(client *binance.Client) {
	log.Println("\n--- Test 3: Liquidity Checker ---")

	config := safety.LiquidityConfig{
		MinOrderBookDepth:   10,
		MinTotalVolume:      100000,
		MaxSpreadPercent:    0.5,
		MinVolumeMultiplier: 0.1,
	}

	lc := safety.NewLiquidityChecker(client, config)

	// Test with RVNUSD (your actual holding)
	symbol := "RVNUSD"
	orderSize := 1000.0

	log.Printf("  Checking liquidity for %s (order size: %.2f)...", symbol, orderSize)

	err := lc.CheckLiquidity(context.Background(), symbol, orderSize, "BUY")
	if err != nil {
		log.Printf("  ⚠️  Liquidity check failed: %v", err)
		log.Println("  (This is expected for low-volume pairs)")
	} else {
		log.Println("  ✅ Liquidity check passed")
	}

	// Get market depth info
	bid, ask, spread, err := lc.GetMarketDepth(context.Background(), symbol)
	if err != nil {
		log.Printf("  ⚠️  Failed to get market depth: %v", err)
	} else {
		log.Printf("  📊 Market depth: Bid=%.8f, Ask=%.8f, Spread=%.4f%%", bid, ask, spread)
	}
}

func testPositionLimits(client *binance.Client) {
	log.Println("\n--- Test 4: Position Limits ---")

	config := safety.PositionLimitsConfig{
		MaxPositionSizeUSD:  1000,
		MaxPortfolioPercent: 20,
		MaxDailyLossUSD:     100,
		MaxTotalPositions:   3,
	}

	pl := safety.NewPositionLimits(client, config)

	// Test position size check
	symbol := "RVNUSD"
	quantity := 10000.0 // Large quantity
	price := 0.01

	log.Printf("  Checking position size: %.0f %s @ %.4f = $%.2f", quantity, symbol, price, quantity*price)

	err := pl.CheckPositionSize(context.Background(), symbol, quantity, price)
	if err != nil {
		log.Printf("  ✅ Large position rejected: %v", err)
	} else {
		log.Println("  ⚠️  Position should have been rejected")
	}

	// Test smaller position
	quantity = 5000.0
	log.Printf("  Checking smaller position: %.0f %s @ %.4f = $%.2f", quantity, symbol, price, quantity*price)

	err = pl.CheckPositionSize(context.Background(), symbol, quantity, price)
	if err != nil {
		log.Printf("  ⚠️  Position check failed: %v", err)
	} else {
		log.Println("  ✅ Reasonable position size approved")
	}

	// Test daily loss tracking
	log.Println("  Testing daily loss tracking...")
	pl.RecordLoss(50.0)
	log.Printf("  Recorded loss: $50.00, Current daily loss: $%.2f", pl.GetCurrentDailyLoss())

	pl.RecordLoss(60.0)
	log.Printf("  Recorded loss: $60.00, Current daily loss: $%.2f", pl.GetCurrentDailyLoss())

	if pl.IsDailyLimitReached() {
		log.Println("  ✅ Daily loss limit reached as expected")
	} else {
		log.Println("  ❌ Daily loss limit should be reached")
	}

	// Test position counting
	log.Println("  Testing position counting...")
	pl.IncrementPosition()
	pl.IncrementPosition()
	log.Printf("  Open positions: %d", pl.GetOpenPositions())

	pl.DecrementPosition()
	log.Printf("  ✅ After closing one: %d positions", pl.GetOpenPositions())
}

func testRecoveryManager() {
	log.Println("\n--- Test 5: Recovery Manager ---")

	config := safety.RecoveryConfig{
		Strategy:   "exponential",
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   2 * time.Second,
	}

	rm := safety.NewRecoveryManager(config)

	// Test successful retry
	log.Println("  Testing recovery with eventual success...")
	attempt := 0
	err := rm.Retry(func() error {
		attempt++
		if attempt < 3 {
			return fmt.Errorf("simulated failure (attempt %d)", attempt)
		}
		log.Println("  ✓ Operation succeeded on attempt 3")
		return nil
	})

	if err == nil {
		log.Println("  ✅ Recovery succeeded after retries")
	} else {
		log.Printf("  ❌ Recovery failed: %v", err)
	}

	// Test max retries exceeded
	log.Println("  Testing max retries exceeded...")
	err = rm.Retry(func() error {
		return fmt.Errorf("persistent failure")
	})

	if err != nil {
		log.Printf("  ✅ Gave up after max retries: %v", err)
	} else {
		log.Println("  ❌ Should have failed after max retries")
	}
}

func testSafetyManager(client *binance.Client) {
	log.Println("\n--- Test 6: Safety Manager (Integration) ---")

	config := safety.Config{
		Enabled: true,
		CircuitBreaker: safety.CircuitBreakerConfig{
			MaxFailures:  3,
			ResetTimeout: "2s",
		},
		RateLimit: safety.RateLimitConfig{
			MaxRequests: 5,
			Interval:    "10s",
		},
		Liquidity: safety.LiquidityConfig{
			MinOrderBookDepth:   5,
			MinTotalVolume:      10000,
			MaxSpreadPercent:    1.0,
			MinVolumeMultiplier: 0.2,
		},
		PositionLimits: safety.PositionLimitsConfig{
			MaxPositionSizeUSD:  500,
			MaxPortfolioPercent: 30,
			MaxDailyLossUSD:     50,
			MaxTotalPositions:   2,
		},
		Recovery: safety.RecoveryConfig{
			Strategy:   "exponential",
			MaxRetries: 2,
			BaseDelay:  100 * time.Millisecond,
			MaxDelay:   1 * time.Second,
		},
	}

	sm, err := safety.NewSafetyManager(client, config)
	if err != nil {
		log.Fatalf("  ❌ Failed to create safety manager: %v", err)
	}

	// Test CheckTradeAllowed
	log.Println("  Testing integrated trade checks...")
	err = sm.CheckTradeAllowed(context.Background(), "RVNUSD", 1000.0, 0.01, "BUY")
	if err != nil {
		log.Printf("  ⚠️  Trade check: %v", err)
	} else {
		log.Println("  ✅ Trade checks passed")
	}

	// Test ExecuteWithSafety
	log.Println("  Testing safe execution wrapper...")
	err = sm.ExecuteWithSafety(func() error {
		log.Println("  ✓ Protected operation executed")
		return nil
	})
	if err == nil {
		log.Println("  ✅ Safe execution completed")
	}

	// Test position tracking
	log.Println("  Testing position tracking...")
	sm.OpenPosition()
	sm.RecordTrade(25.0, true) // Record $25 profit
	status := sm.GetStatus()
	log.Printf("  📊 Safety status: %+v", status)
	log.Println("  ✅ Position tracking working")

	sm.ClosePosition()
	status = sm.GetStatus()
	log.Printf("  📊 After close: %+v", status)
}
