package strategy

import (
	"fmt"
	"sync"
	"time"

	"rsi-bot/pkg/indicators"
)

// TimeframeIndicators holds indicators for a specific timeframe
type TimeframeIndicators struct {
	Timeframe Timeframe
	RSI       *indicators.RSI
	MACD      *indicators.MACD
	BBands    *indicators.BollingerBands
}

// MultiTimeframeManager manages data and indicators across multiple timeframes
type MultiTimeframeManager struct {
	mu sync.RWMutex

	// Data storage for each timeframe
	TimeframeData map[Timeframe]*TimeframeData

	// Indicators for each timeframe
	Indicators map[Timeframe]*TimeframeIndicators

	// Configuration
	config MultiTimeframeConfig
}

// MultiTimeframeConfig configures the multi-timeframe manager
type MultiTimeframeConfig struct {
	// Timeframes to track (e.g., 5m, 1h, 1d)
	Timeframes []Timeframe

	// Maximum candles to keep per timeframe
	MaxCandles int

	// Indicator parameters
	RSIPeriod       int
	MACDFast        int
	MACDSlow        int
	MACDSignal      int
	BBandsPeriod    int
	BBandsStdDev    float64
}

// DefaultMultiTimeframeConfig returns sensible defaults
func DefaultMultiTimeframeConfig() MultiTimeframeConfig {
	return MultiTimeframeConfig{
		Timeframes: []Timeframe{Timeframe5m, Timeframe1h, Timeframe1d},
		MaxCandles: 200, // Keep 200 candles per timeframe
		RSIPeriod:  14,
		MACDFast:   12,
		MACDSlow:   26,
		MACDSignal: 9,
		BBandsPeriod: 20,
		BBandsStdDev: 2.0,
	}
}

// NewMultiTimeframeManager creates a new multi-timeframe manager
func NewMultiTimeframeManager(config MultiTimeframeConfig) (*MultiTimeframeManager, error) {
	if len(config.Timeframes) == 0 {
		return nil, fmt.Errorf("at least one timeframe must be specified")
	}

	mtf := &MultiTimeframeManager{
		TimeframeData: make(map[Timeframe]*TimeframeData),
		Indicators:    make(map[Timeframe]*TimeframeIndicators),
		config:        config,
	}

	// Initialize timeframe data and indicators
	for _, tf := range config.Timeframes {
		// Create timeframe data container
		mtf.TimeframeData[tf] = NewTimeframeData(tf, config.MaxCandles)

		// Create indicators for this timeframe
		rsi, err := indicators.NewRSI(config.RSIPeriod)
		if err != nil {
			return nil, fmt.Errorf("failed to create RSI for %s: %w", tf, err)
		}

		macd, err := indicators.NewMACD(config.MACDFast, config.MACDSlow, config.MACDSignal)
		if err != nil {
			return nil, fmt.Errorf("failed to create MACD for %s: %w", tf, err)
		}

		bbands, err := indicators.NewBollingerBands(config.BBandsPeriod, config.BBandsStdDev)
		if err != nil {
			return nil, fmt.Errorf("failed to create Bollinger Bands for %s: %w", tf, err)
		}

		mtf.Indicators[tf] = &TimeframeIndicators{
			Timeframe: tf,
			RSI:       rsi,
			MACD:      macd,
			BBands:    bbands,
		}
	}

	return mtf, nil
}

// Update processes new price data and updates all timeframes
// This should be called with each new price tick (e.g., from 1-minute klines)
func (mtf *MultiTimeframeManager) Update(price float64, volume float64, timestamp time.Time) error {
	mtf.mu.Lock()
	defer mtf.mu.Unlock()

	// Update each timeframe's data
	for tf, tfData := range mtf.TimeframeData {
		if err := tfData.Update(price, volume, timestamp); err != nil {
			return fmt.Errorf("failed to update %s timeframe: %w", tf, err)
		}

		// Update indicators if we have a completed candle
		if candle, ok := tfData.GetLatestCandle(); ok {
			tfIndicators := mtf.Indicators[tf]

			// Update all indicators with the close price
			if err := tfIndicators.RSI.Update(candle.Close, candle.Timestamp); err != nil {
				return fmt.Errorf("failed to update RSI for %s: %w", tf, err)
			}

			if err := tfIndicators.MACD.Update(candle.Close, candle.Timestamp); err != nil {
				return fmt.Errorf("failed to update MACD for %s: %w", tf, err)
			}

			if err := tfIndicators.BBands.Update(candle.Close, candle.Timestamp); err != nil {
				return fmt.Errorf("failed to update BBands for %s: %w", tf, err)
			}
		}
	}

	return nil
}

// GetIndicatorValues returns all indicator values for a specific timeframe
func (mtf *MultiTimeframeManager) GetIndicatorValues(tf Timeframe) (IndicatorSnapshot, bool) {
	mtf.mu.RLock()
	defer mtf.mu.RUnlock()

	tfIndicators, exists := mtf.Indicators[tf]
	if !exists {
		return IndicatorSnapshot{}, false
	}

	snapshot := IndicatorSnapshot{
		Timeframe: tf,
		Timestamp: time.Now(),
	}

	// Get RSI
	if rsiVals, ready := tfIndicators.RSI.GetValue(); ready {
		snapshot.RSI = rsiVals[indicators.ValueKeyRSI]
		snapshot.RSIReady = true
	}

	// Get MACD
	if macdVals, ready := tfIndicators.MACD.GetValue(); ready {
		snapshot.MACD = macdVals[indicators.ValueKeyMACD]
		snapshot.MACDSignal = macdVals[indicators.ValueKeySignal]
		snapshot.MACDHistogram = macdVals[indicators.ValueKeyHistogram]
		snapshot.MACDReady = true
	}

	// Get Bollinger Bands
	if bbandsVals, ready := tfIndicators.BBands.GetValue(); ready {
		snapshot.BBandsUpper = bbandsVals[indicators.ValueKeyUpper]
		snapshot.BBandsMiddle = bbandsVals[indicators.ValueKeyMiddle]
		snapshot.BBandsLower = bbandsVals[indicators.ValueKeyLower]
		snapshot.BBandsReady = true

		// Calculate band width (volatility indicator)
		if width, ok := tfIndicators.BBands.GetBandWidth(); ok {
			snapshot.BBandsWidth = width
		}
	}

	// Get current price from latest candle
	if tfData, ok := mtf.TimeframeData[tf]; ok {
		if candle, hasCandle := tfData.GetLatestCandle(); hasCandle {
			snapshot.Price = candle.Close
		}
	}

	return snapshot, snapshot.RSIReady || snapshot.MACDReady || snapshot.BBandsReady
}

// GetAllSnapshots returns indicator snapshots for all timeframes
func (mtf *MultiTimeframeManager) GetAllSnapshots() map[Timeframe]IndicatorSnapshot {
	mtf.mu.RLock()
	defer mtf.mu.RUnlock()

	snapshots := make(map[Timeframe]IndicatorSnapshot)
	for _, tf := range mtf.config.Timeframes {
		if snapshot, ok := mtf.GetIndicatorValues(tf); ok {
			snapshots[tf] = snapshot
		}
	}

	return snapshots
}

// IsReady returns true if all timeframes have enough data
func (mtf *MultiTimeframeManager) IsReady() bool {
	mtf.mu.RLock()
	defer mtf.mu.RUnlock()

	for _, tfIndicators := range mtf.Indicators {
		// At least one indicator must be ready for each timeframe
		if !tfIndicators.RSI.IsReady() && !tfIndicators.MACD.IsReady() && !tfIndicators.BBands.IsReady() {
			return false
		}
	}

	return true
}

// Reset clears all data and indicators
func (mtf *MultiTimeframeManager) Reset() {
	mtf.mu.Lock()
	defer mtf.mu.Unlock()

	for _, tfData := range mtf.TimeframeData {
		tfData.Reset()
	}

	for _, tfIndicators := range mtf.Indicators {
		tfIndicators.RSI.Reset()
		tfIndicators.MACD.Reset()
		tfIndicators.BBands.Reset()
	}
}

// IndicatorSnapshot represents all indicator values at a specific timeframe
type IndicatorSnapshot struct {
	Timeframe Timeframe
	Timestamp time.Time
	Price     float64

	// RSI values
	RSI      float64
	RSIReady bool

	// MACD values
	MACD          float64
	MACDSignal    float64
	MACDHistogram float64
	MACDReady     bool

	// Bollinger Bands values
	BBandsUpper  float64
	BBandsMiddle float64
	BBandsLower  float64
	BBandsWidth  float64 // Volatility indicator
	BBandsReady  bool
}

// String returns a human-readable representation
func (is IndicatorSnapshot) String() string {
	return fmt.Sprintf(
		"[%s] Price: %.8f | RSI: %.2f | MACD: %.4f/%.4f/%.4f | BBands: %.8f/%.8f/%.8f (width: %.2f%%)",
		is.Timeframe,
		is.Price,
		is.RSI,
		is.MACD, is.MACDSignal, is.MACDHistogram,
		is.BBandsUpper, is.BBandsMiddle, is.BBandsLower,
		is.BBandsWidth,
	)
}
