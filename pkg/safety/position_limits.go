package safety

import (
	"context"
	"fmt"
	"strconv"

	"github.com/adshao/go-binance/v2"
)

// PositionLimits enforces position sizing rules
type PositionLimits struct {
	client                *binance.Client
	maxPositionSizeUSD    float64 // Maximum position size in USD
	maxPortfolioPercent   float64 // Maximum % of portfolio in single position
	maxDailyLossUSD       float64 // Maximum daily loss limit
	maxTotalPositions     int     // Maximum number of open positions
	currentDailyLoss      float64 // Current day's losses
	openPositions         int     // Current open positions count
}

// PositionLimitsConfig holds configuration for position limits
type PositionLimitsConfig struct {
	MaxPositionSizeUSD  float64 `yaml:"max_position_size_usd"`
	MaxPortfolioPercent float64 `yaml:"max_portfolio_percent"`
	MaxDailyLossUSD     float64 `yaml:"max_daily_loss_usd"`
	MaxTotalPositions   int     `yaml:"max_total_positions"`
}

// NewPositionLimits creates a new position limits enforcer
func NewPositionLimits(client *binance.Client, config PositionLimitsConfig) *PositionLimits {
	return &PositionLimits{
		client:              client,
		maxPositionSizeUSD:  config.MaxPositionSizeUSD,
		maxPortfolioPercent: config.MaxPortfolioPercent,
		maxDailyLossUSD:     config.MaxDailyLossUSD,
		maxTotalPositions:   config.MaxTotalPositions,
		currentDailyLoss:    0,
		openPositions:       0,
	}
}

// CheckPositionSize verifies if a new position is within limits
func (pl *PositionLimits) CheckPositionSize(ctx context.Context, symbol string, quantity float64, price float64) error {
	positionValueUSD := quantity * price

	// Check absolute position size limit
	if positionValueUSD > pl.maxPositionSizeUSD {
		return fmt.Errorf("position size (%.2f USD) exceeds maximum (%.2f USD)",
			positionValueUSD, pl.maxPositionSizeUSD)
	}

	// Get account balance to check portfolio percentage
	account, err := pl.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}

	// Calculate total portfolio value in USD
	totalPortfolioUSD := 0.0
	for _, balance := range account.Balances {
		free, _ := strconv.ParseFloat(balance.Free, 64)
		locked, _ := strconv.ParseFloat(balance.Locked, 64)
		total := free + locked

		if total > 0 {
			// For USD-based assets
			if balance.Asset == "USD" || balance.Asset == "USDT" || balance.Asset == "BUSD" {
				totalPortfolioUSD += total
			}
			// For other assets, would need price conversion (simplified here)
		}
	}

	// Check portfolio percentage limit
	if totalPortfolioUSD > 0 {
		positionPercent := (positionValueUSD / totalPortfolioUSD) * 100
		if positionPercent > pl.maxPortfolioPercent {
			return fmt.Errorf("position would be %.2f%% of portfolio (max: %.2f%%)",
				positionPercent, pl.maxPortfolioPercent)
		}
	}

	// Check daily loss limit
	if pl.currentDailyLoss >= pl.maxDailyLossUSD {
		return fmt.Errorf("daily loss limit reached: %.2f USD (max: %.2f USD)",
			pl.currentDailyLoss, pl.maxDailyLossUSD)
	}

	// Check maximum number of positions
	if pl.openPositions >= pl.maxTotalPositions {
		return fmt.Errorf("maximum number of positions reached: %d (max: %d)",
			pl.openPositions, pl.maxTotalPositions)
	}

	return nil
}

// RecordLoss adds to the daily loss counter
func (pl *PositionLimits) RecordLoss(lossUSD float64) {
	pl.currentDailyLoss += lossUSD
}

// RecordProfit subtracts from the daily loss counter (can go negative = net profit)
func (pl *PositionLimits) RecordProfit(profitUSD float64) {
	pl.currentDailyLoss -= profitUSD
}

// IncrementPosition increments the open position counter
func (pl *PositionLimits) IncrementPosition() {
	pl.openPositions++
}

// DecrementPosition decrements the open position counter
func (pl *PositionLimits) DecrementPosition() {
	if pl.openPositions > 0 {
		pl.openPositions--
	}
}

// ResetDailyLoss resets the daily loss counter (call at start of new day)
func (pl *PositionLimits) ResetDailyLoss() {
	pl.currentDailyLoss = 0
}

// GetCurrentDailyLoss returns the current daily loss
func (pl *PositionLimits) GetCurrentDailyLoss() float64 {
	return pl.currentDailyLoss
}

// GetOpenPositions returns the current number of open positions
func (pl *PositionLimits) GetOpenPositions() int {
	return pl.openPositions
}

// IsDailyLimitReached returns true if daily loss limit is reached
func (pl *PositionLimits) IsDailyLimitReached() bool {
	return pl.currentDailyLoss >= pl.maxDailyLossUSD
}
