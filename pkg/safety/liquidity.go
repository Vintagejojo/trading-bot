package safety

import (
	"context"
	"fmt"
	"strconv"

	"github.com/adshao/go-binance/v2"
)

// LiquidityChecker verifies market depth before executing trades
type LiquidityChecker struct {
	client               *binance.Client
	minOrderBookDepth    int     // Minimum number of orders on each side
	minTotalVolume       float64 // Minimum total volume in order book
	maxSpreadPercent     float64 // Maximum allowed bid-ask spread %
	minVolumeMultiplier  float64 // Order size must be < this * available volume
}

// LiquidityConfig holds configuration for liquidity checks
type LiquidityConfig struct {
	MinOrderBookDepth   int     `yaml:"min_order_book_depth"`
	MinTotalVolume      float64 `yaml:"min_total_volume"`
	MaxSpreadPercent    float64 `yaml:"max_spread_percent"`
	MinVolumeMultiplier float64 `yaml:"min_volume_multiplier"`
}

// NewLiquidityChecker creates a new liquidity checker
func NewLiquidityChecker(client *binance.Client, config LiquidityConfig) *LiquidityChecker {
	return &LiquidityChecker{
		client:               client,
		minOrderBookDepth:    config.MinOrderBookDepth,
		minTotalVolume:       config.MinTotalVolume,
		maxSpreadPercent:     config.MaxSpreadPercent,
		minVolumeMultiplier:  config.MinVolumeMultiplier,
	}
}

// CheckLiquidity verifies if there's sufficient liquidity for a trade
func (lc *LiquidityChecker) CheckLiquidity(ctx context.Context, symbol string, orderSize float64, side string) error {
	// Get order book depth
	depth, err := lc.client.NewDepthService().Symbol(symbol).Limit(100).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to get order book: %w", err)
	}

	// Check minimum order book depth
	if len(depth.Bids) < lc.minOrderBookDepth || len(depth.Asks) < lc.minOrderBookDepth {
		return fmt.Errorf("insufficient order book depth: bids=%d, asks=%d, required=%d",
			len(depth.Bids), len(depth.Asks), lc.minOrderBookDepth)
	}

	// Calculate bid-ask spread
	if len(depth.Bids) == 0 || len(depth.Asks) == 0 {
		return fmt.Errorf("empty order book")
	}

	bestBid, err := strconv.ParseFloat(depth.Bids[0].Price, 64)
	if err != nil {
		return fmt.Errorf("invalid bid price: %w", err)
	}

	bestAsk, err := strconv.ParseFloat(depth.Asks[0].Price, 64)
	if err != nil {
		return fmt.Errorf("invalid ask price: %w", err)
	}

	spreadPercent := ((bestAsk - bestBid) / bestBid) * 100
	if spreadPercent > lc.maxSpreadPercent {
		return fmt.Errorf("spread too wide: %.2f%% (max: %.2f%%)", spreadPercent, lc.maxSpreadPercent)
	}

	// Check total volume availability
	var totalVolume float64
	var orders []binance.Bid

	if side == "BUY" {
		orders = depth.Asks
	} else {
		orders = depth.Bids
	}

	for _, order := range orders {
		qty, err := strconv.ParseFloat(order.Quantity, 64)
		if err != nil {
			continue
		}
		totalVolume += qty
	}

	if totalVolume < lc.minTotalVolume {
		return fmt.Errorf("insufficient total volume: %.2f (min: %.2f)", totalVolume, lc.minTotalVolume)
	}

	// Check if order size is reasonable compared to available volume
	if orderSize > totalVolume*lc.minVolumeMultiplier {
		return fmt.Errorf("order size too large: %.2f > %.2f%% of available volume (%.2f)",
			orderSize, lc.minVolumeMultiplier*100, totalVolume)
	}

	return nil
}

// GetMarketDepth returns current market depth information
func (lc *LiquidityChecker) GetMarketDepth(ctx context.Context, symbol string) (bestBid, bestAsk, spreadPercent float64, err error) {
	depth, err := lc.client.NewDepthService().Symbol(symbol).Limit(10).Do(ctx)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get order book: %w", err)
	}

	if len(depth.Bids) == 0 || len(depth.Asks) == 0 {
		return 0, 0, 0, fmt.Errorf("empty order book")
	}

	bestBid, _ = strconv.ParseFloat(depth.Bids[0].Price, 64)
	bestAsk, _ = strconv.ParseFloat(depth.Asks[0].Price, 64)
	spreadPercent = ((bestAsk - bestBid) / bestBid) * 100

	return bestBid, bestAsk, spreadPercent, nil
}
