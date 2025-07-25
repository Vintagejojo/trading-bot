package calculator

import "math"

type RSI struct {
	period int
	closes []float64
}

func NewRSI(period int) *RSI {
	return &RSI{
		period: period,
		closes: make([]float64, 0),
	}
}

func (r *RSI) AddPrice(price float64) {
	r.closes = append(r.closes, price)

	// Keep only what we need (period + buffer for accuracy)
	maxKeep := r.period + 20
	if len(r.closes) > maxKeep {
		r.closes = r.closes[len(r.closes)-maxKeep:]
	}
}

func (r *RSI) Calculate() (float64, bool) {
	if len(r.closes) < r.period+1 {
		return 50.0, false // Not enough data
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

	if losses == 0 {
		return 100.0, true
	}

	avgGain := gains / float64(r.period)
	avgLoss := losses / float64(r.period)
	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi, true
}

func (r *RSI) HasEnoughData() bool {
	return len(r.closes) >= r.period+1
}

func (r *RSI) GetDataCount() int {
	return len(r.closes)
}
