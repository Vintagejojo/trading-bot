package strategy

import (
	"time"

	"rsi-bot/pkg/indicators"
	"rsi-bot/pkg/models"
)

// Signal represents a trading signal
type Signal int

const (
	SignalNone Signal = iota // No action
	SignalBuy                // Buy signal
	SignalSell               // Sell signal
)

func (s Signal) String() string {
	switch s {
	case SignalBuy:
		return "BUY"
	case SignalSell:
		return "SELL"
	default:
		return "NONE"
	}
}

// SignalContext provides context for signal generation
type SignalContext struct {
	CurrentPrice  float64
	Position      *models.Position
	IndicatorData map[string]float64
}

// Strategy defines the interface for trading strategies
type Strategy interface {
	// Name returns the strategy identifier
	Name() string

	// GetIndicator returns the indicator used by this strategy
	GetIndicator() indicators.Indicator

	// Update processes new price data (updates all indicators)
	Update(price float64, volume float64, timestamp time.Time) error

	// IsReady returns true when the strategy has enough data to generate signals
	IsReady() bool

	// GenerateSignal analyzes indicator data and returns a trading signal
	GenerateSignal(ctx SignalContext) Signal

	// GetSignalReason returns a human-readable explanation of the last signal
	GetSignalReason() string

	// Reset resets the strategy state
	Reset()
}
