package indicators

import (
	"fmt"
	"math"
	"time"
)

// BollingerBands indicator
// Upper Band = SMA + (stdDev * multiplier)
// Middle Band = SMA
// Lower Band = SMA - (stdDev * multiplier)
type BollingerBands struct {
	period     int
	stdDevMult float64

	prices     []float64
	timestamps []time.Time
	isReady    bool
}

// NewBollingerBands creates a new Bollinger Bands indicator
// Standard parameters: period=20, stdDevMult=2.0
func NewBollingerBands(period int, stdDevMult float64) (*BollingerBands, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}

	if stdDevMult <= 0 {
		return nil, fmt.Errorf("standard deviation multiplier must be positive, got %.2f", stdDevMult)
	}

	return &BollingerBands{
		period:     period,
		stdDevMult: stdDevMult,
		prices:     make([]float64, 0, period+50),
		timestamps: make([]time.Time, 0, period+50),
		isReady:    false,
	}, nil
}

// Name returns the indicator identifier
func (bb *BollingerBands) Name() string {
	return "BBands"
}

// Update adds new price data and recalculates Bollinger Bands
func (bb *BollingerBands) Update(price float64, timestamp time.Time) error {
	if price <= 0 {
		return fmt.Errorf("price must be positive, got %.8f", price)
	}

	bb.prices = append(bb.prices, price)
	bb.timestamps = append(bb.timestamps, timestamp)

	// Mark as ready when we have enough data
	if len(bb.prices) >= bb.period {
		bb.isReady = true
	}

	// Keep buffer size manageable (keep last period + 50 values)
	if len(bb.prices) > bb.period+50 {
		bb.prices = bb.prices[1:]
		bb.timestamps = bb.timestamps[1:]
	}

	return nil
}

// calculateSMA calculates Simple Moving Average for the given slice
func (bb *BollingerBands) calculateSMA(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev calculates standard deviation
func (bb *BollingerBands) calculateStdDev(values []float64, mean float64) float64 {
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	variance := sumSquares / float64(len(values))
	return math.Sqrt(variance)
}

// GetValue returns current Bollinger Bands values
func (bb *BollingerBands) GetValue() (map[string]float64, bool) {
	if !bb.isReady {
		return nil, false
	}

	// Get the last 'period' prices
	dataCount := len(bb.prices)
	startIdx := dataCount - bb.period
	if startIdx < 0 {
		return nil, false
	}

	recentPrices := bb.prices[startIdx:]

	// Calculate middle band (SMA)
	middle := bb.calculateSMA(recentPrices)

	// Calculate standard deviation
	stdDev := bb.calculateStdDev(recentPrices, middle)

	// Calculate upper and lower bands
	upper := middle + (stdDev * bb.stdDevMult)
	lower := middle - (stdDev * bb.stdDevMult)

	return map[string]float64{
		ValueKeyUpper:  upper,
		ValueKeyMiddle: middle,
		ValueKeyLower:  lower,
	}, true
}

// IsReady returns true when indicator has enough data
func (bb *BollingerBands) IsReady() bool {
	return bb.isReady
}

// Reset clears all data
func (bb *BollingerBands) Reset() {
	bb.prices = make([]float64, 0, bb.period+50)
	bb.timestamps = make([]time.Time, 0, bb.period+50)
	bb.isReady = false
}

// GetDataCount returns number of data points stored
func (bb *BollingerBands) GetDataCount() int {
	return len(bb.prices)
}

// GetBandWidth returns the width of the bands (useful for volatility analysis)
func (bb *BollingerBands) GetBandWidth() (float64, bool) {
	values, isValid := bb.GetValue()
	if !isValid {
		return 0, false
	}

	upper := values[ValueKeyUpper]
	lower := values[ValueKeyLower]
	middle := values[ValueKeyMiddle]

	// Band width as percentage of middle band
	width := ((upper - lower) / middle) * 100
	return width, true
}

// GetPercentB returns %B indicator (price position within bands)
// %B = (price - lower) / (upper - lower)
// Values > 1 mean price is above upper band
// Values < 0 mean price is below lower band
func (bb *BollingerBands) GetPercentB(currentPrice float64) (float64, bool) {
	values, isValid := bb.GetValue()
	if !isValid {
		return 0, false
	}

	upper := values[ValueKeyUpper]
	lower := values[ValueKeyLower]

	if upper == lower {
		return 0, false
	}

	percentB := (currentPrice - lower) / (upper - lower)
	return percentB, true
}
