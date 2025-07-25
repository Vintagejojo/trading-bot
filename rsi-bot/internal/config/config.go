package config

import (
	"log"
	"path/filepath"

	"rsi-bot/internal/models"

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
