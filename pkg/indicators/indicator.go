package indicators

import "time"

// Indicator represents a technical indicator calculator
// This is the base interface that all indicators must implement
type Indicator interface {
	// Name returns the indicator identifier (e.g., "RSI", "MACD", "BBands")
	Name() string

	// Update adds new price data and recalculates the indicator
	// Returns error if the update fails
	Update(price float64, timestamp time.Time) error

	// GetValue returns the current indicator value(s)
	// For single-value indicators (RSI), returns map with one key: {"value": 65.5}
	// For multi-value indicators (MACD), returns map with multiple keys:
	//   {"macd": 2.5, "signal": 1.8, "histogram": 0.7}
	// The second return value indicates if the indicator has enough data for a valid calculation
	GetValue() (map[string]float64, bool)

	// IsReady returns true when the indicator has enough data for valid calculations
	// For example, RSI needs (period + 1) data points
	IsReady() bool

	// Reset clears all historical data and resets the indicator to initial state
	Reset()

	// GetDataCount returns the number of data points currently stored
	GetDataCount() int
}

// Common indicator value keys for consistency
const (
	ValueKeyRSI       = "rsi"
	ValueKeyMACD      = "macd"
	ValueKeySignal    = "signal"
	ValueKeyHistogram = "histogram"
	ValueKeyUpper     = "upper"
	ValueKeyMiddle    = "middle"
	ValueKeyLower     = "lower"
	ValueKeyStochRSI  = "stoch_rsi"
	ValueKeyStochK    = "stoch_k"
	ValueKeyStochD    = "stoch_d"
)
