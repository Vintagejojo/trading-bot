package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"rsi-bot/pkg/bot"
	"rsi-bot/pkg/database"
	"rsi-bot/pkg/indicators"
	"rsi-bot/pkg/models"
	"rsi-bot/pkg/strategy"

	"github.com/adshao/go-binance/v2"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Helper function for safe min calculation
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// App struct
type App struct {
	ctx       context.Context
	bot       *bot.Bot
	config    *models.Config
	botCtx    context.Context
	botCancel context.CancelFunc
	botRunning bool
	mu        sync.Mutex
	auth      *AuthManager
	setup     *SetupManager
}

// StrategyInfo represents strategy metadata for the frontend
type StrategyInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// BotStatus represents current bot state
type BotStatus struct {
	Running      bool                   `json:"running"`
	Strategy     string                 `json:"strategy"`
	Symbol       string                 `json:"symbol"`
	TradingMode  string                 `json:"trading_mode"` // "paper" or "live"
	Position     *database.Position     `json:"position"`
	LastTrade    *database.Trade        `json:"last_trade"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	auth := NewAuthManager()
	if err := auth.Initialize(); err != nil {
		log.Printf("Auth initialization error: %v", err)
	}

	setup := NewSetupManager()
	if err := setup.Initialize(); err != nil {
		log.Printf("Setup initialization error: %v", err)
	}

	return &App{
		botRunning: false,
		auth:       auth,
		setup:      setup,
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	log.Println("Wails app started")
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	log.Println("Wails app shutting down")
	if a.botRunning {
		a.StopBot()
	}
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	// Optional: Load default config
}

// beforeClose is called when the application is about to quit
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// GetAvailableStrategies returns list of available trading strategies
func (a *App) GetAvailableStrategies() []StrategyInfo {
	return []StrategyInfo{
		{
			Name:        "rsi",
			Description: "RSI (Relative Strength Index) - Mean reversion strategy",
		},
		{
			Name:        "macd",
			Description: "MACD (Moving Average Convergence Divergence) - Trend following",
		},
		{
			Name:        "bbands",
			Description: "Bollinger Bands - Volatility-based trading",
		},
	}
}

// GetBotStatus returns current bot status
func (a *App) GetBotStatus() BotStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	status := BotStatus{
		Running:     a.botRunning,
		TradingMode: "paper",
	}

	if a.config != nil {
		status.Symbol = a.config.Symbol
		if a.config.TradingEnabled {
			status.TradingMode = "live"
		}
		if a.config.Strategy.Type != "" {
			status.Strategy = a.config.Strategy.Type
		}
	}

	if a.bot != nil {
		// Get current position
		pos, _ := a.bot.GetOpenPosition()
		status.Position = pos

		// Get last trade
		trades, err := a.bot.GetRecentTrades(1)
		if err == nil && len(trades) > 0 {
			status.LastTrade = &trades[0]
		}
	}

	return status
}

// StartBot starts the trading bot with given configuration
func (a *App) StartBot(strategyType, symbol string, quantity float64, paperTrading bool, strategyParams map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.botRunning {
		return fmt.Errorf("bot is already running")
	}

	// Build configuration
	config := &models.Config{
		Symbol:         symbol,
		Quantity:       quantity,
		TradingEnabled: !paperTrading,
	}

	// Build strategy config
	strategyFactory := strategy.NewFactory()
	defaultConfig := strategyFactory.GetDefaultConfig(strategyType)

	// Merge user params with defaults
	if strategyParams != nil {
		for k, v := range strategyParams {
			if defaultConfig.IndicatorConfig.Params == nil {
				defaultConfig.IndicatorConfig.Params = make(map[string]interface{})
			}
			defaultConfig.IndicatorConfig.Params[k] = v
		}
	}

	config.Strategy = models.StrategyConfig{
		Type:            defaultConfig.Type,
		OverboughtLevel: defaultConfig.OverboughtLevel,
		OversoldLevel:   defaultConfig.OversoldLevel,
		Indicator: models.IndicatorConfig{
			Type:   defaultConfig.IndicatorConfig.Type,
			Params: defaultConfig.IndicatorConfig.Params,
		},
	}

	// Copy overbought/oversold from params if provided
	if val, ok := strategyParams["overbought_level"].(float64); ok {
		config.Strategy.OverboughtLevel = val
	}
	if val, ok := strategyParams["oversold_level"].(float64); ok {
		config.Strategy.OversoldLevel = val
	}

	// Create bot instance
	a.config = config
	a.bot = bot.New(config)

	// Set up event callback for the bot
	a.bot.SetEventCallback(func(eventType string, message string, data map[string]interface{}) {
		runtime.EventsEmit(a.ctx, eventType, map[string]interface{}{
			"message": message,
			"data":    data,
		})
	})

	// Start bot in background
	a.botCtx, a.botCancel = context.WithCancel(context.Background())
	go func() {
		if err := a.bot.Start(a.botCtx); err != nil {
			log.Printf("Bot error: %v", err)
			runtime.EventsEmit(a.ctx, "bot:error", err.Error())
		}
	}()

	a.botRunning = true
	log.Printf("Bot started: %s strategy on %s", strategyType, symbol)
	runtime.EventsEmit(a.ctx, "bot:started", strategyType)

	return nil
}

// StopBot stops the trading bot
func (a *App) StopBot() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.botRunning {
		return fmt.Errorf("bot is not running")
	}

	if a.botCancel != nil {
		a.botCancel()
	}

	if a.bot != nil {
		a.bot.CloseDatabase()
	}

	a.botRunning = false
	log.Println("Bot stopped")
	runtime.EventsEmit(a.ctx, "bot:stopped", "")

	return nil
}

// GetTradeHistory returns recent trades
func (a *App) GetTradeHistory(limit int) ([]database.Trade, error) {
	if a.bot == nil {
		return []database.Trade{}, nil
	}
	return a.bot.GetRecentTrades(limit)
}

// GetTradesByDateRange returns trades in date range
func (a *App) GetTradesByDateRange(startStr, endStr string) ([]database.Trade, error) {
	if a.bot == nil {
		return []database.Trade{}, nil
	}

	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	return a.bot.GetTradesByDateRange(start, end)
}

// GetTradeSummary returns aggregate statistics
func (a *App) GetTradeSummary() (*database.TradeSummary, error) {
	if a.bot == nil {
		// Return empty summary if no bot
		return &database.TradeSummary{}, nil
	}
	return a.bot.GetTradeSummary()
}

// GetCurrentPosition returns the current open position
func (a *App) GetCurrentPosition() (*database.Position, error) {
	if a.bot == nil {
		return nil, nil
	}
	return a.bot.GetOpenPosition()
}

// GetDefaultStrategyParams returns default parameters for a strategy
func (a *App) GetDefaultStrategyParams(strategyType string) map[string]interface{} {
	factory := strategy.NewFactory()
	config := factory.GetDefaultConfig(strategyType)

	params := make(map[string]interface{})

	// Add indicator params
	if config.IndicatorConfig.Params != nil {
		for k, v := range config.IndicatorConfig.Params {
			params[k] = v
		}
	}

	// Add strategy-specific params
	if strategyType == "rsi" {
		params["overbought_level"] = config.OverboughtLevel
		params["oversold_level"] = config.OversoldLevel
	}

	return params
}

// ValidateConfig validates strategy configuration
func (a *App) ValidateConfig(strategyType string, params map[string]interface{}) error {
	factory := strategy.NewFactory()

	indicatorConfig := indicators.IndicatorConfig{
		Type:   strategyType,
		Params: params,
	}

	stratConfig := strategy.StrategyConfig{
		Type:            strategyType,
		IndicatorConfig: indicatorConfig,
	}

	// Add RSI levels if present
	if val, ok := params["overbought_level"].(float64); ok {
		stratConfig.OverboughtLevel = val
	}
	if val, ok := params["oversold_level"].(float64); ok {
		stratConfig.OversoldLevel = val
	}

	return factory.ValidateConfig(stratConfig)
}

// ============= Authentication Methods =============

// IsLocked returns whether app is locked
func (a *App) IsLocked() bool {
	return a.auth.IsLocked()
}

// HasPIN returns whether PIN is set
func (a *App) HasPIN() bool {
	return a.auth.HasPIN()
}

// UnlockApp unlocks the app with PIN
func (a *App) UnlockApp(pin string) error {
	return a.auth.Unlock(pin)
}

// LockApp locks the app
func (a *App) LockApp() {
	// Stop bot if running
	if a.botRunning {
		a.StopBot()
	}
	a.auth.Lock()
}

// SetPIN sets initial PIN (only when unlocked and no PIN exists)
func (a *App) SetPIN(pin string) error {
	if a.auth.HasPIN() {
		return fmt.Errorf("PIN already set, use ChangePIN instead")
	}
	return a.auth.SetPIN(pin)
}

// ChangePIN changes existing PIN
func (a *App) ChangePIN(oldPIN, newPIN string) error {
	return a.auth.ChangePIN(oldPIN, newPIN)
}

// RemovePIN removes PIN protection
func (a *App) RemovePIN() error {
	return a.auth.RemovePIN()
}

// ============= Setup/Configuration Methods =============

// IsSetupComplete returns whether initial setup is done
func (a *App) IsSetupComplete() bool {
	return a.setup.IsSetupComplete()
}

// SaveAPIKeys saves user's Binance API keys
func (a *App) SaveAPIKeys(apiKey, apiSecret string) error {
	// Validate first
	if err := a.setup.ValidateAPIKeys(apiKey, apiSecret); err != nil {
		return err
	}

	// Save to .env file
	return a.setup.SaveAPIKeys(apiKey, apiSecret)
}

// GetAPIKeyMasked returns masked API key for display
func (a *App) GetAPIKeyMasked() (string, error) {
	return a.setup.GetMaskedAPIKey()
}

// GetSetupInstructions returns help text for setup
func (a *App) GetSetupInstructions() string {
	return a.setup.GetSetupInstructions()
}

// UpdateAPIKeys updates existing API keys
func (a *App) UpdateAPIKeys(apiKey, apiSecret string) error {
	if err := a.setup.ValidateAPIKeys(apiKey, apiSecret); err != nil {
		return err
	}
	return a.setup.UpdateAPIKeys(apiKey, apiSecret)
}

// GetEnvFilePath returns path to .env file
func (a *App) GetEnvFilePath() string {
	return a.setup.GetEnvFilePath()
}

// ResetSetup clears all API credentials and returns to setup wizard
func (a *App) ResetSetup() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Stop bot if running
	if a.botRunning {
		if err := a.StopBot(); err != nil {
			log.Printf("Warning: Failed to stop bot during reset: %v", err)
		}
	}

	// Reset setup (deletes .env file)
	if err := a.setup.ResetSetup(); err != nil {
		return fmt.Errorf("failed to reset setup: %w", err)
	}

	log.Println("Setup has been reset - API credentials cleared")
	return nil
}

// ============= Wallet Balance Methods =============

// WalletBalance represents a single asset balance
type WalletBalance struct {
	Asset     string  `json:"asset"`
	Free      string  `json:"free"`
	Locked    string  `json:"locked"`
	USDValue  float64 `json:"usd_value"`  // USD value of this asset
}

// GetWalletBalance returns user's Binance wallet balances
func (a *App) GetWalletBalance() ([]WalletBalance, error) {
	// Get API keys from environment/setup
	apiKey, apiSecret, err := a.setup.LoadAPIKeys()
	if err != nil {
		return nil, fmt.Errorf("API keys not configured: %w", err)
	}

	// Trim whitespace from keys (common issue)
	apiKey = strings.TrimSpace(apiKey)
	apiSecret = strings.TrimSpace(apiSecret)

	if apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("API keys not configured")
	}

	log.Printf("API Key length: %d, Secret length: %d", len(apiKey), len(apiSecret))
	log.Printf("API Key first 8 chars: %s...", apiKey[:min(8, len(apiKey))])

	// Create Binance client for wallet balance
	client := binance.NewClient(apiKey, apiSecret)

	// Use configured API endpoint
	apiEndpoint := a.setup.GetAPIEndpoint()
	client.BaseURL = apiEndpoint
	log.Printf("Using Binance API endpoint: %s", apiEndpoint)

	// Enable debug mode to see the actual request
	client.Debug = true

	// Get server time to sync, but use a simpler approach
	// The TimeOffset field doesn't seem to work reliably, so we'll wait to ensure our time is behind
	log.Printf("GetWalletBalance: Starting to fetch balance...")
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to get server time: %v", err)
	} else {
		localTime := time.Now().UnixMilli()
		timeOffset := serverTime - localTime
		log.Printf("Time sync: Server=%d, Local=%d, Offset=%d ms", serverTime, localTime, timeOffset)

		// If our clock is ahead of server time, wait until we're behind
		if timeOffset < 0 {
			// Our clock is ahead by abs(timeOffset) ms, wait a bit longer
			sleepDuration := time.Duration(-timeOffset+1000) * time.Millisecond
			log.Printf("Clock is %d ms ahead, waiting %v to get behind server time", -timeOffset, sleepDuration)
			time.Sleep(sleepDuration)
			log.Printf("Sleep complete, proceeding with account request")
		}
	}

	// Now make the account request - our timestamp should be behind server time
	log.Printf("Calling Binance US GetAccountService...")
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		log.Printf("ERROR: GetAccountService failed: %v", err)
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}
	log.Printf("SUCCESS: Got account info with %d balances", len(account.Balances))

	// Get current prices for all trading pairs
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to get prices: %v", err)
		// Continue without prices - will show 0 USD values
	}

	// Build price map for quick lookup (symbol -> USD price)
	priceMap := make(map[string]float64)
	if prices != nil {
		for _, price := range prices {
			if priceFloat, err := strconv.ParseFloat(price.Price, 64); err == nil {
				priceMap[price.Symbol] = priceFloat
			}
		}
	}

	// Convert to our format with USD values
	balances := make([]WalletBalance, 0, len(account.Balances))
	for _, balance := range account.Balances {
		freeAmount, _ := strconv.ParseFloat(balance.Free, 64)
		lockedAmount, _ := strconv.ParseFloat(balance.Locked, 64)
		totalAmount := freeAmount + lockedAmount

		usdValue := 0.0
		asset := balance.Asset

		// Calculate USD value based on asset type
		if asset == "USD" || asset == "USDT" || asset == "BUSD" || asset == "USDC" || asset == "TUSD" {
			// Stablecoins are 1:1 with USD
			usdValue = totalAmount
		} else if totalAmount > 0 {
			// For other assets, look up price in USDT or USD pairs
			symbol := asset + "USDT"
			if price, ok := priceMap[symbol]; ok {
				usdValue = totalAmount * price
			} else {
				symbol = asset + "USD"
				if price, ok := priceMap[symbol]; ok {
					usdValue = totalAmount * price
				}
			}
		}

		balances = append(balances, WalletBalance{
			Asset:    balance.Asset,
			Free:     balance.Free,
			Locked:   balance.Locked,
			USDValue: usdValue,
		})
	}

	return balances, nil
}
