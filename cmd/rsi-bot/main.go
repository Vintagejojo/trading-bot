// main.go - Entry point for the RSI trading bot application.
// This file coordinates startup, configuration loading, bot lifecycle management,
// and graceful shutdown in response to system signals (e.g. Ctrl+C).

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"rsi-bot/internal/bot"
	"rsi-bot/internal/models"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

func LoadConfig() (*models.Config, error) {
	viper.SetConfigName("config")  // name of config file (without extension)
	viper.SetConfigType("yaml")    // file format
	viper.AddConfigPath("configs") // path to look for the file

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg models.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &cfg, nil
}

func main() {
	log.Println("Starting RSI Trading Bot...")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create bot instance
	bot := bot.New(config)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start bot in goroutine
	go func() {
		if err := bot.Start(ctx); err != nil {
			log.Printf("Bot error: %v", err)
			cancel()
		}
	}()

	log.Println("Bot started! Press Ctrl+C to stop...")

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()

	// Give bot time to cleanup
	time.Sleep(2 * time.Second)
	log.Println("Bot stopped.")
}
