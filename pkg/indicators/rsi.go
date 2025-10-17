package indicators

import (
	"fmt"
	"math"
	"time"
)

// RSI implements the Relative Strength Index indicator
type RSI struct {
	period     int
	closes     []float64
	timestamps []time.Time
	lastRSI    float64
	isReady    bool
}

// NewRSI creates a new RSI indicator with the specified period
// Typical periods: 14 (default), 9, 25
func NewRSI(period int) (*RSI, error) {
	if period < 2 {
		return nil, fmt.Errorf("RSI period must be at least 2, got %d", period)
	}

	return &RSI{
		period:     period,
		closes:     make([]float64, 0),
		timestamps: make([]time.Time, 0),
		lastRSI:    50.0, // Neutral value when not ready
		isReady:    false,
	}, nil
}

// Name returns the indicator identifier
func (r *RSI) Name() string {
	return "RSI"
}

// Update adds new price data and recalculates RSI
func (r *RSI) Update(price float64, timestamp time.Time) error {
	if price <= 0 {
		return fmt.Errorf("price must be positive, got %.8f", price)
	}

	// Add new price data
	r.closes = append(r.closes, price)
	r.timestamps = append(r.timestamps, timestamp)

	// Keep only what we need (period + buffer for accuracy)
	// Buffer of 20 helps with smoothing and accuracy
	maxKeep := r.period + 20
	if len(r.closes) > maxKeep {
		r.closes = r.closes[len(r.closes)-maxKeep:]
		r.timestamps = r.timestamps[len(r.timestamps)-maxKeep:]
	}

	// Calculate RSI if we have enough data
	if len(r.closes) >= r.period+1 {
		rsi, err := r.calculate()
		if err != nil {
			return fmt.Errorf("RSI calculation failed: %w", err)
		}
		r.lastRSI = rsi
		r.isReady = true
	}

	return nil
}

// GetValue returns the current RSI value
// Returns (map with "rsi" key, true) if ready, (map with neutral value, false) if not ready
func (r *RSI) GetValue() (map[string]float64, bool) {
	return map[string]float64{
		ValueKeyRSI: r.lastRSI,
	}, r.isReady
}

// IsReady returns true when enough data exists for valid RSI calculation
func (r *RSI) IsReady() bool {
	return r.isReady
}

// Reset clears all historical data
func (r *RSI) Reset() {
	r.closes = make([]float64, 0)
	r.timestamps = make([]time.Time, 0)
	r.lastRSI = 50.0
	r.isReady = false
}

// GetDataCount returns the number of price points currently stored
func (r *RSI) GetDataCount() int {
	return len(r.closes)
}

// calculate computes the RSI value using the standard formula
// RSI = 100 - (100 / (1 + RS))
// where RS = Average Gain / Average Loss over the period
func (r *RSI) calculate() (float64, error) {
	if len(r.closes) < r.period+1 {
		return 50.0, fmt.Errorf("insufficient data: need %d points, have %d", r.period+1, len(r.closes))
	}

	gains := 0.0
	losses := 0.0

	// Calculate gains and losses over the period
	for i := len(r.closes) - r.period; i < len(r.closes); i++ {
		change := r.closes[i] - r.closes[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += math.Abs(change)
		}
	}

	// Handle edge case: no losses means RSI = 100
	if losses == 0 {
		return 100.0, nil
	}

	// Handle edge case: no gains means RSI = 0
	if gains == 0 {
		return 0.0, nil
	}

	// Standard RSI calculation
	avgGain := gains / float64(r.period)
	avgLoss := losses / float64(r.period)
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi, nil
}

// GetPeriod returns the RSI period setting
func (r *RSI) GetPeriod() int {
	return r.period
}

// GetRequiredDataPoints returns how many more data points are needed
// Returns 0 if already ready
func (r *RSI) GetRequiredDataPoints() int {
	needed := (r.period + 1) - len(r.closes)
	if needed < 0 {
		return 0
	}
	return needed
}

// GetLastTimestamp returns the timestamp of the most recent data point
func (r *RSI) GetLastTimestamp() (time.Time, bool) {
	if len(r.timestamps) == 0 {
		return time.Time{}, false
	}
	return r.timestamps[len(r.timestamps)-1], true
}
