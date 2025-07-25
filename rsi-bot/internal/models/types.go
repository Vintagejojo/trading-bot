package models

import "time"

type Config struct {
	Symbol          string  `mapstructure:"symbol"`
	RSIPeriod       int     `mapstructure:"rsi_period"`
	OverboughtLevel float64 `mapstructure:"overbought_level"`
	OversoldLevel   float64 `mapstructure:"oversold_level"`
	Quantity        float64 `mapstructure:"quantity"`
	TradingEnabled  bool    `mapstructure:"trading_enabled"`
}

type Position struct {
	InPosition bool
	Quantity   float64
	EntryPrice float64
	LastUpdate time.Time
}

type KlineEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	Kline     struct {
		Symbol   string `json:"s"`
		OpenTime int64  `json:"t"`
		Close    string `json:"c"`
		IsClosed bool   `json:"x"`
	} `json:"k"`
}
