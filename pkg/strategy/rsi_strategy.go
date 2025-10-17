package strategy

import (
	"fmt"
	"rsi-bot/pkg/indicators"
)

// RSIStrategy implements a trading strategy based on RSI overbought/oversold levels
type RSIStrategy struct {
	indicator       indicators.Indicator
	overboughtLevel float64
	oversoldLevel   float64
	lastSignalReason string
}

// NewRSIStrategy creates a new RSI-based trading strategy
func NewRSIStrategy(indicator indicators.Indicator, overboughtLevel, oversoldLevel float64) (*RSIStrategy, error) {
	if indicator.Name() != "RSI" {
		return nil, fmt.Errorf("RSIStrategy requires RSI indicator, got %s", indicator.Name())
	}

	if overboughtLevel <= oversoldLevel {
		return nil, fmt.Errorf("overbought level (%.1f) must be greater than oversold level (%.1f)",
			overboughtLevel, oversoldLevel)
	}

	return &RSIStrategy{
		indicator:       indicator,
		overboughtLevel: overboughtLevel,
		oversoldLevel:   oversoldLevel,
	}, nil
}

// Name returns the strategy identifier
func (s *RSIStrategy) Name() string {
	return "RSI"
}

// GetIndicator returns the underlying indicator
func (s *RSIStrategy) GetIndicator() indicators.Indicator {
	return s.indicator
}

// GenerateSignal analyzes RSI and generates trading signals
func (s *RSIStrategy) GenerateSignal(ctx SignalContext) Signal {
	// Get RSI value from indicator data
	rsi, ok := ctx.IndicatorData[indicators.ValueKeyRSI]
	if !ok {
		s.lastSignalReason = "RSI value not available"
		return SignalNone
	}

	// SELL signal: RSI overbought AND we have a position
	if rsi >= s.overboughtLevel && ctx.Position.InPosition {
		profitPercent := ((ctx.CurrentPrice - ctx.Position.EntryPrice) / ctx.Position.EntryPrice) * 100
		s.lastSignalReason = fmt.Sprintf("RSI %.2f >= %.1f (OVERBOUGHT), Profit: %.2f%%",
			rsi, s.overboughtLevel, profitPercent)
		return SignalSell
	}

	// BUY signal: RSI oversold AND we don't have a position
	if rsi <= s.oversoldLevel && !ctx.Position.InPosition {
		s.lastSignalReason = fmt.Sprintf("RSI %.2f <= %.1f (OVERSOLD)",
			rsi, s.oversoldLevel)
		return SignalBuy
	}

	// No signal
	if ctx.Position.InPosition {
		profitPercent := ((ctx.CurrentPrice - ctx.Position.EntryPrice) / ctx.Position.EntryPrice) * 100
		s.lastSignalReason = fmt.Sprintf("HOLDING: RSI %.2f (%.2f%% profit)", rsi, profitPercent)
	} else {
		s.lastSignalReason = fmt.Sprintf("WAITING: RSI %.2f (no position)", rsi)
	}

	return SignalNone
}

// GetSignalReason returns the explanation for the last signal
func (s *RSIStrategy) GetSignalReason() string {
	return s.lastSignalReason
}

// Reset resets the strategy state
func (s *RSIStrategy) Reset() {
	s.lastSignalReason = ""
}

// GetOverboughtLevel returns the overbought threshold
func (s *RSIStrategy) GetOverboughtLevel() float64 {
	return s.overboughtLevel
}

// GetOversoldLevel returns the oversold threshold
func (s *RSIStrategy) GetOversoldLevel() float64 {
	return s.oversoldLevel
}

// SetOverboughtLevel updates the overbought threshold
func (s *RSIStrategy) SetOverboughtLevel(level float64) error {
	if level <= s.oversoldLevel {
		return fmt.Errorf("overbought level (%.1f) must be greater than oversold level (%.1f)",
			level, s.oversoldLevel)
	}
	s.overboughtLevel = level
	return nil
}

// SetOversoldLevel updates the oversold threshold
func (s *RSIStrategy) SetOversoldLevel(level float64) error {
	if level >= s.overboughtLevel {
		return fmt.Errorf("oversold level (%.1f) must be less than overbought level (%.1f)",
			level, s.overboughtLevel)
	}
	s.oversoldLevel = level
	return nil
}
