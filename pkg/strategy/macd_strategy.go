package strategy

import (
	"fmt"
	"time"

	"rsi-bot/pkg/indicators"
)

// MACDStrategy implements a trading strategy based on MACD crossovers
type MACDStrategy struct {
	indicator        indicators.Indicator
	lastSignalReason string

	// Track previous MACD values for crossover detection
	prevMACD   float64
	prevSignal float64
	initialized bool
}

// NewMACDStrategy creates a new MACD-based trading strategy
func NewMACDStrategy(indicator indicators.Indicator) (*MACDStrategy, error) {
	if indicator.Name() != "MACD" {
		return nil, fmt.Errorf("MACDStrategy requires MACD indicator, got %s", indicator.Name())
	}

	return &MACDStrategy{
		indicator:   indicator,
		initialized: false,
	}, nil
}

// Name returns the strategy identifier
func (s *MACDStrategy) Name() string {
	return "MACD"
}

// GetIndicator returns the underlying indicator
func (s *MACDStrategy) GetIndicator() indicators.Indicator {
	return s.indicator
}

// Update processes new price data
func (s *MACDStrategy) Update(price float64, volume float64, timestamp time.Time) error {
	return s.indicator.Update(price, timestamp)
}

// IsReady returns true when the strategy has enough data
func (s *MACDStrategy) IsReady() bool {
	return s.indicator.IsReady()
}

// GenerateSignal analyzes MACD crossovers and generates trading signals
func (s *MACDStrategy) GenerateSignal(ctx SignalContext) Signal {
	// Get MACD values from indicator data
	macdLine, hasMacd := ctx.IndicatorData[indicators.ValueKeyMACD]
	signalLine, hasSignal := ctx.IndicatorData[indicators.ValueKeySignal]
	histogram, hasHist := ctx.IndicatorData[indicators.ValueKeyHistogram]

	if !hasMacd || !hasSignal || !hasHist {
		s.lastSignalReason = "MACD values not available"
		return SignalNone
	}

	// Need at least 2 data points to detect crossover
	if !s.initialized {
		s.prevMACD = macdLine
		s.prevSignal = signalLine
		s.initialized = true
		s.lastSignalReason = "Initializing MACD crossover detection"
		return SignalNone
	}

	// Detect crossovers
	bullishCrossover := s.prevMACD <= s.prevSignal && macdLine > signalLine
	bearishCrossover := s.prevMACD >= s.prevSignal && macdLine < signalLine

	var signal Signal = SignalNone

	// BUY signal: Bullish crossover (MACD crosses above signal) AND no position
	if bullishCrossover && !ctx.Position.InPosition {
		s.lastSignalReason = fmt.Sprintf("MACD BULLISH CROSSOVER: MACD %.4f crossed above Signal %.4f, Histogram: %.4f",
			macdLine, signalLine, histogram)
		signal = SignalBuy
	} else if bearishCrossover && ctx.Position.InPosition {
		// SELL signal: Bearish crossover (MACD crosses below signal) AND holding position
		profitPercent := ((ctx.CurrentPrice - ctx.Position.EntryPrice) / ctx.Position.EntryPrice) * 100
		s.lastSignalReason = fmt.Sprintf("MACD BEARISH CROSSOVER: MACD %.4f crossed below Signal %.4f, Profit: %.2f%%",
			macdLine, signalLine, profitPercent)
		signal = SignalSell
	} else {
		// No crossover or wrong position state
		if ctx.Position.InPosition {
			profitPercent := ((ctx.CurrentPrice - ctx.Position.EntryPrice) / ctx.Position.EntryPrice) * 100
			s.lastSignalReason = fmt.Sprintf("HOLDING: MACD %.4f, Signal %.4f, Hist %.4f (%.2f%% profit)",
				macdLine, signalLine, histogram, profitPercent)
		} else {
			s.lastSignalReason = fmt.Sprintf("WAITING: MACD %.4f, Signal %.4f, Hist %.4f (no position)",
				macdLine, signalLine, histogram)
		}
	}

	// Update previous values for next crossover detection
	s.prevMACD = macdLine
	s.prevSignal = signalLine

	return signal
}

// GetSignalReason returns the explanation for the last signal
func (s *MACDStrategy) GetSignalReason() string {
	return s.lastSignalReason
}

// Reset resets the strategy state
func (s *MACDStrategy) Reset() {
	s.lastSignalReason = ""
	s.prevMACD = 0
	s.prevSignal = 0
	s.initialized = false
}

// GetCurrentMACD returns the current MACD line value
func (s *MACDStrategy) GetCurrentMACD() float64 {
	return s.prevMACD
}

// GetCurrentSignal returns the current signal line value
func (s *MACDStrategy) GetCurrentSignal() float64 {
	return s.prevSignal
}
