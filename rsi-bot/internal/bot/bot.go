package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"rsi-bot/internal/calculator"
	"rsi-bot/internal/models"
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
	config        *models.Config
	rsiCalculator *calculator.RSI
	position      *models.Position
	conn          *websocket.Conn
	client        *binance.Client
}

func New(config *models.Config) *Bot {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load API credentials from environment
	config.APIKey = os.Getenv("BINANCE_API_KEY")
	config.APISecret = os.Getenv("BINANCE_API_SECRET")

	// Validate API credentials
	if config.APIKey == "" || config.APISecret == "" {
		log.Fatal("BINANCE_API_KEY and BINANCE_API_SECRET must be set in .env file or environment variables")
	}
	//creating binance client below
	client := binance.NewClient(config.APIKey, config.APISecret)
	client.BaseURL = "https://testnet.binance.vision"

	return &Bot{
		config:        config,
		rsiCalculator: calculator.NewRSI(config.RSIPeriod),
		position: &models.Position{
			InPosition: false,
			Quantity:   0,
			EntryPrice: 0,
			LastUpdate: time.Now(),
		},
		client: client,
	}
}

func (b *Bot) Start(ctx context.Context) error {
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

	// Add price to RSI calculator
	b.rsiCalculator.AddPrice(closePrice)

	log.Printf("📊 Candle closed: %s = %.8f", b.config.Symbol, closePrice)

	// Calculate RSI if we have enough data
	rsi, hasEnoughData := b.rsiCalculator.Calculate()

	if !hasEnoughData {
		needed := b.config.RSIPeriod + 1 - b.rsiCalculator.GetDataCount()
		log.Printf("⏳ Need %d more candles for RSI calculation", needed)
		return nil
	}

	log.Printf("📈 RSI: %.2f (Overbought: %.1f, Oversold: %.1f)",
		rsi, b.config.OverboughtLevel, b.config.OversoldLevel)

	// Generate trading signals
	b.processRSISignal(rsi, closePrice)

	return nil
}

func (b *Bot) processRSISignal(rsi, currentPrice float64) {
	now := time.Now()

	// Check for SELL signal (overbought and we have a position)
	if rsi >= b.config.OverboughtLevel && b.position.InPosition {
		profitPercent := ((currentPrice - b.position.EntryPrice) / b.position.EntryPrice) * 100

		log.Printf("🔴 SELL SIGNAL: RSI %.2f >= %.1f (OVERBOUGHT)", rsi, b.config.OverboughtLevel)
		log.Printf("   📍 Position: %.0f @ %.8f", b.position.Quantity, b.position.EntryPrice)
		log.Printf("   💰 Current: %.8f (%.2f%% profit)", currentPrice, profitPercent)

		if b.config.TradingEnabled {
			log.Println("   🚨 WOULD EXECUTE SELL ORDER")
			// TODO: Add actual Binance API call here using b.config.APIKey and b.config.APISecret
			b.executeSellOrder(currentPrice)

			// Update position
			b.position.InPosition = false
			b.position.Quantity = 0
			b.position.EntryPrice = 0
			b.position.LastUpdate = now
		} else {
			log.Println("   📝 PAPER TRADE: Trading disabled")
		}

		// Check for BUY signal (oversold and we don't have a position)
	} else if rsi <= b.config.OversoldLevel && !b.position.InPosition {
		log.Printf("🟢 BUY SIGNAL: RSI %.2f <= %.1f (OVERSOLD)", rsi, b.config.OversoldLevel)
		log.Printf("   💵 Quantity: %.0f @ %.8f", b.config.Quantity, currentPrice)

		if b.config.TradingEnabled {
			log.Println("   🚨 WOULD EXECUTE BUY ORDER")
			// TODO: Add actual Binance API call here using b.config.APIKey and b.config.APISecret
			b.executeBuyOrder(currentPrice)

			// Update position
			b.position.InPosition = true
			b.position.Quantity = b.config.Quantity
			b.position.EntryPrice = currentPrice
			b.position.LastUpdate = now
		} else {
			log.Println("   📝 PAPER TRADE: Trading disabled")
		}

	} else {
		// No signal
		if b.position.InPosition {
			profitPercent := ((currentPrice - b.position.EntryPrice) / b.position.EntryPrice) * 100
			log.Printf("⏳ HOLDING: RSI %.2f (%.2f%% profit)", rsi, profitPercent)
		} else {
			log.Printf("⌛ WAITING: RSI %.2f (no position)", rsi)
		}
	}
}

// TODO: buy and sell orders below need to be tested rigoursly
func (b *Bot) executeBuyOrder(price float64) error {
	log.Printf("🚀 Executing BUY order: %.0f @ %.8f", b.config.Quantity, price)

	order, err := b.client.NewCreateOrderService().
		Symbol(b.config.Symbol).
		Side(binance.SideTypeBuy).
		Type(binance.OrderTypeMarket). // Market order
		Quantity(fmt.Sprintf("%.8f", b.config.Quantity)).
		Do(context.Background())

	if err != nil {
		return fmt.Errorf("buy order failed: %w", err)
	}

	log.Printf("✅ BUY order executed: %+v", order)
	return nil
}

func (b *Bot) executeSellOrder(price float64) error {
	log.Printf("💥 Executing SELL order: %.0f @ %.8f", b.position.Quantity, price)

	order, err := b.client.NewCreateOrderService().
		Symbol(b.config.Symbol).
		Side(binance.SideTypeSell).
		Type(binance.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.8f", b.position.Quantity)).
		Do(context.Background())

	if err != nil {
		return fmt.Errorf("sell order failed: %w", err)
	}

	log.Printf("✅ SELL order executed: %+v", order)
	return nil
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
