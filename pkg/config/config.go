package config

import (
	"log"
	"path/filepath"

	"rsi-bot/pkg/models"

	"github.com/spf13/viper"
)

func Load(configPath string) (*models.Config, error) {
	// Get directory and filename
	dir := filepath.Dir(configPath)
	filename := filepath.Base(configPath)
	name := filename[:len(filename)-len(filepath.Ext(filename))]

	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)
	viper.AddConfigPath(".") // fallback to current directory

	// Set sensible defaults
	viper.SetDefault("symbol", "SHIBUSDT")
	viper.SetDefault("rsi_period", 14)
	viper.SetDefault("overbought_level", 70.0)
	viper.SetDefault("oversold_level", 30.0)
	viper.SetDefault("quantity", 150000.0)
	viper.SetDefault("trading_enabled", false)

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		log.Println("Config file not found, using defaults")
	}

	var config models.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

/*
Configuration Loader

This package handles loading and managing the bot's configuration from YAML files
using Viper configuration library. It provides sensible defaults and flexible
configuration file locations.

Key Features:
- Loads configuration from YAML files
- Supports multiple search paths (specified directory and current directory)
- Provides sensible default values for all configuration parameters
- Gracefully falls back to defaults if config file is not found

Configuration Parameters:
- symbol: Trading pair (default: "SHIBUSDT")
- rsi_period: Number of periods for RSI calculation (default: 14)
- overbought_level: RSI threshold for sell signals (default: 70.0)
- oversold_level: RSI threshold for buy signals (default: 30.0)
- quantity: Base order size (default: 150000.0)
- trading_enabled: Switch between live/paper trading (default: false)

Usage:
1. Call Load(configPath) with optional config file path
2. Returns populated Config struct or error
3. If config file doesn't exist, uses defaults and logs warning

File Search Order:
1. Specified config path
2. Current working directory
3. Falls back to defaults if no config found

Note: Configuration is loaded once at startup and not refreshed automatically.
*/
