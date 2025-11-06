package strategy

import (
	"fmt"
	"math"
)

// RiskManager handles position sizing, stop-loss, and take-profit calculations
type RiskManager struct {
	config RiskConfig
}

// RiskConfig defines risk management parameters
type RiskConfig struct {
	// Position Sizing
	MaxPositionSizePercent float64 // Maximum % of portfolio per trade (default: 10%)
	RiskPerTradePercent    float64 // Maximum % to risk per trade (default: 2%)

	// Stop-Loss
	StopLossPercent    float64 // Fixed stop-loss % (default: 3%)
	UseATRStopLoss     bool    // Use ATR-based dynamic stop-loss
	ATRMultiplier      float64 // ATR multiplier for stop-loss (default: 2.0)

	// Take-Profit
	TakeProfitPercent  float64 // Fixed take-profit % (default: 6%)
	UseRiskRewardRatio bool    // Use risk/reward ratio instead
	RiskRewardRatio    float64 // Risk/reward ratio (default: 2.0 = 2:1)

	// Trailing Stop
	UseTrailingStop       bool    // Enable trailing stop
	TrailingStopPercent   float64 // Trailing stop activation % (default: 4%)
	TrailingStopDistance  float64 // Distance from peak % (default: 2%)

	// Portfolio Constraints
	MaxOpenPositions   int     // Maximum concurrent positions (default: 3)
	MaxPortfolioRisk   float64 // Maximum total portfolio risk % (default: 6%)
}

// DefaultRiskConfig returns conservative risk management defaults
func DefaultRiskConfig() RiskConfig {
	return RiskConfig{
		MaxPositionSizePercent: 10.0,
		RiskPerTradePercent:    2.0,

		StopLossPercent: 3.0,
		UseATRStopLoss:  false,
		ATRMultiplier:   2.0,

		TakeProfitPercent:  6.0,
		UseRiskRewardRatio: true,
		RiskRewardRatio:    2.0,

		UseTrailingStop:      true,
		TrailingStopPercent:  4.0,
		TrailingStopDistance: 2.0,

		MaxOpenPositions: 3,
		MaxPortfolioRisk: 6.0,
	}
}

// NewRiskManager creates a new risk manager
func NewRiskManager(config RiskConfig) *RiskManager {
	return &RiskManager{config: config}
}

// PositionSizeResult contains position sizing calculations
type PositionSizeResult struct {
	Quantity           float64 // Calculated quantity to buy
	EntryPrice         float64 // Entry price
	StopLossPrice      float64 // Stop-loss price
	TakeProfitPrice    float64 // Take-profit price
	RiskAmount         float64 // Dollar amount at risk
	PotentialProfit    float64 // Potential profit amount
	PositionValue      float64 // Total position value
	RiskRewardRatio    float64 // Actual risk/reward ratio
	MaxLossPercent     float64 // Maximum loss as % of portfolio
}

// CalculatePositionSize determines the appropriate position size based on risk parameters
func (rm *RiskManager) CalculatePositionSize(
	portfolioValue float64,
	entryPrice float64,
	volatility float64, // ATR or similar volatility measure
) (PositionSizeResult, error) {
	if portfolioValue <= 0 {
		return PositionSizeResult{}, fmt.Errorf("portfolio value must be positive")
	}
	if entryPrice <= 0 {
		return PositionSizeResult{}, fmt.Errorf("entry price must be positive")
	}

	result := PositionSizeResult{
		EntryPrice: entryPrice,
	}

	// Calculate stop-loss price
	if rm.config.UseATRStopLoss && volatility > 0 {
		// Dynamic ATR-based stop-loss
		stopDistance := volatility * rm.config.ATRMultiplier
		result.StopLossPrice = entryPrice - stopDistance
	} else {
		// Fixed percentage stop-loss
		result.StopLossPrice = entryPrice * (1 - rm.config.StopLossPercent/100.0)
	}

	// Calculate take-profit price
	if rm.config.UseRiskRewardRatio {
		// Based on risk/reward ratio
		riskPerUnit := entryPrice - result.StopLossPrice
		rewardPerUnit := riskPerUnit * rm.config.RiskRewardRatio
		result.TakeProfitPrice = entryPrice + rewardPerUnit
	} else {
		// Fixed percentage take-profit
		result.TakeProfitPrice = entryPrice * (1 + rm.config.TakeProfitPercent/100.0)
	}

	// Calculate maximum position value based on portfolio percentage
	maxPositionValue := portfolioValue * (rm.config.MaxPositionSizePercent / 100.0)

	// Calculate position size based on risk per trade
	riskPerUnit := entryPrice - result.StopLossPrice
	if riskPerUnit <= 0 {
		return PositionSizeResult{}, fmt.Errorf("invalid stop-loss: must be below entry price")
	}

	// Maximum risk amount
	maxRiskAmount := portfolioValue * (rm.config.RiskPerTradePercent / 100.0)

	// Position size = risk amount / risk per unit
	quantityByRisk := maxRiskAmount / riskPerUnit

	// Position value constraint
	maxQuantityByValue := maxPositionValue / entryPrice

	// Take the smaller of the two
	result.Quantity = math.Min(quantityByRisk, maxQuantityByValue)

	// Calculate actual values
	result.PositionValue = result.Quantity * entryPrice
	result.RiskAmount = result.Quantity * riskPerUnit
	result.PotentialProfit = result.Quantity * (result.TakeProfitPrice - entryPrice)
	result.MaxLossPercent = (result.RiskAmount / portfolioValue) * 100.0
	result.RiskRewardRatio = result.PotentialProfit / result.RiskAmount

	return result, nil
}

// TrailingStopTracker tracks the trailing stop for an open position
type TrailingStopTracker struct {
	EntryPrice       float64
	HighestPrice     float64 // Highest price since entry
	StopLossPrice    float64 // Current stop-loss price
	TrailingActive   bool    // Whether trailing stop is activated
	ActivationPrice  float64 // Price at which trailing stop activates
	TrailingDistance float64 // Distance from peak (%)
}

// NewTrailingStopTracker creates a new trailing stop tracker
func NewTrailingStopTracker(entryPrice float64, initialStopLoss float64, activationPercent float64, trailingDistance float64) *TrailingStopTracker {
	return &TrailingStopTracker{
		EntryPrice:       entryPrice,
		HighestPrice:     entryPrice,
		StopLossPrice:    initialStopLoss,
		TrailingActive:   false,
		ActivationPrice:  entryPrice * (1 + activationPercent/100.0),
		TrailingDistance: trailingDistance,
	}
}

// Update updates the trailing stop based on current price
// Returns true if stop-loss was triggered
func (tst *TrailingStopTracker) Update(currentPrice float64) bool {
	// Update highest price
	if currentPrice > tst.HighestPrice {
		tst.HighestPrice = currentPrice
	}

	// Activate trailing stop if price reaches activation level
	if !tst.TrailingActive && currentPrice >= tst.ActivationPrice {
		tst.TrailingActive = true
	}

	// Update trailing stop-loss if active
	if tst.TrailingActive {
		newStopLoss := tst.HighestPrice * (1 - tst.TrailingDistance/100.0)
		if newStopLoss > tst.StopLossPrice {
			tst.StopLossPrice = newStopLoss
		}
	}

	// Check if stop-loss triggered
	return currentPrice <= tst.StopLossPrice
}

// GetStopLossPrice returns the current stop-loss price
func (tst *TrailingStopTracker) GetStopLossPrice() float64 {
	return tst.StopLossPrice
}

// GetUnrealizedProfit calculates current unrealized profit/loss
func (tst *TrailingStopTracker) GetUnrealizedProfit(currentPrice float64, quantity float64) float64 {
	return (currentPrice - tst.EntryPrice) * quantity
}

// GetUnrealizedProfitPercent calculates unrealized profit as percentage
func (tst *TrailingStopTracker) GetUnrealizedProfitPercent(currentPrice float64) float64 {
	return ((currentPrice - tst.EntryPrice) / tst.EntryPrice) * 100.0
}

// ShouldExit checks if exit conditions are met
func (rm *RiskManager) ShouldExit(
	entryPrice float64,
	currentPrice float64,
	stopLossPrice float64,
	takeProfitPrice float64,
) (shouldExit bool, reason string) {
	// Check stop-loss
	if currentPrice <= stopLossPrice {
		lossPercent := ((stopLossPrice - entryPrice) / entryPrice) * 100.0
		return true, fmt.Sprintf("Stop-loss triggered at %.8f (%.2f%% loss)", stopLossPrice, lossPercent)
	}

	// Check take-profit
	if currentPrice >= takeProfitPrice {
		profitPercent := ((takeProfitPrice - entryPrice) / entryPrice) * 100.0
		return true, fmt.Sprintf("Take-profit reached at %.8f (%.2f%% profit)", takeProfitPrice, profitPercent)
	}

	return false, ""
}

// ValidatePositionRisk checks if a new position would exceed risk limits
func (rm *RiskManager) ValidatePositionRisk(
	portfolioValue float64,
	newPositionRisk float64,
	existingPositions int,
	existingTotalRisk float64,
) error {
	// Check maximum open positions
	if existingPositions >= rm.config.MaxOpenPositions {
		return fmt.Errorf("maximum open positions (%d) reached", rm.config.MaxOpenPositions)
	}

	// Check portfolio risk limit
	totalRiskPercent := ((existingTotalRisk + newPositionRisk) / portfolioValue) * 100.0
	if totalRiskPercent > rm.config.MaxPortfolioRisk {
		return fmt.Errorf("total portfolio risk (%.2f%%) would exceed limit (%.2f%%)",
			totalRiskPercent, rm.config.MaxPortfolioRisk)
	}

	// Check individual trade risk
	newTradeRiskPercent := (newPositionRisk / portfolioValue) * 100.0
	if newTradeRiskPercent > rm.config.RiskPerTradePercent {
		return fmt.Errorf("trade risk (%.2f%%) exceeds limit (%.2f%%)",
			newTradeRiskPercent, rm.config.RiskPerTradePercent)
	}

	return nil
}

// PositionSummary provides a summary of position risk metrics
type PositionSummary struct {
	EntryPrice         float64
	CurrentPrice       float64
	Quantity           float64
	StopLossPrice      float64
	TakeProfitPrice    float64
	UnrealizedPL       float64
	UnrealizedPLPercent float64
	RiskAmount         float64
	PotentialReward    float64
	CurrentRiskReward  float64
}

// GetPositionSummary calculates current position metrics
func (rm *RiskManager) GetPositionSummary(
	entryPrice float64,
	currentPrice float64,
	quantity float64,
	stopLossPrice float64,
	takeProfitPrice float64,
) PositionSummary {
	unrealizedPL := (currentPrice - entryPrice) * quantity
	unrealizedPLPercent := ((currentPrice - entryPrice) / entryPrice) * 100.0

	riskAmount := (entryPrice - stopLossPrice) * quantity
	potentialReward := (takeProfitPrice - currentPrice) * quantity

	currentRiskReward := 0.0
	if riskAmount > 0 {
		currentRiskReward = potentialReward / riskAmount
	}

	return PositionSummary{
		EntryPrice:          entryPrice,
		CurrentPrice:        currentPrice,
		Quantity:            quantity,
		StopLossPrice:       stopLossPrice,
		TakeProfitPrice:     takeProfitPrice,
		UnrealizedPL:        unrealizedPL,
		UnrealizedPLPercent: unrealizedPLPercent,
		RiskAmount:          riskAmount,
		PotentialReward:     potentialReward,
		CurrentRiskReward:   currentRiskReward,
	}
}
