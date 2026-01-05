package strategy

import (
	"fmt"
	"strings"
	"time"

	"rsi-bot/pkg/indicators"
)

// StrategyConfig represents configuration for creating a strategy
type StrategyConfig struct {
	Type              string                 // "rsi", "macd", "bbands"
	IndicatorConfig   indicators.IndicatorConfig
	OverboughtLevel   float64 // For RSI strategy
	OversoldLevel     float64 // For RSI strategy
}

// Factory creates trading strategies
type Factory struct {
	indicatorFactory *indicators.Factory
}

// NewFactory creates a new strategy factory
func NewFactory() *Factory {
	return &Factory{
		indicatorFactory: indicators.NewFactory(),
	}
}

// Create builds a strategy based on the provided configuration
func (f *Factory) Create(config StrategyConfig) (Strategy, error) {
	strategyType := strings.ToLower(config.Type)

	// Handle multitimeframe strategy separately (it doesn't use a single indicator)
	if strategyType == "multitimeframe" || strategyType == "multi_timeframe" {
		strategyConfig := DefaultMultiTimeframeStrategyConfig()
		if config.OverboughtLevel != 0 {
			strategyConfig.RSIOverbought = config.OverboughtLevel
		}
		if config.OversoldLevel != 0 {
			strategyConfig.RSIOversold = config.OversoldLevel
		}
		return NewMultiTimeframeStrategy(strategyConfig)
	}

	// Handle DCA strategy (doesn't use indicators)
	if strategyType == "dca" {
		// Default: Weekly on Monday at 9am
		dayOfWeek := time.Monday
		hourOfDay := 9
		buyTheDip := false
		dipThreshold := 5.0
		dipMultiplier := 1.5

		// Parse params if provided
		if params := config.IndicatorConfig.Params; params != nil {
			if dow, ok := params["day_of_week"].(float64); ok {
				dayOfWeek = time.Weekday(int(dow))
			}
			if hour, ok := params["hour_of_day"].(float64); ok {
				hourOfDay = int(hour)
			}
			if dip, ok := params["buy_the_dip"].(bool); ok {
				buyTheDip = dip
			}
			if threshold, ok := params["dip_threshold"].(float64); ok {
				dipThreshold = threshold
			}
			if multiplier, ok := params["dip_multiplier"].(float64); ok {
				dipMultiplier = multiplier
			}
		}

		if buyTheDip {
			return NewDCAStrategyWithDip(dayOfWeek, hourOfDay, dipThreshold, dipMultiplier), nil
		}
		return NewDCAStrategy(dayOfWeek, hourOfDay), nil
	}

	// Create the indicator for other strategies
	indicator, err := f.indicatorFactory.Create(config.IndicatorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create indicator: %w", err)
	}

	// Create the appropriate strategy
	switch strategyType {
	case "rsi":
		if config.OverboughtLevel == 0 {
			config.OverboughtLevel = 70.0 // default
		}
		if config.OversoldLevel == 0 {
			config.OversoldLevel = 30.0 // default
		}
		return NewRSIStrategy(indicator, config.OverboughtLevel, config.OversoldLevel)

	case "macd":
		return NewMACDStrategy(indicator)

	case "bbands", "bollinger_bands":
		return NewBollingerBandsStrategy(indicator)

	default:
		return nil, fmt.Errorf("unknown strategy type: %s", config.Type)
	}
}

// ValidateConfig checks if a strategy configuration is valid
func (f *Factory) ValidateConfig(config StrategyConfig) error {
	if config.Type == "" {
		return fmt.Errorf("strategy type cannot be empty")
	}

	strategyType := strings.ToLower(config.Type)

	// Skip indicator validation for strategies that don't use indicators
	if strategyType != "multitimeframe" && strategyType != "multi_timeframe" && strategyType != "dca" {
		// Validate indicator config for strategies that use indicators
		if err := f.indicatorFactory.ValidateConfig(config.IndicatorConfig); err != nil {
			return fmt.Errorf("invalid indicator config: %w", err)
		}
	}

	// Validate strategy-specific parameters
	switch strategyType {
	case "rsi":
		if config.OverboughtLevel != 0 && config.OversoldLevel != 0 {
			if config.OverboughtLevel <= config.OversoldLevel {
				return fmt.Errorf("overbought level (%.1f) must be greater than oversold level (%.1f)",
					config.OverboughtLevel, config.OversoldLevel)
			}
		}
	case "macd":
		// No additional validation needed
	case "bbands", "bollinger_bands":
		// No additional validation needed
	case "multitimeframe", "multi_timeframe":
		// Multi-timeframe strategy has its own validation
		if config.OverboughtLevel != 0 && config.OversoldLevel != 0 {
			if config.OverboughtLevel <= config.OversoldLevel {
				return fmt.Errorf("overbought level (%.1f) must be greater than oversold level (%.1f)",
					config.OverboughtLevel, config.OversoldLevel)
			}
		}
	case "dca":
		// DCA strategy has no specific validation requirements
	default:
		return fmt.Errorf("unknown strategy type: %s", config.Type)
	}

	return nil
}

// GetAvailableStrategies returns a list of all available strategy types
func (f *Factory) GetAvailableStrategies() []string {
	return []string{
		"dca",
		"rsi",
		"macd",
		"bbands",
		"multitimeframe",
	}
}

// GetDefaultConfig returns default configuration for a strategy type
func (f *Factory) GetDefaultConfig(strategyType string) StrategyConfig {
	strategyType = strings.ToLower(strategyType)

	switch strategyType {
	case "dca":
		return StrategyConfig{
			Type: "dca",
			IndicatorConfig: indicators.IndicatorConfig{
				Type: "dca",
				Params: map[string]interface{}{
					"day_of_week": 1,   // Monday
					"hour_of_day": 9,   // 9am
				},
			},
		}

	case "rsi":
		return StrategyConfig{
			Type: "rsi",
			IndicatorConfig: indicators.IndicatorConfig{
				Type: "rsi",
				Params: map[string]interface{}{
					"period": 14,
				},
			},
			OverboughtLevel: 70.0,
			OversoldLevel:   30.0,
		}

	case "macd":
		return StrategyConfig{
			Type: "macd",
			IndicatorConfig: indicators.IndicatorConfig{
				Type: "macd",
				Params: map[string]interface{}{
					"fast_period":   12,
					"slow_period":   26,
					"signal_period": 9,
				},
			},
		}

	case "bbands", "bollinger_bands":
		return StrategyConfig{
			Type: "bbands",
			IndicatorConfig: indicators.IndicatorConfig{
				Type: "bbands",
				Params: map[string]interface{}{
					"period":  20,
					"std_dev": 2.0,
				},
			},
		}

	case "multitimeframe", "multi_timeframe":
		return StrategyConfig{
			Type: "multitimeframe",
			IndicatorConfig: indicators.IndicatorConfig{
				Type: "multitimeframe",
				Params: map[string]interface{}{
					"rsi_period":    14,
					"macd_fast":     12,
					"macd_slow":     26,
					"macd_signal":   9,
					"bbands_period": 20,
					"bbands_stddev": 2.0,
				},
			},
			OverboughtLevel: 70.0,
			OversoldLevel:   30.0,
		}

	default:
		return StrategyConfig{
			Type:            strategyType,
			IndicatorConfig: indicators.IndicatorConfig{},
		}
	}
}
