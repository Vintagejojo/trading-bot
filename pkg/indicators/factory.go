package indicators

import (
	"fmt"
	"strings"
)

// IndicatorConfig represents configuration for creating an indicator
type IndicatorConfig struct {
	Type   string                 // "rsi", "macd", "bbands", "stoch_rsi"
	Params map[string]interface{} // Indicator-specific parameters
}

// Factory creates indicators based on configuration
type Factory struct{}

// NewFactory creates a new indicator factory
func NewFactory() *Factory {
	return &Factory{}
}

// Create builds an indicator based on the provided configuration
func (f *Factory) Create(config IndicatorConfig) (Indicator, error) {
	indicatorType := strings.ToLower(config.Type)

	switch indicatorType {
	case "rsi":
		return f.createRSI(config.Params)
	case "macd":
		return f.createMACD(config.Params)
	case "bbands", "bollinger_bands":
		return f.createBollingerBands(config.Params)
	case "stoch_rsi", "stochastic_rsi":
		return nil, fmt.Errorf("Stochastic RSI indicator not yet implemented (coming in Phase 4)")
	default:
		return nil, fmt.Errorf("unknown indicator type: %s", config.Type)
	}
}

// createRSI creates an RSI indicator from parameters
func (f *Factory) createRSI(params map[string]interface{}) (Indicator, error) {
	// Default period
	period := 14

	// Check if period is specified in params
	if p, ok := params["period"]; ok {
		switch v := p.(type) {
		case int:
			period = v
		case float64:
			period = int(v)
		case string:
			// Try to parse as int
			var parsed int
			if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
				period = parsed
			}
		}
	}

	return NewRSI(period)
}

// createMACD creates a MACD indicator from parameters
func (f *Factory) createMACD(params map[string]interface{}) (Indicator, error) {
	// Default parameters
	fastPeriod := 12
	slowPeriod := 26
	signalPeriod := 9

	// Parse parameters
	if p, ok := params["fast_period"]; ok {
		switch v := p.(type) {
		case int:
			fastPeriod = v
		case float64:
			fastPeriod = int(v)
		}
	}

	if p, ok := params["slow_period"]; ok {
		switch v := p.(type) {
		case int:
			slowPeriod = v
		case float64:
			slowPeriod = int(v)
		}
	}

	if p, ok := params["signal_period"]; ok {
		switch v := p.(type) {
		case int:
			signalPeriod = v
		case float64:
			signalPeriod = int(v)
		}
	}

	return NewMACD(fastPeriod, slowPeriod, signalPeriod)
}

// createBollingerBands creates a Bollinger Bands indicator from parameters
func (f *Factory) createBollingerBands(params map[string]interface{}) (Indicator, error) {
	// Default parameters
	period := 20
	stdDev := 2.0

	// Parse parameters
	if p, ok := params["period"]; ok {
		switch v := p.(type) {
		case int:
			period = v
		case float64:
			period = int(v)
		}
	}

	if p, ok := params["std_dev"]; ok {
		switch v := p.(type) {
		case float64:
			stdDev = v
		case int:
			stdDev = float64(v)
		}
	}

	return NewBollingerBands(period, stdDev)
}

// GetAvailableIndicators returns a list of all available indicator types
func (f *Factory) GetAvailableIndicators() []string {
	return []string{
		"rsi",           // Available
		"macd",          // Available
		"bbands",        // Available
		"stoch_rsi",     // Coming in future release
	}
}

// ValidateConfig checks if an indicator configuration is valid
func (f *Factory) ValidateConfig(config IndicatorConfig) error {
	if config.Type == "" {
		return fmt.Errorf("indicator type cannot be empty")
	}

	indicatorType := strings.ToLower(config.Type)

	switch indicatorType {
	case "rsi":
		return f.validateRSIConfig(config.Params)
	case "macd":
		return f.validateMACDConfig(config.Params)
	case "bbands", "bollinger_bands":
		return f.validateBollingerBandsConfig(config.Params)
	case "stoch_rsi", "stochastic_rsi":
		return fmt.Errorf("Stochastic RSI not yet implemented")
	default:
		return fmt.Errorf("unknown indicator type: %s (available: %v)",
			config.Type, f.GetAvailableIndicators())
	}
}

// validateRSIConfig validates RSI-specific parameters
func (f *Factory) validateRSIConfig(params map[string]interface{}) error {
	if params == nil {
		return nil // Use defaults
	}

	if p, ok := params["period"]; ok {
		var period int
		switch v := p.(type) {
		case int:
			period = v
		case float64:
			period = int(v)
		default:
			return fmt.Errorf("RSI period must be a number, got %T", p)
		}

		if period < 2 {
			return fmt.Errorf("RSI period must be at least 2, got %d", period)
		}
		if period > 100 {
			return fmt.Errorf("RSI period too large: %d (max 100)", period)
		}
	}

	return nil
}

// validateMACDConfig validates MACD-specific parameters
func (f *Factory) validateMACDConfig(params map[string]interface{}) error {
	if params == nil {
		return nil // Use defaults
	}

	var fastPeriod, slowPeriod, signalPeriod int

	// Validate fast_period
	if p, ok := params["fast_period"]; ok {
		switch v := p.(type) {
		case int:
			fastPeriod = v
		case float64:
			fastPeriod = int(v)
		default:
			return fmt.Errorf("MACD fast_period must be a number, got %T", p)
		}
		if fastPeriod < 2 {
			return fmt.Errorf("MACD fast_period must be at least 2, got %d", fastPeriod)
		}
	} else {
		fastPeriod = 12 // default
	}

	// Validate slow_period
	if p, ok := params["slow_period"]; ok {
		switch v := p.(type) {
		case int:
			slowPeriod = v
		case float64:
			slowPeriod = int(v)
		default:
			return fmt.Errorf("MACD slow_period must be a number, got %T", p)
		}
		if slowPeriod < 2 {
			return fmt.Errorf("MACD slow_period must be at least 2, got %d", slowPeriod)
		}
	} else {
		slowPeriod = 26 // default
	}

	// Validate signal_period
	if p, ok := params["signal_period"]; ok {
		switch v := p.(type) {
		case int:
			signalPeriod = v
		case float64:
			signalPeriod = int(v)
		default:
			return fmt.Errorf("MACD signal_period must be a number, got %T", p)
		}
		if signalPeriod < 2 {
			return fmt.Errorf("MACD signal_period must be at least 2, got %d", signalPeriod)
		}
	}

	// Validate fast < slow
	if fastPeriod >= slowPeriod {
		return fmt.Errorf("MACD fast_period (%d) must be less than slow_period (%d)", fastPeriod, slowPeriod)
	}

	return nil
}

// validateBollingerBandsConfig validates Bollinger Bands-specific parameters
func (f *Factory) validateBollingerBandsConfig(params map[string]interface{}) error {
	if params == nil {
		return nil // Use defaults
	}

	// Validate period
	if p, ok := params["period"]; ok {
		var period int
		switch v := p.(type) {
		case int:
			period = v
		case float64:
			period = int(v)
		default:
			return fmt.Errorf("Bollinger Bands period must be a number, got %T", p)
		}
		if period < 2 {
			return fmt.Errorf("Bollinger Bands period must be at least 2, got %d", period)
		}
	}

	// Validate std_dev
	if p, ok := params["std_dev"]; ok {
		var stdDev float64
		switch v := p.(type) {
		case float64:
			stdDev = v
		case int:
			stdDev = float64(v)
		default:
			return fmt.Errorf("Bollinger Bands std_dev must be a number, got %T", p)
		}
		if stdDev <= 0 {
			return fmt.Errorf("Bollinger Bands std_dev must be positive, got %.2f", stdDev)
		}
	}

	return nil
}

// GetDefaultConfig returns default configuration for an indicator type
func (f *Factory) GetDefaultConfig(indicatorType string) IndicatorConfig {
	indicatorType = strings.ToLower(indicatorType)

	switch indicatorType {
	case "rsi":
		return IndicatorConfig{
			Type: "rsi",
			Params: map[string]interface{}{
				"period": 14,
			},
		}
	case "macd":
		return IndicatorConfig{
			Type: "macd",
			Params: map[string]interface{}{
				"fast_period":   12,
				"slow_period":   26,
				"signal_period": 9,
			},
		}
	case "bbands", "bollinger_bands":
		return IndicatorConfig{
			Type: "bbands",
			Params: map[string]interface{}{
				"period": 20,
				"std_dev": 2.0,
			},
		}
	case "stoch_rsi", "stochastic_rsi":
		return IndicatorConfig{
			Type: "stoch_rsi",
			Params: map[string]interface{}{
				"rsi_period": 14,
				"stoch_period": 14,
				"k_period": 3,
				"d_period": 3,
			},
		}
	default:
		return IndicatorConfig{
			Type:   indicatorType,
			Params: map[string]interface{}{},
		}
	}
}
