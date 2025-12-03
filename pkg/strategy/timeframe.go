package strategy

import (
	"fmt"
	"log"
	"time"
)

// Timeframe represents a chart timeframe
type Timeframe string

const (
	Timeframe5m  Timeframe = "5m"  // 5 minutes
	Timeframe15m Timeframe = "15m" // 15 minutes
	Timeframe1h  Timeframe = "1h"  // 1 hour
	Timeframe4h  Timeframe = "4h"  // 4 hours
	Timeframe1d  Timeframe = "1d"  // 1 day
)

// GetDuration returns the duration for a timeframe
func (tf Timeframe) GetDuration() (time.Duration, error) {
	switch tf {
	case Timeframe5m:
		return 5 * time.Minute, nil
	case Timeframe15m:
		return 15 * time.Minute, nil
	case Timeframe1h:
		return time.Hour, nil
	case Timeframe4h:
		return 4 * time.Hour, nil
	case Timeframe1d:
		return 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown timeframe: %s", tf)
	}
}

// String returns the string representation
func (tf Timeframe) String() string {
	return string(tf)
}

// OHLCV represents a candlestick with Open, High, Low, Close, Volume
type OHLCV struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// TimeframeData stores candlestick data for a specific timeframe
type TimeframeData struct {
	Timeframe   Timeframe
	Candles     []OHLCV
	MaxCandles  int // Maximum number of candles to keep
	currentBar  *OHLCV // Current incomplete candle being built
	barStartTime time.Time
}

// NewTimeframeData creates a new timeframe data container
func NewTimeframeData(tf Timeframe, maxCandles int) *TimeframeData {
	return &TimeframeData{
		Timeframe:  tf,
		Candles:    make([]OHLCV, 0, maxCandles),
		MaxCandles: maxCandles,
	}
}

// Update aggregates tick data into the appropriate timeframe candle
// This is called for every price update (e.g., from 1-minute klines)
func (td *TimeframeData) Update(price float64, volume float64, timestamp time.Time) error {
	duration, err := td.Timeframe.GetDuration()
	if err != nil {
		return err
	}

	// Calculate the start time of the current bar
	barStart := timestamp.Truncate(duration)

	// If this is a new bar or first update
	if td.currentBar == nil || barStart.After(td.barStartTime) {
		// Save the previous completed bar if it exists
		if td.currentBar != nil {
			td.Candles = append(td.Candles, *td.currentBar)
			log.Printf("[%s] Bar completed! Total candles: %d", td.Timeframe, len(td.Candles))

			// Keep only the last MaxCandles
			if len(td.Candles) > td.MaxCandles {
				td.Candles = td.Candles[1:]
			}
		}

		// Start a new bar
		td.currentBar = &OHLCV{
			Timestamp: barStart,
			Open:      price,
			High:      price,
			Low:       price,
			Close:     price,
			Volume:    volume,
		}
		td.barStartTime = barStart
		log.Printf("[%s] New bar started at %s, price=%.2f", td.Timeframe, barStart.Format("15:04:05"), price)
	} else {
		// Update the current bar
		if price > td.currentBar.High {
			td.currentBar.High = price
		}
		if price < td.currentBar.Low {
			td.currentBar.Low = price
		}
		td.currentBar.Close = price
		td.currentBar.Volume += volume
	}

	return nil
}

// GetLatestCandle returns the most recent completed candle
func (td *TimeframeData) GetLatestCandle() (*OHLCV, bool) {
	if len(td.Candles) == 0 {
		return nil, false
	}
	return &td.Candles[len(td.Candles)-1], true
}

// GetCurrentCandle returns the incomplete current candle
func (td *TimeframeData) GetCurrentCandle() (*OHLCV, bool) {
	if td.currentBar == nil {
		return nil, false
	}
	return td.currentBar, true
}

// GetCandles returns all completed candles
func (td *TimeframeData) GetCandles() []OHLCV {
	return td.Candles
}

// GetCandleCount returns the number of completed candles
func (td *TimeframeData) GetCandleCount() int {
	return len(td.Candles)
}

// IsReady returns true if we have enough candles for analysis
func (td *TimeframeData) IsReady(minCandles int) bool {
	return len(td.Candles) >= minCandles
}

// Reset clears all candle data
func (td *TimeframeData) Reset() {
	td.Candles = make([]OHLCV, 0, td.MaxCandles)
	td.currentBar = nil
	td.barStartTime = time.Time{}
}
