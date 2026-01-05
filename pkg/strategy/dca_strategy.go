package strategy

import (
	"time"

	"rsi-bot/pkg/indicators"
)

// DCAStrategy implements dollar-cost averaging with optional buy-the-dip logic
type DCAStrategy struct {
	name           string
	dayOfWeek      time.Weekday // e.g., time.Monday
	hourOfDay      int          // e.g., 9 for 9am
	nextBuyTime    time.Time

	// Buy-the-dip settings
	buyTheDip      bool
	dipThreshold   float64  // e.g., 5.0 = buy on -5% days
	dipMultiplier  float64  // e.g., 1.5 = buy 1.5x normal amount
	last24hHigh    float64  // Track 24h high for dip detection
	last24hReset   time.Time
	lastDipBuy     time.Time // Prevent multiple dip buys per day
}

// NewDCAStrategy creates a new DCA strategy
func NewDCAStrategy(dayOfWeek time.Weekday, hourOfDay int) *DCAStrategy {
	s := &DCAStrategy{
		name:          "DCA",
		dayOfWeek:     dayOfWeek,
		hourOfDay:     hourOfDay,
		buyTheDip:     false,
		dipThreshold:  5.0,
		dipMultiplier: 1.5,
		last24hReset:  time.Now(),
	}
	s.nextBuyTime = s.calculateNextBuyTime(time.Now())
	return s
}

// NewDCAStrategyWithDip creates a DCA strategy with buy-the-dip enabled
func NewDCAStrategyWithDip(dayOfWeek time.Weekday, hourOfDay int, dipThreshold, dipMultiplier float64) *DCAStrategy {
	s := NewDCAStrategy(dayOfWeek, hourOfDay)
	s.buyTheDip = true
	s.dipThreshold = dipThreshold
	s.dipMultiplier = dipMultiplier
	return s
}

// Name returns the strategy name
func (s *DCAStrategy) Name() string {
	return s.name
}

// GetIndicator returns nil (DCA doesn't use indicators)
func (s *DCAStrategy) GetIndicator() indicators.Indicator {
	return nil
}

// Update tracks price for buy-the-dip logic
func (s *DCAStrategy) Update(price float64, volume float64, timestamp time.Time) error {
	// Reset 24h high every 24 hours
	if time.Since(s.last24hReset) > 24*time.Hour {
		s.last24hHigh = price
		s.last24hReset = time.Now()
	}

	// Track 24h high
	if price > s.last24hHigh {
		s.last24hHigh = price
	}

	return nil
}

// IsReady always returns true (no warmup needed)
func (s *DCAStrategy) IsReady() bool {
	return true
}

// GenerateSignal returns BUY when it's time or on dips
func (s *DCAStrategy) GenerateSignal(ctx SignalContext) Signal {
	now := time.Now()

	// Regular scheduled buy
	if now.After(s.nextBuyTime) {
		s.nextBuyTime = s.calculateNextBuyTime(now)
		return SignalBuy
	}

	// Buy-the-dip logic
	if s.buyTheDip && s.isDipDay(ctx.CurrentPrice) {
		return SignalBuy
	}

	return SignalNone
}

// GetSignalReason returns the reason for the signal
func (s *DCAStrategy) GetSignalReason() string {
	return "DCA scheduled buy"
}

// isDipDay checks if current price represents a dip worth buying
func (s *DCAStrategy) isDipDay(currentPrice float64) bool {
	if s.last24hHigh == 0 {
		return false
	}

	// Prevent multiple dip buys in same day
	if time.Since(s.lastDipBuy) < 24*time.Hour {
		return false
	}

	// Calculate percent down from 24h high
	percentDown := ((s.last24hHigh - currentPrice) / s.last24hHigh) * 100

	if percentDown >= s.dipThreshold {
		s.lastDipBuy = time.Now()
		return true
	}

	return false
}

// GetNextBuyTime returns the next scheduled buy time (for email notifications)
func (s *DCAStrategy) GetNextBuyTime() time.Time {
	return s.nextBuyTime
}

// IsDipBuyEnabled returns whether buy-the-dip is enabled
func (s *DCAStrategy) IsDipBuyEnabled() bool {
	return s.buyTheDip
}

// Reset resets the strategy state
func (s *DCAStrategy) Reset() {
	s.nextBuyTime = s.calculateNextBuyTime(time.Now())
}

// calculateNextBuyTime finds the next occurrence of the target day/hour
func (s *DCAStrategy) calculateNextBuyTime(from time.Time) time.Time {
	// Start from tomorrow to avoid buying multiple times on same day
	next := from.Add(24 * time.Hour)

	// Find next occurrence of target weekday
	for next.Weekday() != s.dayOfWeek {
		next = next.Add(24 * time.Hour)
	}

	// Set to target hour (9am, 10am, etc)
	next = time.Date(
		next.Year(), next.Month(), next.Day(),
		s.hourOfDay, 0, 0, 0,
		next.Location(),
	)

	return next
}
