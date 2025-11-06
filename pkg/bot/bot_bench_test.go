package bot

import (
	"encoding/json"
	"rsi-bot/pkg/models"
	"testing"
	"time"
)

// BenchmarkBot_HandleMessage benchmarks the message handling logic
func BenchmarkBot_HandleMessage(b *testing.B) {
	config := &models.Config{
		Symbol:          "BTCUSDT",
		RSIPeriod:       14,
		OverboughtLevel: 70.0,
		OversoldLevel:   30.0,
		Quantity:        0.001,
		TradingEnabled:  false,
		APIKey:          "test_key",
		APISecret:       "test_secret",
	}

	bot := New(config)
	if bot.client == nil {
		b.Skip("Bot not initialized - missing credentials")
	}

	// Create a realistic kline event message
	klineEvent := models.KlineEvent{
		EventType: "kline",
		EventTime: time.Now().UnixMilli(),
		Symbol:    "BTCUSDT",
	}
	klineEvent.Kline.Symbol = "BTCUSDT"
	klineEvent.Kline.OpenTime = time.Now().UnixMilli()
	klineEvent.Kline.Close = "45050.00"
	klineEvent.Kline.IsClosed = true

	message, _ := json.Marshal(klineEvent)

	// Warm up the indicator
	for i := 0; i < 20; i++ {
		bot.handleMessage(message)
		klineEvent.Kline.Close = string(rune(45000 + i*10))
		message, _ = json.Marshal(klineEvent)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bot.handleMessage(message)
	}
}

// BenchmarkBot_ProcessSignal benchmarks signal processing
func BenchmarkBot_ProcessSignal(b *testing.B) {
	config := &models.Config{
		Symbol:          "BTCUSDT",
		RSIPeriod:       14,
		OverboughtLevel: 70.0,
		OversoldLevel:   30.0,
		Quantity:        0.001,
		TradingEnabled:  false,
		APIKey:          "test_key",
		APISecret:       "test_secret",
	}

	bot := New(config)
	if bot.client == nil {
		b.Skip("Bot not initialized - missing credentials")
	}

	indicatorValues := map[string]float64{
		"rsi": 50.0,
	}
	currentPrice := 45000.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bot.processSignal(indicatorValues, currentPrice)
	}
}

// BenchmarkBot_JSONUnmarshal benchmarks JSON parsing of kline events
func BenchmarkBot_JSONUnmarshal(b *testing.B) {
	klineEvent := models.KlineEvent{
		EventType: "kline",
		EventTime: time.Now().UnixMilli(),
		Symbol:    "BTCUSDT",
	}
	klineEvent.Kline.Symbol = "BTCUSDT"
	klineEvent.Kline.OpenTime = time.Now().UnixMilli()
	klineEvent.Kline.Close = "45050.00"
	klineEvent.Kline.IsClosed = true

	message, _ := json.Marshal(klineEvent)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var event models.KlineEvent
		json.Unmarshal(message, &event)
	}
}

// BenchmarkBot_PositionTracking benchmarks position updates
func BenchmarkBot_PositionTracking(b *testing.B) {
	position := &models.Position{
		InPosition: false,
		Quantity:   0,
		EntryPrice: 0,
		LastUpdate: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate alternating buy/sell
		if i%2 == 0 {
			position.InPosition = true
			position.Quantity = 0.001
			position.EntryPrice = 45000.0
			position.LastUpdate = time.Now()
		} else {
			position.InPosition = false
			position.Quantity = 0
			position.EntryPrice = 0
			position.LastUpdate = time.Now()
		}
	}
}

// BenchmarkBot_New benchmarks bot creation
func BenchmarkBot_New(b *testing.B) {
	config := &models.Config{
		Symbol:          "BTCUSDT",
		RSIPeriod:       14,
		OverboughtLevel: 70.0,
		OversoldLevel:   30.0,
		Quantity:        0.001,
		TradingEnabled:  false,
		APIKey:          "test_key",
		APISecret:       "test_secret",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(config)
	}
}

// BenchmarkBot_EventEmission benchmarks event callback system
func BenchmarkBot_EventEmission(b *testing.B) {
	config := &models.Config{
		Symbol:          "BTCUSDT",
		RSIPeriod:       14,
		OverboughtLevel: 70.0,
		OversoldLevel:   30.0,
		Quantity:        0.001,
		TradingEnabled:  false,
		APIKey:          "test_key",
		APISecret:       "test_secret",
	}

	bot := New(config)
	if bot.client == nil {
		b.Skip("Bot not initialized - missing credentials")
	}

	// Set up a no-op callback
	callbackCount := 0
	bot.SetEventCallback(func(eventType string, message string, data map[string]interface{}) {
		callbackCount++
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bot.emit("test:event", "Test message", map[string]interface{}{
			"index": i,
		})
	}
}

// BenchmarkBot_MemoryAllocation benchmarks overall memory usage
func BenchmarkBot_MemoryAllocation(b *testing.B) {
	config := &models.Config{
		Symbol:          "BTCUSDT",
		RSIPeriod:       14,
		OverboughtLevel: 70.0,
		OversoldLevel:   30.0,
		Quantity:        0.001,
		TradingEnabled:  false,
		APIKey:          "test_key",
		APISecret:       "test_secret",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bot := New(config)
		_ = bot
	}
}

// BenchmarkBot_FullCycle benchmarks a complete trading cycle
func BenchmarkBot_FullCycle(b *testing.B) {
	config := &models.Config{
		Symbol:          "BTCUSDT",
		RSIPeriod:       14,
		OverboughtLevel: 70.0,
		OversoldLevel:   30.0,
		Quantity:        0.001,
		TradingEnabled:  false,
		APIKey:          "test_key",
		APISecret:       "test_secret",
	}

	bot := New(config)
	if bot.client == nil {
		b.Skip("Bot not initialized - missing credentials")
	}

	basePrice := 45000.0
	timestamp := time.Now()

	// Create messages for buy and sell scenarios
	createMessage := func(price float64, closeTime int64) []byte {
		klineEvent := models.KlineEvent{
			EventType: "kline",
			EventTime: closeTime,
			Symbol:    "BTCUSDT",
		}
		klineEvent.Kline.Symbol = "BTCUSDT"
		klineEvent.Kline.OpenTime = closeTime - 60000
		klineEvent.Kline.Close = string(rune(int(price)))
		klineEvent.Kline.IsClosed = true
		msg, _ := json.Marshal(klineEvent)
		return msg
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate price movements
		price := basePrice + float64(i%100)*10
		closeTime := timestamp.Add(time.Duration(i) * time.Minute).UnixMilli()
		message := createMessage(price, closeTime)
		bot.handleMessage(message)
	}
}

// BenchmarkBot_DatabaseOperations benchmarks database interactions
func BenchmarkBot_DatabaseOperations(b *testing.B) {
	config := &models.Config{
		Symbol:          "BTCUSDT",
		RSIPeriod:       14,
		OverboughtLevel: 70.0,
		OversoldLevel:   30.0,
		Quantity:        0.001,
		TradingEnabled:  false,
		APIKey:          "test_key",
		APISecret:       "test_secret",
	}

	bot := New(config)
	if bot.client == nil {
		b.Skip("Bot not initialized - missing credentials")
	}
	defer bot.CloseDatabase()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bot.GetRecentTrades(10)
		bot.GetOpenPosition()
	}
}
