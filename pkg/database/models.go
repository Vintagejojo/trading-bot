package database

import "time"

// Trade represents a completed trade (buy or sell)
type Trade struct {
	ID              int64     `json:"id"`
	Symbol          string    `json:"symbol"`
	Side            string    `json:"side"` // "BUY" or "SELL"
	Quantity        float64   `json:"quantity"`
	Price           float64   `json:"price"`
	Total           float64   `json:"total"` // quantity * price
	Strategy        string    `json:"strategy"` // "RSI", "MACD", "BBands"
	IndicatorValues string    `json:"indicator_values"` // JSON string of indicator values at time of trade
	SignalReason    string    `json:"signal_reason"` // Human-readable reason for the trade
	PaperTrade      bool      `json:"paper_trade"` // true if trading_enabled was false
	Timestamp       time.Time `json:"timestamp"`
	BinanceOrderID  string    `json:"binance_order_id,omitempty"` // Only populated for real trades

	// Profit/Loss tracking (only for SELL trades)
	ProfitLoss        float64 `json:"profit_loss,omitempty"` // Absolute profit/loss
	ProfitLossPercent float64 `json:"profit_loss_percent,omitempty"` // Percentage
	RelatedBuyID      int64   `json:"related_buy_id,omitempty"` // Links SELL to its BUY
}

// Position represents the current or historical position
type Position struct {
	ID         int64     `json:"id"`
	Symbol     string    `json:"symbol"`
	Quantity   float64   `json:"quantity"`
	EntryPrice float64   `json:"entry_price"`
	EntryTime  time.Time `json:"entry_time"`
	ExitPrice  float64   `json:"exit_price,omitempty"`
	ExitTime   *time.Time `json:"exit_time,omitempty"` // NULL if position is still open
	Strategy   string    `json:"strategy"`
	IsOpen     bool      `json:"is_open"`

	// Profit/Loss (calculated when position closes)
	ProfitLoss        float64 `json:"profit_loss,omitempty"`
	ProfitLossPercent float64 `json:"profit_loss_percent,omitempty"`

	// Trade references
	BuyTradeID  int64 `json:"buy_trade_id"`
	SellTradeID int64 `json:"sell_trade_id,omitempty"`
}

// TradeSummary provides aggregate statistics
type TradeSummary struct {
	TotalTrades       int       `json:"total_trades"`
	TotalBuys         int       `json:"total_buys"`
	TotalSells        int       `json:"total_sells"`
	TotalProfitLoss   float64   `json:"total_profit_loss"`
	WinRate           float64   `json:"win_rate"` // Percentage of profitable trades
	AverageProfitLoss float64   `json:"average_profit_loss"`
	LargestWin        float64   `json:"largest_win"`
	LargestLoss       float64   `json:"largest_loss"`
	StartDate         time.Time `json:"start_date"`
	EndDate           time.Time `json:"end_date"`
}
