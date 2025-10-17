package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"rsi-bot/pkg/database"
	"rsi-bot/pkg/indicators"
	"rsi-bot/pkg/models"
	"rsi-bot/pkg/strategy"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Helper function for safe string slicing
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type Bot struct {
	config   *models.Config
	strategy strategy.Strategy
	position *models.Position
	conn     *websocket.Conn
	client   *binance.Client
	db       *database.DB
	logs     []string

	// Track current position in database
	currentPositionID int64

	// Event callback for real-time updates to UI
	eventCallback func(eventType string, message string, data map[string]interface{})
}

func New(config *models.Config) *Bot {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load API credentials from environment
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")

	// Use credentials from config if available, otherwise from environment
	if config.APIKey == "" {
		config.APIKey = apiKey
	}
	if config.APISecret == "" {
		config.APISecret = apiSecret
	}

	// Validate API credentials - return nil instead of crashing
	if config.APIKey == "" || config.APISecret == "" {
		log.Println("⚠️  BINANCE_API_KEY and BINANCE_API_SECRET must be set")
		// Return a minimal bot that will fail gracefully when started
		return &Bot{
			config:   config,
			position: &models.Position{},
		}
	}
	//creating binance client below
	client := binance.NewClient(config.APIKey, config.APISecret)
	client.BaseURL = "https://testnet.binance.vision"

	// Create strategy based on config
	var strat strategy.Strategy
	var err error
	stratFactory := strategy.NewFactory()

	// Check if new strategy config is specified
	if config.Strategy.Type != "" {
		// Use new strategy config
		stratConfig := strategy.StrategyConfig{
			Type:            config.Strategy.Type,
			IndicatorConfig: indicators.IndicatorConfig{
				Type:   config.Strategy.Indicator.Type,
				Params: config.Strategy.Indicator.Params,
			},
			OverboughtLevel: config.Strategy.OverboughtLevel,
			OversoldLevel:   config.Strategy.OversoldLevel,
		}

		// Validate config
		if err := stratFactory.ValidateConfig(stratConfig); err != nil {
			log.Fatalf("Invalid strategy configuration: %v", err)
		}

		strat, err = stratFactory.Create(stratConfig)
		if err != nil {
			log.Fatalf("Failed to create strategy: %v", err)
		}

		log.Printf("✅ Created %s strategy with indicator: %s", config.Strategy.Type, config.Strategy.Indicator.Type)
	} else if config.Indicator.Type != "" {
		// Fallback to legacy indicator config (create RSI strategy)
		log.Println("⚠️  Using legacy 'indicator' config. Consider using 'strategy' config instead.")
		stratConfig := strategy.StrategyConfig{
			Type: config.Indicator.Type,
			IndicatorConfig: indicators.IndicatorConfig{
				Type:   config.Indicator.Type,
				Params: config.Indicator.Params,
			},
			OverboughtLevel: config.OverboughtLevel,
			OversoldLevel:   config.OversoldLevel,
		}

		strat, err = stratFactory.Create(stratConfig)
		if err != nil {
			log.Fatalf("Failed to create strategy from indicator config: %v", err)
		}

		log.Printf("✅ Created %s strategy from legacy config", config.Indicator.Type)
	} else {
		// Fallback to legacy RSI config for backward compatibility
		log.Println("⚠️  Using legacy RSI configuration (rsi_period). Consider using 'strategy' config instead.")
		stratConfig := strategy.StrategyConfig{
			Type: "rsi",
			IndicatorConfig: indicators.IndicatorConfig{
				Type: "rsi",
				Params: map[string]interface{}{
					"period": config.RSIPeriod,
				},
			},
			OverboughtLevel: config.OverboughtLevel,
			OversoldLevel:   config.OversoldLevel,
		}

		strat, err = stratFactory.Create(stratConfig)
		if err != nil {
			log.Fatalf("Failed to create RSI strategy: %v", err)
		}

		log.Printf("✅ Created RSI strategy with period: %d", config.RSIPeriod)
	}

	// Initialize database
	db, err := database.New("trading_bot.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("✅ Database initialized")

	// Check for existing open position in database
	dbPosition, err := db.GetOpenPosition(config.Symbol)
	if err != nil {
		log.Printf("⚠️  Error checking for open position: %v", err)
	}

	position := &models.Position{
		InPosition: false,
		Quantity:   0,
		EntryPrice: 0,
		LastUpdate: time.Now(),
	}

	var currentPosID int64 = 0

	// Restore position from database if exists
	if dbPosition != nil {
		position.InPosition = true
		position.Quantity = dbPosition.Quantity
		position.EntryPrice = dbPosition.EntryPrice
		position.LastUpdate = dbPosition.EntryTime
		currentPosID = dbPosition.ID
		log.Printf("📍 Restored open position from database: %.0f @ %.8f", position.Quantity, position.EntryPrice)
	}

	return &Bot{
		config:            config,
		strategy:          strat,
		position:          position,
		client:            client,
		db:                db,
		currentPositionID: currentPosID,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	// Check if bot was initialized properly
	if b.client == nil {
		return fmt.Errorf("bot not properly initialized: missing API credentials")
	}

	// Safely log API key (first 8 chars only if long enough)
	if len(b.config.APIKey) >= 16 {
		log.Printf("🔑 API Key loaded: %s...%s",
			b.config.APIKey[:8],
			b.config.APIKey[len(b.config.APIKey)-8:])
	} else {
		log.Printf("🔑 API Key loaded: %s...", b.config.APIKey[:min(8, len(b.config.APIKey))])
	}

	// Try multiple WebSocket endpoints
	wsURLs := []string{
		fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@kline_1m", strings.ToLower(b.config.Symbol)),
		fmt.Sprintf("wss://stream.binance.com/ws/%s@kline_1m", strings.ToLower(b.config.Symbol)),
		fmt.Sprintf("wss://data-stream.binance.vision/ws/%s@kline_1m", strings.ToLower(b.config.Symbol)),
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			log.Println("Attempting to connect to Binance WebSocket...")

			var lastErr error
			for _, wsURL := range wsURLs {
				if err := b.connectAndRun(ctx, wsURL); err != nil {
					log.Printf("Failed to connect to %s: %v", wsURL, err)
					lastErr = err
					continue
				}
			}

			if lastErr != nil {
				log.Printf("All connection attempts failed. Last error: %v", lastErr)
				log.Println("Retrying in 5 seconds...")

				select {
				case <-ctx.Done():
					return nil
				case <-time.After(5 * time.Second):
					continue
				}
			}
		}
	}
}

func (b *Bot) connectAndRun(ctx context.Context, wsURL string) error {
	// Create dialer with timeout and proper headers
	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
	}

	// Add proper headers for Binance
	headers := http.Header{}
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	headers.Set("Origin", "https://www.binance.com")

	log.Printf("🔗 Connecting to: %s", wsURL)
	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		if resp != nil {
			log.Printf("❌ HTTP Response Status: %s", resp.Status)
			log.Printf("❌ Response Headers: %v", resp.Header)
		}
		return fmt.Errorf("websocket dial failed: %w", err)
	}
	defer conn.Close()

	// Set connection timeouts
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	b.conn = conn
	log.Printf("✅ Connected to %s", wsURL)
	b.emit("bot:connected", fmt.Sprintf("Connected to %s", wsURL), map[string]interface{}{
		"url": wsURL,
	})

	// Start ping routine to keep connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Start ping routine in goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Printf("Ping failed: %v", err)
					return
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, closing connection")
			return nil
		default:
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					return fmt.Errorf("websocket unexpected close: %w", err)
				}
				return fmt.Errorf("websocket read error: %w", err)
			}

			if err := b.handleMessage(message); err != nil {
				log.Printf("Error handling message: %v", err)
				// Continue processing other messages
			}
		}
	}
}

func (b *Bot) handleMessage(message []byte) error {
	var event models.KlineEvent
	if err := json.Unmarshal(message, &event); err != nil {
		return fmt.Errorf("failed to unmarshal kline event: %w", err)
	}

	// Only process closed candles
	if !event.Kline.IsClosed {
		return nil
	}

	closePrice, err := strconv.ParseFloat(event.Kline.Close, 64)
	if err != nil {
		return fmt.Errorf("failed to parse close price: %w", err)
	}

	// Update indicator with new price
	indicator := b.strategy.GetIndicator()
	if err := indicator.Update(closePrice, time.Unix(event.Kline.OpenTime/1000, 0)); err != nil {
		return fmt.Errorf("failed to update indicator: %w", err)
	}

	log.Printf("📊 Candle closed: %s = %.8f", b.config.Symbol, closePrice)
	b.emit("bot:candle", fmt.Sprintf("Candle closed: %s = %.8f", b.config.Symbol, closePrice), map[string]interface{}{
		"symbol": b.config.Symbol,
		"price":  closePrice,
	})

	// Check if indicator has enough data
	if !indicator.IsReady() {
		log.Printf("⏳ Waiting for indicator to initialize (current data: %d)", indicator.GetDataCount())
		b.emit("bot:status", fmt.Sprintf("Waiting for indicator (%d data points)", indicator.GetDataCount()), map[string]interface{}{
			"dataPoints": indicator.GetDataCount(),
		})
		return nil
	}

	// Get indicator values
	values, isValid := indicator.GetValue()
	if !isValid {
		log.Printf("⚠️ Indicator not ready yet")
		return nil
	}

	// Log indicator values
	log.Printf("📈 %s: %v", b.strategy.Name(), values)
	b.emit("bot:indicator", fmt.Sprintf("%s: %v", b.strategy.Name(), values), map[string]interface{}{
		"strategy": b.strategy.Name(),
		"values":   values,
	})

	// Generate trading signal using strategy
	b.processSignal(values, closePrice)

	return nil
}

func (b *Bot) processSignal(indicatorValues map[string]float64, currentPrice float64) {
	now := time.Now()

	// Create signal context
	ctx := strategy.SignalContext{
		CurrentPrice:  currentPrice,
		Position:      b.position,
		IndicatorData: indicatorValues,
	}

	// Generate signal from strategy
	signal := b.strategy.GenerateSignal(ctx)
	reason := b.strategy.GetSignalReason()

	// Process signal
	switch signal {
	case strategy.SignalBuy:
		log.Printf("🟢 BUY SIGNAL: %s", reason)
		log.Printf("   💵 Quantity: %.0f @ %.8f", b.config.Quantity, currentPrice)
		b.emit("bot:trade", fmt.Sprintf("BUY Signal: %s", reason), map[string]interface{}{
			"side":     "BUY",
			"price":    currentPrice,
			"quantity": b.config.Quantity,
			"reason":   reason,
		})

		var binanceOrderID string
		if b.config.TradingEnabled {
			log.Println("   🚨 EXECUTING BUY ORDER")
			orderID, err := b.executeBuyOrder(currentPrice)
			if err != nil {
				log.Printf("   ❌ BUY ORDER FAILED: %v", err)
				return
			}
			binanceOrderID = orderID
			log.Println("   ✅ Order executed")
		} else {
			log.Println("   📝 PAPER TRADE: Trading disabled")
		}

		// Log trade to database
		trade := &database.Trade{
			Symbol:          b.config.Symbol,
			Side:            "BUY",
			Quantity:        b.config.Quantity,
			Price:           currentPrice,
			Total:           b.config.Quantity * currentPrice,
			Strategy:        b.strategy.Name(),
			IndicatorValues: database.SerializeIndicatorValues(indicatorValues),
			SignalReason:    reason,
			PaperTrade:      !b.config.TradingEnabled,
			Timestamp:       now,
			BinanceOrderID:  binanceOrderID,
		}

		tradeID, err := b.db.InsertTrade(trade)
		if err != nil {
			log.Printf("   ⚠️  Failed to log trade to database: %v", err)
		} else {
			log.Printf("   💾 Trade logged (ID: %d)", tradeID)

			// Create new position in database
			dbPos := &database.Position{
				Symbol:      b.config.Symbol,
				Quantity:    b.config.Quantity,
				EntryPrice:  currentPrice,
				EntryTime:   now,
				Strategy:    b.strategy.Name(),
				IsOpen:      true,
				BuyTradeID:  tradeID,
			}

			posID, err := b.db.InsertPosition(dbPos)
			if err != nil {
				log.Printf("   ⚠️  Failed to log position to database: %v", err)
			} else {
				b.currentPositionID = posID
				log.Printf("   💾 Position logged (ID: %d)", posID)
			}
		}

		// Update in-memory position
		b.position.InPosition = true
		b.position.Quantity = b.config.Quantity
		b.position.EntryPrice = currentPrice
		b.position.LastUpdate = now

	case strategy.SignalSell:
		profitLoss := (currentPrice - b.position.EntryPrice) * b.position.Quantity
		profitPercent := ((currentPrice - b.position.EntryPrice) / b.position.EntryPrice) * 100
		log.Printf("🔴 SELL SIGNAL: %s", reason)
		log.Printf("   📍 Position: %.0f @ %.8f", b.position.Quantity, b.position.EntryPrice)
		log.Printf("   💰 Current: %.8f (%.2f%% profit, $%.2f)", currentPrice, profitPercent, profitLoss)
		b.emit("bot:trade", fmt.Sprintf("SELL Signal: %s", reason), map[string]interface{}{
			"side":          "SELL",
			"price":         currentPrice,
			"quantity":      b.position.Quantity,
			"reason":        reason,
			"profitLoss":    profitLoss,
			"profitPercent": profitPercent,
		})

		var binanceOrderID string
		if b.config.TradingEnabled {
			log.Println("   🚨 EXECUTING SELL ORDER")
			orderID, err := b.executeSellOrder(currentPrice)
			if err != nil {
				log.Printf("   ❌ SELL ORDER FAILED: %v", err)
				return
			}
			binanceOrderID = orderID
			log.Println("   ✅ Order executed")
		} else {
			log.Println("   📝 PAPER TRADE: Trading disabled")
		}

		// Log trade to database
		trade := &database.Trade{
			Symbol:            b.config.Symbol,
			Side:              "SELL",
			Quantity:          b.position.Quantity,
			Price:             currentPrice,
			Total:             b.position.Quantity * currentPrice,
			Strategy:          b.strategy.Name(),
			IndicatorValues:   database.SerializeIndicatorValues(indicatorValues),
			SignalReason:      reason,
			PaperTrade:        !b.config.TradingEnabled,
			Timestamp:         now,
			BinanceOrderID:    binanceOrderID,
			ProfitLoss:        profitLoss,
			ProfitLossPercent: profitPercent,
		}

		tradeID, err := b.db.InsertTrade(trade)
		if err != nil {
			log.Printf("   ⚠️  Failed to log trade to database: %v", err)
		} else {
			log.Printf("   💾 Trade logged (ID: %d)", tradeID)

			// Update position in database
			if b.currentPositionID > 0 {
				err := b.db.UpdatePosition(
					b.currentPositionID,
					currentPrice,
					now,
					profitLoss,
					profitPercent,
					tradeID,
				)
				if err != nil {
					log.Printf("   ⚠️  Failed to update position in database: %v", err)
				} else {
					log.Printf("   💾 Position closed (ID: %d)", b.currentPositionID)
				}
			}
		}

		// Update in-memory position
		b.position.InPosition = false
		b.position.Quantity = 0
		b.position.EntryPrice = 0
		b.position.LastUpdate = now
		b.currentPositionID = 0

	default:
		// No signal - just log status
		log.Printf("⌛ %s", reason)
	}
}

// TODO: buy and sell orders below need to be tested rigoursly
func (b *Bot) executeBuyOrder(price float64) (string, error) {
	log.Printf("🚀 Executing BUY order: %.0f @ %.8f", b.config.Quantity, price)

	order, err := b.client.NewCreateOrderService().
		Symbol(b.config.Symbol).
		Side(binance.SideTypeBuy).
		Type(binance.OrderTypeMarket). // Market order
		Quantity(fmt.Sprintf("%.8f", b.config.Quantity)).
		Do(context.Background())

	if err != nil {
		return "", fmt.Errorf("buy order failed: %w", err)
	}

	orderID := fmt.Sprintf("%d", order.OrderID)
	log.Printf("✅ BUY order executed: OrderID=%s", orderID)
	return orderID, nil
}

func (b *Bot) executeSellOrder(price float64) (string, error) {
	log.Printf("💥 Executing SELL order: %.0f @ %.8f", b.position.Quantity, price)

	order, err := b.client.NewCreateOrderService().
		Symbol(b.config.Symbol).
		Side(binance.SideTypeSell).
		Type(binance.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.8f", b.position.Quantity)).
		Do(context.Background())

	if err != nil {
		return "", fmt.Errorf("sell order failed: %w", err)
	}

	orderID := fmt.Sprintf("%d", order.OrderID)
	log.Printf("✅ SELL order executed: OrderID=%s", orderID)
	return orderID, nil
}

// GetRecentTrades returns the most recent trades from the database
func (b *Bot) GetRecentTrades(limit int) ([]database.Trade, error) {
	if b.db == nil {
		return []database.Trade{}, nil
	}
	return b.db.GetRecentTrades(limit)
}

// GetTradesByDateRange returns trades within a specific date range
func (b *Bot) GetTradesByDateRange(start, end time.Time) ([]database.Trade, error) {
	if b.db == nil {
		return []database.Trade{}, nil
	}
	return b.db.GetTradesByDateRange(start, end)
}

// GetTradeSummary returns aggregate trading statistics
func (b *Bot) GetTradeSummary() (*database.TradeSummary, error) {
	if b.db == nil {
		return &database.TradeSummary{}, nil
	}
	return b.db.GetTradeSummary()
}

// GetOpenPosition returns the current open position from database
func (b *Bot) GetOpenPosition() (*database.Position, error) {
	if b.db == nil {
		return nil, nil
	}
	return b.db.GetOpenPosition(b.config.Symbol)
}

// CloseDatabase closes the database connection (call on shutdown)
func (b *Bot) CloseDatabase() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// SetEventCallback sets a callback function for real-time UI updates
func (b *Bot) SetEventCallback(callback func(eventType string, message string, data map[string]interface{})) {
	b.eventCallback = callback
}

// emit sends an event to the UI if callback is set
func (b *Bot) emit(eventType string, message string, data map[string]interface{}) {
	if b.eventCallback != nil {
		b.eventCallback(eventType, message, data)
	}
}

// This bot implements a simple RSI-based trading strategy for Binance cryptocurrency exchange.
// It connects to Binance's WebSocket API to receive real-time price data and executes trades
// based on RSI (Relative Strength Index) signals.

// Main Components:
// 1. Bot Structure:
//    - Manages WebSocket connection to Binance
//    - Trades based on RSI signals
//    - Maintains position state (entry price, quantity, etc.)
//    - Handles configuration (symbol, RSI period, thresholds)

// 2. Key Functionality:
//    - Connects to Binance WebSocket for 1-minute kline/candlestick data
//    - Calculates RSI in real-time as new candles close
//    - Generates BUY signals when RSI falls below oversold level
//    - Generates SELL signals when RSI rises above overbought level
//    - Supports both live trading and paper trading modes

// 3. Trading Logic:
//    - BUY when RSI <= Oversold level (no existing position)
//    - SELL when RSI >= Overbought level (with existing position)
//    - Position tracking with profit/loss calculation
//    - Configurable RSI period (default 14) and threshold levels

// 4. Technical Implementation:
//    - Uses Gorilla WebSocket for persistent connection
//    - Automatic reconnection on failure
//    - Connection keep-alive with ping/pong
//    - Proper timeout handling
//    - Thread-safe position management

// Configuration Options:
//    - Symbol: Trading pair (e.g., "BTCUSDT")
//    - RSIPeriod: Number of periods for RSI calculation
//    - OverboughtLevel: RSI threshold for sell signals
//    - OversoldLevel: RSI threshold for buy signals
//    - Quantity: Base order size
//    - TradingEnabled: Switch between live/paper trading

// Usage:
// 1. Create config with desired parameters
// 2. Initialize bot with New(config)
// 3. Start bot with Start(ctx)

// Note: The bot currently logs trade actions rather than executing them when TradingEnabled is false.
