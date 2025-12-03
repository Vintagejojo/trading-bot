package strategy

import (
	"fmt"
	"log"
	"strings"
	"time"

	"rsi-bot/pkg/indicators"
	"rsi-bot/pkg/models"
)

// MultiTimeframeStrategy implements a strategy using multiple timeframes
// - Daily (1d): Trend bias filtering
// - 1-Hour (1h): Signal generation
// - 5-Minute (5m): Entry precision
type MultiTimeframeStrategy struct {
	name             string
	mtfManager       *MultiTimeframeManager
	lastSignalReason string

	// Strategy thresholds
	config MultiTimeframeStrategyConfig
}

// MultiTimeframeStrategyConfig defines the strategy parameters
type MultiTimeframeStrategyConfig struct {
	// RSI thresholds
	RSIOversold    float64 // Default: 30
	RSIOverbought  float64 // Default: 70
	RSINeutralLow  float64 // Default: 40
	RSINeutralHigh float64 // Default: 60

	// MACD thresholds
	MACDMinHistogram float64 // Minimum histogram value for signal strength

	// Bollinger Bands thresholds
	BBandsMinWidth float64 // Minimum band width % for sufficient volatility
	BBandsMaxWidth float64 // Maximum band width % to avoid extreme volatility

	// Multi-timeframe confirmation requirements
	RequireDailyTrendConfirmation bool // Require daily trend alignment
	RequireHourlySignal           bool // Require hourly signal
	Require5MinuteEntry           bool // Require 5-minute entry precision
}

// DefaultMultiTimeframeStrategyConfig returns sensible defaults
func DefaultMultiTimeframeStrategyConfig() MultiTimeframeStrategyConfig {
	return MultiTimeframeStrategyConfig{
		RSIOversold:    30.0,
		RSIOverbought:  70.0,
		RSINeutralLow:  40.0,
		RSINeutralHigh: 60.0,

		MACDMinHistogram: 0.0001, // Small positive value for confirmation

		BBandsMinWidth: 1.0,  // 1% minimum volatility
		BBandsMaxWidth: 10.0, // 10% maximum volatility

		RequireDailyTrendConfirmation: true,
		RequireHourlySignal:           true,
		Require5MinuteEntry:           true,
	}
}

// NewMultiTimeframeStrategy creates a new multi-timeframe strategy
func NewMultiTimeframeStrategy(config MultiTimeframeStrategyConfig) (*MultiTimeframeStrategy, error) {
	// Create multi-timeframe manager with daily, hourly, and 5-minute timeframes
	mtfConfig := DefaultMultiTimeframeConfig()
	mtfConfig.Timeframes = []Timeframe{Timeframe5m, Timeframe1h, Timeframe1d}

	mtfManager, err := NewMultiTimeframeManager(mtfConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create multi-timeframe manager: %w", err)
	}

	return &MultiTimeframeStrategy{
		name:       "MultiTimeframe",
		mtfManager: mtfManager,
		config:     config,
	}, nil
}

// Name returns the strategy identifier
func (mts *MultiTimeframeStrategy) Name() string {
	return mts.name
}

// GetIndicator returns the primary indicator (not used in multi-timeframe context)
func (mts *MultiTimeframeStrategy) GetIndicator() indicators.Indicator {
	// Return the 1-hour RSI as the "primary" indicator for interface compatibility
	if tfIndicators, ok := mts.mtfManager.Indicators[Timeframe1h]; ok {
		return tfIndicators.RSI
	}
	return nil
}

// Update processes new price data across all timeframes
func (mts *MultiTimeframeStrategy) Update(price float64, volume float64, timestamp time.Time) error {
	err := mts.mtfManager.Update(price, volume, timestamp)
	if err != nil {
		return err
	}
	// Debug: Check if data is being accumulated
	for tf, tfData := range mts.mtfManager.TimeframeData {
		log.Printf("[MTF Update] %s: %d candles", tf, len(tfData.Candles))
	}
	return nil
}

// GenerateSignal analyzes all timeframes and generates a trading signal
func (mts *MultiTimeframeStrategy) GenerateSignal(ctx SignalContext) Signal {
	// Get snapshots for all timeframes
	snapshots := mts.mtfManager.GetAllSnapshots()

	daily, hasDaily := snapshots[Timeframe1d]
	hourly, hasHourly := snapshots[Timeframe1h]
	fiveMin, hasFiveMin := snapshots[Timeframe5m]

	// Check if we have enough data
	if !hasHourly {
		mts.lastSignalReason = "Insufficient data on 1-hour timeframe"
		return SignalNone
	}

	// Build signal reasons
	var reasons []string

	// === PHASE 1: Daily Trend Bias (Optional Filter) ===
	dailyTrend := mts.analyzeDailyTrend(daily)
	if mts.config.RequireDailyTrendConfirmation && hasDaily {
		if dailyTrend == TrendNone {
			mts.lastSignalReason = "Daily trend is neutral - no clear bias"
			return SignalNone
		}
		reasons = append(reasons, fmt.Sprintf("Daily Trend: %s", dailyTrend))
	}

	// === PHASE 2: 1-Hour Signal Generation ===
	hourlySignal, hourlyReason := mts.analyzeHourlySignal(hourly, ctx.Position)
	reasons = append(reasons, hourlyReason)

	if hourlySignal == SignalNone {
		mts.lastSignalReason = strings.Join(reasons, " | ")
		return SignalNone
	}

	// === PHASE 3: Daily-Hourly Alignment Check ===
	if mts.config.RequireDailyTrendConfirmation && hasDaily {
		if !mts.checkTrendAlignment(dailyTrend, hourlySignal) {
			mts.lastSignalReason = fmt.Sprintf("Trend misalignment: Daily=%s, Hourly Signal=%s", dailyTrend, hourlySignal)
			return SignalNone
		}
	}

	// === PHASE 4: 5-Minute Entry Precision ===
	if mts.config.Require5MinuteEntry && hasFiveMin {
		if !mts.checkEntryPrecision(fiveMin, hourlySignal) {
			mts.lastSignalReason = strings.Join(append(reasons, "5-min entry conditions not met"), " | ")
			return SignalNone
		}
		reasons = append(reasons, "5-min entry confirmed")
	}

	// === PHASE 5: Volatility & Liquidity Check ===
	if !mts.checkVolatility(hourly) {
		mts.lastSignalReason = strings.Join(append(reasons, "Volatility outside acceptable range"), " | ")
		return SignalNone
	}
	reasons = append(reasons, fmt.Sprintf("Volatility OK (%.2f%%)", hourly.BBandsWidth))

	// All checks passed
	mts.lastSignalReason = strings.Join(reasons, " | ")
	return hourlySignal
}

// TrendDirection represents the market trend
type TrendDirection int

const (
	TrendNone TrendDirection = iota
	TrendBullish
	TrendBearish
)

func (td TrendDirection) String() string {
	switch td {
	case TrendBullish:
		return "BULLISH"
	case TrendBearish:
		return "BEARISH"
	default:
		return "NEUTRAL"
	}
}

// analyzeDailyTrend determines the daily trend bias
func (mts *MultiTimeframeStrategy) analyzeDailyTrend(daily IndicatorSnapshot) TrendDirection {
	if !daily.RSIReady || !daily.MACDReady {
		return TrendNone
	}

	bullishSignals := 0
	bearishSignals := 0

	// RSI trend analysis
	if daily.RSI < mts.config.RSINeutralLow {
		bearishSignals++ // Oversold on daily = bearish trend
	} else if daily.RSI > mts.config.RSINeutralHigh {
		bullishSignals++ // Overbought on daily = bullish trend
	}

	// MACD trend analysis
	if daily.MACDHistogram > mts.config.MACDMinHistogram {
		bullishSignals++ // Positive MACD histogram = bullish momentum
	} else if daily.MACDHistogram < -mts.config.MACDMinHistogram {
		bearishSignals++ // Negative MACD histogram = bearish momentum
	}

	// Price vs Bollinger Bands middle (SMA)
	if daily.BBandsReady {
		if daily.Price > daily.BBandsMiddle {
			bullishSignals++
		} else if daily.Price < daily.BBandsMiddle {
			bearishSignals++
		}
	}

	// Determine trend
	if bullishSignals > bearishSignals {
		return TrendBullish
	} else if bearishSignals > bullishSignals {
		return TrendBearish
	}

	return TrendNone
}

// analyzeHourlySignal generates buy/sell signals from 1-hour timeframe
func (mts *MultiTimeframeStrategy) analyzeHourlySignal(hourly IndicatorSnapshot, position *models.Position) (Signal, string) {
	if !hourly.RSIReady || !hourly.MACDReady || !hourly.BBandsReady {
		return SignalNone, "1h indicators not ready"
	}

	// === BUY SIGNAL CONDITIONS ===
	if !position.InPosition {
		buyConditions := 0
		var buyReasons []string

		// RSI oversold
		if hourly.RSI <= mts.config.RSIOversold {
			buyConditions++
			buyReasons = append(buyReasons, fmt.Sprintf("RSI oversold (%.2f)", hourly.RSI))
		}

		// MACD bullish crossover (histogram turning positive)
		if hourly.MACDHistogram > 0 && hourly.MACD > hourly.MACDSignal {
			buyConditions++
			buyReasons = append(buyReasons, "MACD bullish crossover")
		}

		// Price near or below lower Bollinger Band
		if hourly.Price <= hourly.BBandsLower*1.01 { // Within 1% of lower band
			buyConditions++
			buyReasons = append(buyReasons, "Price at lower BB")
		}

		// Need at least 2 out of 3 conditions
		if buyConditions >= 2 {
			return SignalBuy, "1h BUY: " + strings.Join(buyReasons, ", ")
		}

		return SignalNone, fmt.Sprintf("1h no signal (buy conditions: %d/2)", buyConditions)
	}

	// === SELL SIGNAL CONDITIONS ===
	if position.InPosition {
		sellConditions := 0
		var sellReasons []string

		// RSI overbought
		if hourly.RSI >= mts.config.RSIOverbought {
			sellConditions++
			sellReasons = append(sellReasons, fmt.Sprintf("RSI overbought (%.2f)", hourly.RSI))
		}

		// MACD bearish crossover (histogram turning negative)
		if hourly.MACDHistogram < 0 && hourly.MACD < hourly.MACDSignal {
			sellConditions++
			sellReasons = append(sellReasons, "MACD bearish crossover")
		}

		// Price near or above upper Bollinger Band
		if hourly.Price >= hourly.BBandsUpper*0.99 { // Within 1% of upper band
			sellConditions++
			sellReasons = append(sellReasons, "Price at upper BB")
		}

		// Need at least 2 out of 3 conditions
		if sellConditions >= 2 {
			return SignalSell, "1h SELL: " + strings.Join(sellReasons, ", ")
		}

		return SignalNone, fmt.Sprintf("1h no signal (sell conditions: %d/2)", sellConditions)
	}

	return SignalNone, "1h no signal"
}

// checkTrendAlignment ensures hourly signal aligns with daily trend
func (mts *MultiTimeframeStrategy) checkTrendAlignment(dailyTrend TrendDirection, hourlySignal Signal) bool {
	if dailyTrend == TrendNone {
		return true // No trend filter
	}

	if dailyTrend == TrendBullish && hourlySignal == SignalBuy {
		return true
	}

	if dailyTrend == TrendBearish && hourlySignal == SignalSell {
		return true
	}

	return false
}

// checkEntryPrecision validates 5-minute timeframe for precise entry
func (mts *MultiTimeframeStrategy) checkEntryPrecision(fiveMin IndicatorSnapshot, signal Signal) bool {
	if !fiveMin.RSIReady || !fiveMin.MACDReady {
		return false
	}

	if signal == SignalBuy {
		// For buy: RSI should still be oversold or recovering
		// MACD should be turning up
		return fiveMin.RSI < mts.config.RSINeutralHigh && fiveMin.MACDHistogram >= 0
	}

	if signal == SignalSell {
		// For sell: RSI should be overbought or weakening
		// MACD should be turning down
		return fiveMin.RSI > mts.config.RSINeutralLow && fiveMin.MACDHistogram <= 0
	}

	return false
}

// checkVolatility ensures volatility is within acceptable range
func (mts *MultiTimeframeStrategy) checkVolatility(hourly IndicatorSnapshot) bool {
	if !hourly.BBandsReady {
		return false
	}

	width := hourly.BBandsWidth

	// Volatility must be sufficient but not extreme
	return width >= mts.config.BBandsMinWidth && width <= mts.config.BBandsMaxWidth
}

// GetSignalReason returns explanation of the last signal
func (mts *MultiTimeframeStrategy) GetSignalReason() string {
	return mts.lastSignalReason
}

// Reset resets the strategy state
func (mts *MultiTimeframeStrategy) Reset() {
	mts.mtfManager.Reset()
	mts.lastSignalReason = ""
}

// GetMultiTimeframeManager returns the underlying manager (for debugging/monitoring)
func (mts *MultiTimeframeStrategy) GetMultiTimeframeManager() *MultiTimeframeManager {
	return mts.mtfManager
}

// IsReady returns true when all timeframes have sufficient data
func (mts *MultiTimeframeStrategy) IsReady() bool {
	return mts.mtfManager.IsReady()
}
