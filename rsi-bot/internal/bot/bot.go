package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rsi-bot/internal/calculator"
	"rsi-bot/internal/models"

	"github.com/gorilla/websocket"
)

type Bot struct {
	config        *models.Config
	rsiCalculator *calculator.RSI
	position      *models.Position
	conn          *websocket.Conn
}

func New(config *models.Config) *Bot {
	return &Bot{
		config:        config,
		rsiCalculator: calculator.NewRSI(config.RSIPeriod),
		position: &models.Position{
			InPosition: false,
			Quantity:   0,
			EntryPrice: 0,
			LastUpdate: time.Now(),
		},
	}
}

func (b *Bot) Start(ctx context.Context) error {
	wsURL := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@kline_1m",
		strings.ToLower(b.config.Symbol))

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			log.Println("Attempting to connect to Binance WebSocket...")
			if err := b.connectAndRun(ctx, wsURL); err != nil {
				log.Printf("Connection failed: %v", err)
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
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
	}

	// Add proper headers
	headers := http.Header{}
	headers.Set("User-Agent", "RSI-Bot/1.0")

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
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
			// TODO: Add actual Binance API call here

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
			// TODO: Add actual Binance API call here

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
