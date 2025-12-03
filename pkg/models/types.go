package models

import (
	"time"
	"rsi-bot/pkg/safety"
)

type Config struct {
	Symbol          string  `mapstructure:"symbol"`
	RSIPeriod       int     `mapstructure:"rsi_period"` // Deprecated: use Strategy config instead
	OverboughtLevel float64 `mapstructure:"overbought_level"` // Deprecated: use Strategy config instead
	OversoldLevel   float64 `mapstructure:"oversold_level"` // Deprecated: use Strategy config instead
	Quantity        float64 `mapstructure:"quantity"`
	TradingEnabled  bool    `mapstructure:"trading_enabled"`
	APIKey          string
	APISecret       string

	// New: Strategy configuration (includes indicator)
	Strategy StrategyConfig `mapstructure:"strategy"`

	// Deprecated: Use Strategy config instead
	Indicator IndicatorConfig `mapstructure:"indicator"`

	// Safety & Resilience (Phase 7.5)
	Safety safety.Config `mapstructure:"safety"`
}

// StrategyConfig defines which strategy to use
type StrategyConfig struct {
	Type            string                 `mapstructure:"type"`   // "rsi", "macd", "bbands"
	OverboughtLevel float64                `mapstructure:"overbought_level"` // For RSI strategy
	OversoldLevel   float64                `mapstructure:"oversold_level"`   // For RSI strategy
	Indicator       IndicatorConfig        `mapstructure:"indicator"` // Indicator configuration
}

// IndicatorConfig defines which indicator to use and its parameters
type IndicatorConfig struct {
	Type   string                 `mapstructure:"type"`   // "rsi", "macd", "bbands", etc.
	Params map[string]interface{} `mapstructure:"params"` // Indicator-specific parameters
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
		Volume   string `json:"v"`
		IsClosed bool   `json:"x"`
	} `json:"k"`
}
