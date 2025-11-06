package strategy

import (
	"fmt"
	"strings"

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

	// Create the indicator first
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

	case "multitimeframe", "multi_timeframe":
		// For multi-timeframe strategy, ignore the indicator parameter
		// as it creates its own indicators for each timeframe
		strategyConfig := DefaultMultiTimeframeStrategyConfig()
		if config.OverboughtLevel != 0 {
			strategyConfig.RSIOverbought = config.OverboughtLevel
		}
		if config.OversoldLevel != 0 {
			strategyConfig.RSIOversold = config.OversoldLevel
		}
		return NewMultiTimeframeStrategy(strategyConfig)

	default:
		return nil, fmt.Errorf("unknown strategy type: %s", config.Type)
	}
}

// ValidateConfig checks if a strategy configuration is valid
func (f *Factory) ValidateConfig(config StrategyConfig) error {
	if config.Type == "" {
		return fmt.Errorf("strategy type cannot be empty")
	}

	// Validate indicator config
	if err := f.indicatorFactory.ValidateConfig(config.IndicatorConfig); err != nil {
		return fmt.Errorf("invalid indicator config: %w", err)
	}

	strategyType := strings.ToLower(config.Type)

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
	default:
		return fmt.Errorf("unknown strategy type: %s", config.Type)
	}

	return nil
}

// GetAvailableStrategies returns a list of all available strategy types
func (f *Factory) GetAvailableStrategies() []string {
	return []string{
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
