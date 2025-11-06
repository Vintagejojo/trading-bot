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
		{
			Name:        "multitimeframe",
			Description: "Multi-Timeframe - Advanced strategy using Daily/1h/5m timeframes with RSI, MACD, and Bollinger Bands",
		},
	}
}

// GetBotStatus returns current bot status
func (a *App) GetBotStatus() BotStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	log.Printf("[GetBotStatus] botRunning=%v, bot=%v", a.botRunning, a.bot != nil)

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

	log.Println("StopBot called")

	// If bot is not running, ensure clean state and return success
	// This handles the case where bot crashed or was already stopped
	if !a.botRunning {
		log.Println("Bot already stopped, ensuring clean state")

		// Cleanup any lingering resources
		if a.botCancel != nil {
			a.botCancel()
			a.botCancel = nil
		}

		if a.bot != nil {
			a.bot.Stop()          // Close WebSocket
			a.bot.CloseDatabase() // Close database
			a.bot = nil
		}

		// Emit stopped event to sync UI
		runtime.EventsEmit(a.ctx, "bot:stopped", "")

		return nil // Return success instead of error
	}

	// Cancel the context to signal goroutines to stop
	if a.botCancel != nil {
		log.Println("Cancelling bot context...")
		a.botCancel()
		a.botCancel = nil
	}

	// Stop bot and close connections
	if a.bot != nil {
		log.Println("Stopping bot and closing connections...")
		a.bot.Stop()          // Close WebSocket immediately
		a.bot.CloseDatabase() // Close database
		a.bot = nil
	}

	a.botRunning = false
	log.Println("✅ Bot stopped successfully")
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

	// Synchronize with Binance server time to avoid timestamp errors
	log.Printf("GetWalletBalance: Synchronizing time with Binance server...")

	// Get server time first
	serverTime, err := client.NewServerTimeService().Do(context.Background())
	if err != nil {
		log.Printf("Warning: Failed to get server time: %v", err)
		// Continue anyway with a default offset
		client.TimeOffset = -2000 // Default to 2 seconds behind
	} else {
		localTime := time.Now().UnixMilli()
		timeOffset := serverTime - localTime
		log.Printf("Time sync: Server=%d, Local=%d, Offset=%d ms", serverTime, localTime, timeOffset)

		// The TimeOffset should be: (server_time - local_time)
		// But we want to be BEHIND server time, so we subtract additional buffer
		// If our clock is ahead (offset is negative), we need to go back even more
		// If our clock is behind (offset is positive), we still want to be a bit more behind for safety

		// Set offset to ensure we're always 2 seconds behind server time
		client.TimeOffset = timeOffset - 2000
		log.Printf("Setting TimeOffset to %d ms (will make requests appear 2s behind server)", client.TimeOffset)
	}

	// Small delay to ensure we're definitely behind server time
	time.Sleep(100 * time.Millisecond)

	// Now make the account request with synchronized time
	log.Printf("Calling Binance GetAccountService with TimeOffset=%d...", client.TimeOffset)
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

// ============= Multi-Timeframe Chart Data Methods =============

// CandleData represents a single candlestick
type CandleData struct {
	Timestamp int64   `json:"timestamp"` // Unix milliseconds
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

// IndicatorData represents indicator values for a timeframe
type IndicatorData struct {
	Timestamp int64   `json:"timestamp"`
	RSI       float64 `json:"rsi"`
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
	BBUpper   float64 `json:"bb_upper"`
	BBMiddle  float64 `json:"bb_middle"`
	BBLower   float64 `json:"bb_lower"`
}

// TimeframeChartData represents chart data for a specific timeframe
type TimeframeChartData struct {
	Timeframe  string          `json:"timeframe"`
	Candles    []CandleData    `json:"candles"`
	Indicators IndicatorData   `json:"indicators"`
	IsReady    bool            `json:"is_ready"`
}

// GetMultiTimeframeData returns chart data for all timeframes
func (a *App) GetMultiTimeframeData() (map[string]TimeframeChartData, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.bot == nil || !a.botRunning {
		return nil, fmt.Errorf("bot is not running")
	}

	// Get multi-timeframe manager from bot
	mtfManager := a.bot.GetMultiTimeframeManager()
	if mtfManager == nil {
		return nil, fmt.Errorf("multi-timeframe manager not available")
	}

	result := make(map[string]TimeframeChartData)

	// Get data for each timeframe
	timeframes := []strategy.Timeframe{
		strategy.Timeframe5m,
		strategy.Timeframe1h,
		strategy.Timeframe1d,
	}

	for _, tf := range timeframes {
		tfData := mtfManager.TimeframeData[tf]
		tfIndicators := mtfManager.Indicators[tf]

		if tfData == nil || tfIndicators == nil {
			continue
		}

		// Convert candles to CandleData
		candles := make([]CandleData, 0, len(tfData.Candles))
		for _, candle := range tfData.Candles {
			candles = append(candles, CandleData{
				Timestamp: candle.Timestamp.UnixMilli(),
				Open:      candle.Open,
				High:      candle.High,
				Low:       candle.Low,
				Close:     candle.Close,
				Volume:    candle.Volume,
			})
		}

		// Get current indicator values
		snapshot, isReady := mtfManager.GetIndicatorValues(tf)

		indicatorData := IndicatorData{
			Timestamp: time.Now().UnixMilli(),
		}

		if isReady {
			indicatorData.RSI = snapshot.RSI
			indicatorData.MACD = snapshot.MACD
			indicatorData.Signal = snapshot.MACDSignal
			indicatorData.Histogram = snapshot.MACDHistogram
			indicatorData.BBUpper = snapshot.BBandsUpper
			indicatorData.BBMiddle = snapshot.BBandsMiddle
			indicatorData.BBLower = snapshot.BBandsLower
		}

		result[tf.String()] = TimeframeChartData{
			Timeframe:  tf.String(),
			Candles:    candles,
			Indicators: indicatorData,
			IsReady:    isReady,
		}
	}

	return result, nil
}

// GetTimeframeData returns chart data for a specific timeframe
func (a *App) GetTimeframeData(timeframe string) (*TimeframeChartData, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.bot == nil || !a.botRunning {
		return nil, fmt.Errorf("bot is not running")
	}

	mtfManager := a.bot.GetMultiTimeframeManager()
	if mtfManager == nil {
		return nil, fmt.Errorf("multi-timeframe manager not available")
	}

	// Convert string to Timeframe
	tf := strategy.Timeframe(timeframe)

	tfData := mtfManager.TimeframeData[tf]
	tfIndicators := mtfManager.Indicators[tf]

	if tfData == nil || tfIndicators == nil {
		return nil, fmt.Errorf("timeframe %s not available", timeframe)
	}

	// Convert candles
	candles := make([]CandleData, 0, len(tfData.Candles))
	for _, candle := range tfData.Candles {
		candles = append(candles, CandleData{
			Timestamp: candle.Timestamp.UnixMilli(),
			Open:      candle.Open,
			High:      candle.High,
			Low:       candle.Low,
			Close:     candle.Close,
			Volume:    candle.Volume,
		})
	}

	// Get indicator values
	snapshot, isReady := mtfManager.GetIndicatorValues(tf)

	indicatorData := IndicatorData{
		Timestamp: time.Now().UnixMilli(),
	}

	if isReady {
		indicatorData.RSI = snapshot.RSI
		indicatorData.MACD = snapshot.MACD
		indicatorData.Signal = snapshot.MACDSignal
		indicatorData.Histogram = snapshot.MACDHistogram
		indicatorData.BBUpper = snapshot.BBandsUpper
		indicatorData.BBMiddle = snapshot.BBandsMiddle
		indicatorData.BBLower = snapshot.BBandsLower
	}

	return &TimeframeChartData{
		Timeframe:  tf.String(),
		Candles:    candles,
		Indicators: indicatorData,
		IsReady:    isReady,
	}, nil
}
