package strategy

import (
	"fmt"
	"math"
)

// MarketConditionAnalyzer evaluates market conditions for trade viability
type MarketConditionAnalyzer struct {
	config MarketConditionConfig
}

// MarketConditionConfig defines thresholds for market condition checks
type MarketConditionConfig struct {
	// Volatility thresholds (using Bollinger Band width)
	MinVolatilityPercent float64 // Minimum volatility for trading (default: 1%)
	MaxVolatilityPercent float64 // Maximum volatility to avoid extreme moves (default: 10%)

	// Volume analysis
	UseVolumeFilter      bool    // Enable volume filtering
	MinVolumeMultiplier  float64 // Minimum volume vs average (default: 0.5 = 50% of avg)
	VolumeAveragePeriod  int     // Period for volume average (default: 20)

	// Spread analysis (bid-ask spread)
	MaxSpreadPercent     float64 // Maximum spread % for liquidity (default: 0.5%)

	// ATR (Average True Range) for dynamic volatility
	UseATR               bool    // Use ATR for volatility measurement
	ATRPeriod            int     // ATR calculation period (default: 14)
	MinATRPercent        float64 // Minimum ATR as % of price
	MaxATRPercent        float64 // Maximum ATR as % of price
}

// DefaultMarketConditionConfig returns sensible defaults
func DefaultMarketConditionConfig() MarketConditionConfig {
	return MarketConditionConfig{
		MinVolatilityPercent: 1.0,
		MaxVolatilityPercent: 10.0,

		UseVolumeFilter:     true,
		MinVolumeMultiplier: 0.5,
		VolumeAveragePeriod: 20,

		MaxSpreadPercent: 0.5,

		UseATR:        false,
		ATRPeriod:     14,
		MinATRPercent: 0.5,
		MaxATRPercent: 5.0,
	}
}

// NewMarketConditionAnalyzer creates a new market condition analyzer
func NewMarketConditionAnalyzer(config MarketConditionConfig) *MarketConditionAnalyzer {
	return &MarketConditionAnalyzer{config: config}
}

// MarketCondition represents the current market state
type MarketCondition struct {
	IsTradeableMarket bool
	Reasons           []string

	// Volatility metrics
	Volatility          float64
	VolatilityStatus    string // "LOW", "NORMAL", "HIGH", "EXTREME"

	// Liquidity metrics
	Volume              float64
	VolumeAverage       float64
	VolumeRatio         float64
	LiquidityStatus     string // "POOR", "ADEQUATE", "GOOD"

	// Spread metrics
	SpreadPercent       float64
	SpreadStatus        string // "TIGHT", "NORMAL", "WIDE"
}

// AnalyzeMarketConditions evaluates if market conditions are suitable for trading
func (mca *MarketConditionAnalyzer) AnalyzeMarketConditions(
	currentVolatility float64, // Bollinger Band width or ATR
	currentVolume float64,
	volumeHistory []float64,
	bidPrice float64,
	askPrice float64,
) MarketCondition {
	mc := MarketCondition{
		IsTradeableMarket: true,
		Reasons:           make([]string, 0),
		Volatility:        currentVolatility,
		Volume:            currentVolume,
	}

	// === VOLATILITY CHECK ===
	mc.analyzeVolatility(mca.config)

	if currentVolatility < mca.config.MinVolatilityPercent {
		mc.IsTradeableMarket = false
		mc.Reasons = append(mc.Reasons, fmt.Sprintf("Volatility too low (%.2f%% < %.2f%%)", currentVolatility, mca.config.MinVolatilityPercent))
	} else if currentVolatility > mca.config.MaxVolatilityPercent {
		mc.IsTradeableMarket = false
		mc.Reasons = append(mc.Reasons, fmt.Sprintf("Volatility too high (%.2f%% > %.2f%%)", currentVolatility, mca.config.MaxVolatilityPercent))
	} else {
		mc.Reasons = append(mc.Reasons, fmt.Sprintf("Volatility OK (%.2f%%)", currentVolatility))
	}

	// === VOLUME/LIQUIDITY CHECK ===
	if mca.config.UseVolumeFilter && len(volumeHistory) >= mca.config.VolumeAveragePeriod {
		mc.analyzeVolume(volumeHistory, mca.config)

		if mc.VolumeRatio < mca.config.MinVolumeMultiplier {
			mc.IsTradeableMarket = false
			mc.Reasons = append(mc.Reasons, fmt.Sprintf("Volume too low (%.2fx avg)", mc.VolumeRatio))
		} else {
			mc.Reasons = append(mc.Reasons, fmt.Sprintf("Volume adequate (%.2fx avg)", mc.VolumeRatio))
		}
	}

	// === SPREAD CHECK (Liquidity) ===
	if bidPrice > 0 && askPrice > 0 {
		mc.analyzeSpread(bidPrice, askPrice, mca.config)

		if mc.SpreadPercent > mca.config.MaxSpreadPercent {
			mc.IsTradeableMarket = false
			mc.Reasons = append(mc.Reasons, fmt.Sprintf("Spread too wide (%.3f%% > %.3f%%)", mc.SpreadPercent, mca.config.MaxSpreadPercent))
		} else {
			mc.Reasons = append(mc.Reasons, fmt.Sprintf("Spread acceptable (%.3f%%)", mc.SpreadPercent))
		}
	}

	return mc
}

// analyzeVolatility categorizes volatility level
func (mc *MarketCondition) analyzeVolatility(config MarketConditionConfig) {
	if mc.Volatility < config.MinVolatilityPercent {
		mc.VolatilityStatus = "LOW"
	} else if mc.Volatility > config.MaxVolatilityPercent {
		mc.VolatilityStatus = "EXTREME"
	} else if mc.Volatility > config.MaxVolatilityPercent*0.7 {
		mc.VolatilityStatus = "HIGH"
	} else {
		mc.VolatilityStatus = "NORMAL"
	}
}

// analyzeVolume calculates volume metrics
func (mc *MarketCondition) analyzeVolume(volumeHistory []float64, config MarketConditionConfig) {
	// Calculate volume average
	sum := 0.0
	period := config.VolumeAveragePeriod
	start := len(volumeHistory) - period
	if start < 0 {
		start = 0
	}

	for i := start; i < len(volumeHistory); i++ {
		sum += volumeHistory[i]
	}
	mc.VolumeAverage = sum / float64(len(volumeHistory[start:]))

	// Calculate volume ratio
	if mc.VolumeAverage > 0 {
		mc.VolumeRatio = mc.Volume / mc.VolumeAverage
	}

	// Categorize liquidity
	if mc.VolumeRatio < config.MinVolumeMultiplier {
		mc.LiquidityStatus = "POOR"
	} else if mc.VolumeRatio < 1.0 {
		mc.LiquidityStatus = "ADEQUATE"
	} else {
		mc.LiquidityStatus = "GOOD"
	}
}

// analyzeSpread calculates and categorizes bid-ask spread
func (mc *MarketCondition) analyzeSpread(bidPrice float64, askPrice float64, config MarketConditionConfig) {
	midPrice := (bidPrice + askPrice) / 2.0
	spread := askPrice - bidPrice
	mc.SpreadPercent = (spread / midPrice) * 100.0

	if mc.SpreadPercent < config.MaxSpreadPercent*0.5 {
		mc.SpreadStatus = "TIGHT"
	} else if mc.SpreadPercent < config.MaxSpreadPercent {
		mc.SpreadStatus = "NORMAL"
	} else {
		mc.SpreadStatus = "WIDE"
	}
}

// String returns a human-readable summary
func (mc MarketCondition) String() string {
	status := "TRADEABLE"
	if !mc.IsTradeableMarket {
		status = "NOT TRADEABLE"
	}

	summary := fmt.Sprintf("[%s] Volatility: %s (%.2f%%) | Liquidity: %s (%.2fx) | Spread: %s (%.3f%%)",
		status,
		mc.VolatilityStatus, mc.Volatility,
		mc.LiquidityStatus, mc.VolumeRatio,
		mc.SpreadStatus, mc.SpreadPercent,
	)

	return summary
}

// ATRCalculator calculates Average True Range for volatility measurement
type ATRCalculator struct {
	period     int
	trueRanges []float64
	atr        float64
	isReady    bool
}

// NewATRCalculator creates a new ATR calculator
func NewATRCalculator(period int) *ATRCalculator {
	return &ATRCalculator{
		period:     period,
		trueRanges: make([]float64, 0, period+10),
	}
}

// Update calculates true range and updates ATR
func (atr *ATRCalculator) Update(high, low, previousClose float64) {
	// True Range = max(high-low, abs(high-prevClose), abs(low-prevClose))
	tr1 := high - low
	tr2 := math.Abs(high - previousClose)
	tr3 := math.Abs(low - previousClose)

	trueRange := math.Max(tr1, math.Max(tr2, tr3))
	atr.trueRanges = append(atr.trueRanges, trueRange)

	// Keep buffer manageable
	if len(atr.trueRanges) > atr.period+10 {
		atr.trueRanges = atr.trueRanges[1:]
	}

	// Calculate ATR when we have enough data
	if len(atr.trueRanges) >= atr.period {
		sum := 0.0
		start := len(atr.trueRanges) - atr.period
		for i := start; i < len(atr.trueRanges); i++ {
			sum += atr.trueRanges[i]
		}
		atr.atr = sum / float64(atr.period)
		atr.isReady = true
	}
}

// GetATR returns the current ATR value
func (atr *ATRCalculator) GetATR() (float64, bool) {
	return atr.atr, atr.isReady
}

// GetATRPercent returns ATR as a percentage of current price
func (atr *ATRCalculator) GetATRPercent(currentPrice float64) (float64, bool) {
	if !atr.isReady || currentPrice <= 0 {
		return 0, false
	}
	return (atr.atr / currentPrice) * 100.0, true
}

// VolumeTracker tracks volume history for analysis
type VolumeTracker struct {
	volumes    []float64
	maxHistory int
}

// NewVolumeTracker creates a new volume tracker
func NewVolumeTracker(maxHistory int) *VolumeTracker {
	return &VolumeTracker{
		volumes:    make([]float64, 0, maxHistory),
		maxHistory: maxHistory,
	}
}

// Add adds a new volume data point
func (vt *VolumeTracker) Add(volume float64) {
	vt.volumes = append(vt.volumes, volume)

	if len(vt.volumes) > vt.maxHistory {
		vt.volumes = vt.volumes[1:]
	}
}

// GetHistory returns volume history
func (vt *VolumeTracker) GetHistory() []float64 {
	return vt.volumes
}

// GetAverage calculates average volume over specified period
func (vt *VolumeTracker) GetAverage(period int) (float64, bool) {
	if len(vt.volumes) < period {
		return 0, false
	}

	sum := 0.0
	start := len(vt.volumes) - period
	for i := start; i < len(vt.volumes); i++ {
		sum += vt.volumes[i]
	}

	return sum / float64(period), true
}

// GetCurrentRatio returns current volume as ratio of average
func (vt *VolumeTracker) GetCurrentRatio(currentVolume float64, period int) (float64, bool) {
	avg, ok := vt.GetAverage(period)
	if !ok || avg == 0 {
		return 0, false
	}

	return currentVolume / avg, true
}

// IsReady returns true if enough data is available
func (vt *VolumeTracker) IsReady(minPeriod int) bool {
	return len(vt.volumes) >= minPeriod
}
