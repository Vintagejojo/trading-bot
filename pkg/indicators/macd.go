package indicators

import (
	"fmt"
	"time"
)

// MACD (Moving Average Convergence Divergence) indicator
// MACD = EMA(fast) - EMA(slow)
// Signal Line = EMA of MACD
// Histogram = MACD - Signal
type MACD struct {
	fastPeriod   int
	slowPeriod   int
	signalPeriod int

	prices      []float64
	timestamps  []time.Time
	macdLine    []float64
	signalLine  []float64
	histogram   []float64

	fastEMA   float64
	slowEMA   float64
	signalEMA float64

	fastMultiplier   float64
	slowMultiplier   float64
	signalMultiplier float64

	isReady bool
}

// NewMACD creates a new MACD indicator
// Standard parameters: fast=12, slow=26, signal=9
func NewMACD(fastPeriod, slowPeriod, signalPeriod int) (*MACD, error) {
	if fastPeriod <= 0 || slowPeriod <= 0 || signalPeriod <= 0 {
		return nil, fmt.Errorf("MACD periods must be positive, got fast=%d, slow=%d, signal=%d",
			fastPeriod, slowPeriod, signalPeriod)
	}

	if fastPeriod >= slowPeriod {
		return nil, fmt.Errorf("fast period (%d) must be less than slow period (%d)",
			fastPeriod, slowPeriod)
	}

	return &MACD{
		fastPeriod:       fastPeriod,
		slowPeriod:       slowPeriod,
		signalPeriod:     signalPeriod,
		prices:           make([]float64, 0, slowPeriod+50),
		timestamps:       make([]time.Time, 0, slowPeriod+50),
		macdLine:         make([]float64, 0, 100),
		signalLine:       make([]float64, 0, 100),
		histogram:        make([]float64, 0, 100),
		fastMultiplier:   2.0 / float64(fastPeriod+1),
		slowMultiplier:   2.0 / float64(slowPeriod+1),
		signalMultiplier: 2.0 / float64(signalPeriod+1),
		isReady:          false,
	}, nil
}

// Name returns the indicator identifier
func (m *MACD) Name() string {
	return "MACD"
}

// Update adds new price data and recalculates MACD
func (m *MACD) Update(price float64, timestamp time.Time) error {
	if price <= 0 {
		return fmt.Errorf("price must be positive, got %.8f", price)
	}

	m.prices = append(m.prices, price)
	m.timestamps = append(m.timestamps, timestamp)

	dataCount := len(m.prices)

	// Initialize EMAs when we have enough data for slow period
	if dataCount == m.slowPeriod {
		// Calculate initial SMA for both fast and slow
		m.fastEMA = m.calculateSMA(m.prices[dataCount-m.fastPeriod:])
		m.slowEMA = m.calculateSMA(m.prices)
	}

	// Update EMAs if we have enough initial data
	if dataCount >= m.slowPeriod {
		// Update fast EMA
		m.fastEMA = (price-m.fastEMA)*m.fastMultiplier + m.fastEMA

		// Update slow EMA
		m.slowEMA = (price-m.slowEMA)*m.slowMultiplier + m.slowEMA

		// Calculate MACD line
		macd := m.fastEMA - m.slowEMA
		m.macdLine = append(m.macdLine, macd)

		// Initialize signal line when we have enough MACD values
		if len(m.macdLine) == m.signalPeriod {
			m.signalEMA = m.calculateSMA(m.macdLine)
			m.isReady = true
		}

		// Update signal line if initialized
		if len(m.macdLine) >= m.signalPeriod {
			m.signalEMA = (macd-m.signalEMA)*m.signalMultiplier + m.signalEMA
			m.signalLine = append(m.signalLine, m.signalEMA)

			// Calculate histogram
			hist := macd - m.signalEMA
			m.histogram = append(m.histogram, hist)
		}
	}

	// Keep buffer size manageable (keep last 200 values)
	if len(m.prices) > 200 {
		m.prices = m.prices[1:]
		m.timestamps = m.timestamps[1:]
	}
	if len(m.macdLine) > 200 {
		m.macdLine = m.macdLine[1:]
	}
	if len(m.signalLine) > 200 {
		m.signalLine = m.signalLine[1:]
	}
	if len(m.histogram) > 200 {
		m.histogram = m.histogram[1:]
	}

	return nil
}

// calculateSMA calculates Simple Moving Average
func (m *MACD) calculateSMA(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// GetValue returns current MACD values
func (m *MACD) GetValue() (map[string]float64, bool) {
	if !m.isReady {
		return nil, false
	}

	macdIdx := len(m.macdLine) - 1
	signalIdx := len(m.signalLine) - 1
	histIdx := len(m.histogram) - 1

	if macdIdx < 0 || signalIdx < 0 || histIdx < 0 {
		return nil, false
	}

	return map[string]float64{
		ValueKeyMACD:      m.macdLine[macdIdx],
		ValueKeySignal:    m.signalLine[signalIdx],
		ValueKeyHistogram: m.histogram[histIdx],
	}, true
}

// IsReady returns true when indicator has enough data
func (m *MACD) IsReady() bool {
	return m.isReady
}

// Reset clears all data
func (m *MACD) Reset() {
	m.prices = make([]float64, 0, m.slowPeriod+50)
	m.timestamps = make([]time.Time, 0, m.slowPeriod+50)
	m.macdLine = make([]float64, 0, 100)
	m.signalLine = make([]float64, 0, 100)
	m.histogram = make([]float64, 0, 100)
	m.fastEMA = 0
	m.slowEMA = 0
	m.signalEMA = 0
	m.isReady = false
}

// GetDataCount returns number of price points stored
func (m *MACD) GetDataCount() int {
	return len(m.prices)
}

// GetMACDDataCount returns number of MACD line points calculated
func (m *MACD) GetMACDDataCount() int {
	return len(m.macdLine)
}

// GetHistorySize returns the required number of periods for full calculation
func (m *MACD) GetHistorySize() int {
	return m.slowPeriod + m.signalPeriod
}
