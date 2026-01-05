package portfolio

import (
	"rsi-bot/pkg/database"
)

// Stats represents portfolio statistics
type Stats struct {
	Symbol          string  `json:"symbol"`
	TotalHoldings   float64 `json:"total_holdings"`   // Total BTC held
	TotalCost       float64 `json:"total_cost"`       // Total USD invested
	AverageCost     float64 `json:"average_cost"`     // Average buy price per BTC
	CurrentPrice    float64 `json:"current_price"`    // Current market price
	CurrentValue    float64 `json:"current_value"`    // Current portfolio value (holdings * price)
	UnrealizedGain  float64 `json:"unrealized_gain"`  // Unrealized profit/loss in USD
	UnrealizedROI   float64 `json:"unrealized_roi"`   // Unrealized ROI percentage
	TotalBuys       int     `json:"total_buys"`
	TotalSells      int     `json:"total_sells"`
	RealizedGains   float64 `json:"realized_gains"`   // Profit from closed positions
}

// Calculator calculates portfolio statistics
type Calculator struct {
	db *database.DB
}

// NewCalculator creates a new portfolio calculator
func NewCalculator(db *database.DB) *Calculator {
	return &Calculator{db: db}
}

// CalculateStats calculates current portfolio statistics
func (c *Calculator) CalculateStats(symbol string, currentPrice float64) (*Stats, error) {
	stats := &Stats{
		Symbol:       symbol,
		CurrentPrice: currentPrice,
	}

	// Get all trades for this symbol
	trades, err := c.db.GetRecentTrades(10000) // Get all trades
	if err != nil {
		return nil, err
	}

	var totalBTCBought float64
	var totalBTCSold float64
	var totalUSDSpent float64
	var totalUSDReceived float64

	for _, trade := range trades {
		if trade.Symbol != symbol {
			continue
		}

		if trade.Side == "BUY" {
			totalBTCBought += trade.Quantity
			totalUSDSpent += trade.Total
			stats.TotalBuys++
		} else if trade.Side == "SELL" {
			totalBTCSold += trade.Quantity
			totalUSDReceived += trade.Total
			stats.TotalSells++
		}
	}

	// Calculate holdings
	stats.TotalHoldings = totalBTCBought - totalBTCSold

	// Calculate cost basis (only count buys that haven't been sold)
	if stats.TotalHoldings > 0 && totalBTCBought > 0 {
		// For simplicity, use weighted average of all buys
		// In reality, you'd want to use FIFO/LIFO for tax purposes
		stats.TotalCost = totalUSDSpent * (stats.TotalHoldings / totalBTCBought)
		stats.AverageCost = stats.TotalCost / stats.TotalHoldings
	}

	// Calculate current value and gains
	stats.CurrentValue = stats.TotalHoldings * currentPrice
	stats.UnrealizedGain = stats.CurrentValue - stats.TotalCost
	if stats.TotalCost > 0 {
		stats.UnrealizedROI = (stats.UnrealizedGain / stats.TotalCost) * 100
	}

	// Calculate realized gains (from sells)
	if totalBTCBought > 0 {
		stats.RealizedGains = totalUSDReceived - (totalUSDSpent * (totalBTCSold / totalBTCBought))
	}

	return stats, nil
}

// GetWeeklyStats calculates stats for the past week
func (c *Calculator) GetWeeklyStats(symbol string, currentPrice float64) (*WeeklyStats, error) {
	// Implementation for weekly summary
	// Get trades from last 7 days, calculate accumulated BTC, invested amount, etc.
	return nil, nil // TODO: Implement
}

// WeeklyStats represents weekly portfolio statistics
type WeeklyStats struct {
	NumPurchases   int
	TotalInvested  float64
	BTCAccumulated float64
	BestBuyPrice   float64
	WorstBuyPrice  float64
}
