package strategy

import (
	"fmt"
	"rsi-bot/pkg/indicators"
)

// BollingerBandsStrategy implements a trading strategy based on Bollinger Bands
// Buy when price touches or crosses below lower band
// Sell when price touches or crosses above upper band
type BollingerBandsStrategy struct {
	indicator        indicators.Indicator
	lastSignalReason string

	// Track previous price position for band touch detection
	prevPrice      float64
	prevLower      float64
	prevUpper      float64
	initialized    bool
}

// NewBollingerBandsStrategy creates a new Bollinger Bands trading strategy
func NewBollingerBandsStrategy(indicator indicators.Indicator) (*BollingerBandsStrategy, error) {
	if indicator.Name() != "BBands" {
		return nil, fmt.Errorf("BollingerBandsStrategy requires BBands indicator, got %s", indicator.Name())
	}

	return &BollingerBandsStrategy{
		indicator:   indicator,
		initialized: false,
	}, nil
}

// Name returns the strategy identifier
func (s *BollingerBandsStrategy) Name() string {
	return "BBands"
}

// GetIndicator returns the underlying indicator
func (s *BollingerBandsStrategy) GetIndicator() indicators.Indicator {
	return s.indicator
}

// GenerateSignal analyzes Bollinger Bands and generates trading signals
func (s *BollingerBandsStrategy) GenerateSignal(ctx SignalContext) Signal {
	// Get Bollinger Bands values from indicator data
	upper, hasUpper := ctx.IndicatorData[indicators.ValueKeyUpper]
	middle, hasMiddle := ctx.IndicatorData[indicators.ValueKeyMiddle]
	lower, hasLower := ctx.IndicatorData[indicators.ValueKeyLower]

	if !hasUpper || !hasMiddle || !hasLower {
		s.lastSignalReason = "Bollinger Bands values not available"
		return SignalNone
	}

	currentPrice := ctx.CurrentPrice

	// Initialize tracking variables
	if !s.initialized {
		s.prevPrice = currentPrice
		s.prevLower = lower
		s.prevUpper = upper
		s.initialized = true
		s.lastSignalReason = "Initializing Bollinger Bands tracking"
		return SignalNone
	}

	// Calculate band width percentage (volatility measure)
	bandWidth := ((upper - lower) / middle) * 100

	// Detect band touches/crosses
	// Lower band touch: price was above lower band and now at/below it
	lowerBandTouch := s.prevPrice > s.prevLower && currentPrice <= lower

	// Upper band touch: price was below upper band and now at/above it
	upperBandTouch := s.prevPrice < s.prevUpper && currentPrice >= upper

	var signal Signal = SignalNone

	// BUY signal: Price touches/crosses lower band AND no position
	if lowerBandTouch && !ctx.Position.InPosition {
		percentBelow := ((lower - currentPrice) / middle) * 100
		s.lastSignalReason = fmt.Sprintf("LOWER BAND TOUCH: Price %.8f touched lower band %.8f (%.2f%% below middle, width: %.2f%%)",
			currentPrice, lower, percentBelow, bandWidth)
		signal = SignalBuy
	} else if upperBandTouch && ctx.Position.InPosition {
		// SELL signal: Price touches/crosses upper band AND holding position
		profitPercent := ((currentPrice - ctx.Position.EntryPrice) / ctx.Position.EntryPrice) * 100
		percentAbove := ((currentPrice - upper) / middle) * 100
		s.lastSignalReason = fmt.Sprintf("UPPER BAND TOUCH: Price %.8f touched upper band %.8f (%.2f%% above middle, Profit: %.2f%%)",
			currentPrice, upper, percentAbove, profitPercent)
		signal = SignalSell
	} else {
		// No band touch or wrong position state
		// Calculate price position within bands (percent B)
		percentB := ((currentPrice - lower) / (upper - lower)) * 100

		if ctx.Position.InPosition {
			profitPercent := ((currentPrice - ctx.Position.EntryPrice) / ctx.Position.EntryPrice) * 100
			s.lastSignalReason = fmt.Sprintf("HOLDING: Price %.8f, %%B: %.1f%%, Width: %.2f%% (%.2f%% profit)",
				currentPrice, percentB, bandWidth, profitPercent)
		} else {
			s.lastSignalReason = fmt.Sprintf("WAITING: Price %.8f, %%B: %.1f%%, Width: %.2f%% (no position)",
				currentPrice, percentB, bandWidth)
		}
	}

	// Update previous values for next detection
	s.prevPrice = currentPrice
	s.prevLower = lower
	s.prevUpper = upper

	return signal
}

// GetSignalReason returns the explanation for the last signal
func (s *BollingerBandsStrategy) GetSignalReason() string {
	return s.lastSignalReason
}

// Reset resets the strategy state
func (s *BollingerBandsStrategy) Reset() {
	s.lastSignalReason = ""
	s.prevPrice = 0
	s.prevLower = 0
	s.prevUpper = 0
	s.initialized = false
}

// GetCurrentBands returns the current band values
func (s *BollingerBandsStrategy) GetCurrentBands() (upper, middle, lower float64) {
	return s.prevUpper, 0, s.prevLower // middle not tracked, can be added if needed
}
