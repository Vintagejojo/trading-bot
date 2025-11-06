# Quick Start: Multi-Timeframe Strategy

## Installation

```bash
# Navigate to project directory
cd trading-bot

# Install dependencies
go mod tidy

# Run the example
go run examples/multitimeframe_example.go
```

## Basic Setup (Copy-Paste Ready)

```go
package main

import (
    "fmt"
    "log"
    "time"
    "rsi-bot/pkg/strategy"
    "rsi-bot/pkg/models"
)

func main() {
    // 1. Create multi-timeframe strategy
    strategyConfig := strategy.DefaultMultiTimeframeStrategyConfig()
    mts, err := strategy.NewMultiTimeframeStrategy(strategyConfig)
    if err != nil {
        log.Fatal(err)
    }

    // 2. Create risk manager
    riskManager := strategy.NewRiskManager(strategy.DefaultRiskConfig())

    // 3. Create market condition analyzer
    mcAnalyzer := strategy.NewMarketConditionAnalyzer(
        strategy.DefaultMarketConditionConfig(),
    )

    // 4. Initialize position tracking
    position := &models.Position{InPosition: false}
    portfolioValue := 10000.0

    // 5. Process price updates (from WebSocket in production)
    price := 0.00001000
    volume := 50000000.0
    timestamp := time.Now()

    // Update strategy with new data
    err = mts.Update(price, volume, timestamp)
    if err != nil {
        log.Printf("Update error: %v", err)
        return
    }

    // 6. Generate signal
    signalContext := strategy.SignalContext{
        CurrentPrice: price,
        Position:     position,
    }
    signal := mts.GenerateSignal(signalContext)

    // 7. Execute trades based on signal
    if signal == strategy.SignalBuy && !position.InPosition {
        // Calculate position size
        positionSize, err := riskManager.CalculatePositionSize(
            portfolioValue,
            price,
            0, // volatility (0 for fixed stop-loss)
        )
        if err != nil {
            log.Printf("Position sizing error: %v", err)
            return
        }

        // Execute buy
        fmt.Printf("BUY: %.0f @ %.8f\n", positionSize.Quantity, price)
        position.InPosition = true
        position.EntryPrice = price
        position.Quantity = positionSize.Quantity
    }
}
```

## Configuration Cheat Sheet

### Strategy Configuration

```go
config := strategy.DefaultMultiTimeframeStrategyConfig()

// RSI thresholds
config.RSIOversold = 30.0      // Buy when RSI ≤ 30
config.RSIOverbought = 70.0    // Sell when RSI ≥ 70

// Timeframe requirements
config.RequireDailyTrendConfirmation = true  // Enforce daily trend filter
config.RequireHourlySignal = true            // Require 1h signal
config.Require5MinuteEntry = true            // Require 5m confirmation

// Volatility (Bollinger Bands)
config.BBandsMinWidth = 1.0    // Minimum 1% volatility
config.BBandsMaxWidth = 10.0   // Maximum 10% volatility
```

### Risk Management Configuration

```go
riskConfig := strategy.DefaultRiskConfig()

// Position sizing
riskConfig.MaxPositionSizePercent = 10.0  // Max 10% of portfolio
riskConfig.RiskPerTradePercent = 2.0      // Risk 2% per trade

// Stop-loss
riskConfig.StopLossPercent = 3.0          // 3% fixed stop-loss
riskConfig.UseATRStopLoss = false         // Or use ATR-based (dynamic)
riskConfig.ATRMultiplier = 2.0            // ATR multiplier if enabled

// Take-profit
riskConfig.TakeProfitPercent = 6.0        // Fixed 6% profit target
riskConfig.UseRiskRewardRatio = true      // Or use risk/reward ratio
riskConfig.RiskRewardRatio = 2.0          // 2:1 reward/risk

// Trailing stop
riskConfig.UseTrailingStop = true         // Enable trailing stop
riskConfig.TrailingStopPercent = 4.0      // Activate at 4% profit
riskConfig.TrailingStopDistance = 2.0     // Trail 2% below peak
```

### Market Condition Configuration

```go
mcConfig := strategy.DefaultMarketConditionConfig()

// Volatility thresholds
mcConfig.MinVolatilityPercent = 1.0   // 1% minimum
mcConfig.MaxVolatilityPercent = 10.0  // 10% maximum

// Volume filter
mcConfig.UseVolumeFilter = true
mcConfig.MinVolumeMultiplier = 0.5    // 50% of average volume
mcConfig.VolumeAveragePeriod = 20     // 20-period average

// Spread filter (liquidity)
mcConfig.MaxSpreadPercent = 0.5       // 0.5% max bid-ask spread
```

## Signal Generation Flow

```
1. Update Strategy
   └─> mts.Update(price, volume, timestamp)

2. Check if Ready
   └─> if !mts.IsReady() { continue }

3. Get Snapshots (for debugging)
   └─> snapshots := mts.GetMultiTimeframeManager().GetAllSnapshots()
   └─> hourly := snapshots[strategy.Timeframe1h]

4. Check Market Conditions
   └─> condition := mcAnalyzer.AnalyzeMarketConditions(...)
   └─> if !condition.IsTradeableMarket { skip }

5. Generate Signal
   └─> signal := mts.GenerateSignal(signalContext)
   └─> reason := mts.GetSignalReason()

6. Execute Trade
   └─> if signal == SignalBuy && !position.InPosition { ... }
```

## Common Patterns

### Pattern 1: Basic Buy/Sell Loop

```go
for {
    // Get price data (from WebSocket or API)
    price, volume, timestamp := getLatestPrice()

    // Update strategy
    mts.Update(price, volume, timestamp)

    // Generate signal
    signal := mts.GenerateSignal(strategy.SignalContext{
        CurrentPrice: price,
        Position:     &position,
    })

    // Handle signal
    switch signal {
    case strategy.SignalBuy:
        if !position.InPosition {
            executeBuy(price)
        }
    case strategy.SignalSell:
        if position.InPosition {
            executeSell(price)
        }
    }
}
```

### Pattern 2: Position Sizing with Risk Management

```go
if signal == strategy.SignalBuy && !position.InPosition {
    // Calculate position size
    positionSize, err := riskManager.CalculatePositionSize(
        portfolioValue,
        price,
        0, // or provide ATR volatility
    )
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }

    // Validate risk limits
    err = riskManager.ValidatePositionRisk(
        portfolioValue,
        positionSize.RiskAmount,
        existingPositionCount,
        existingTotalRisk,
    )
    if err != nil {
        log.Printf("Risk rejected: %v", err)
        return
    }

    // Execute order
    executeBuyOrder(positionSize.Quantity, price)

    // Set stop-loss and take-profit
    stopLoss = positionSize.StopLossPrice
    takeProfit = positionSize.TakeProfitPrice
}
```

### Pattern 3: Trailing Stop Management

```go
// Initialize trailing stop when entering position
trailingStop := strategy.NewTrailingStopTracker(
    entryPrice,
    initialStopLoss,
    activationPercent,  // e.g., 4.0
    trailingDistance,   // e.g., 2.0
)

// Update on each price tick
for price := range priceStream {
    stopTriggered := trailingStop.Update(price)

    if stopTriggered {
        executeSell(price)
        fmt.Printf("Trailing stop triggered at %.8f\n", price)
        break
    }

    // Get current stop level
    currentStop := trailingStop.GetStopLossPrice()
    fmt.Printf("Current stop: %.8f\n", currentStop)
}
```

### Pattern 4: Market Condition Filtering

```go
// Setup volume tracker
volumeTracker := strategy.NewVolumeTracker(50)

// Track volume history
volumeTracker.Add(currentVolume)

// Check market conditions before trading
condition := mcAnalyzer.AnalyzeMarketConditions(
    volatility,          // BBands width or ATR
    currentVolume,
    volumeTracker.GetHistory(),
    bidPrice,
    askPrice,
)

if !condition.IsTradeableMarket {
    log.Printf("Market not tradeable: %v", condition.Reasons)
    return
}

fmt.Printf("Market status: %s\n", condition.String())
// Proceed with trading...
```

## Debugging Tips

### 1. Log All Timeframe Snapshots

```go
snapshots := mts.GetMultiTimeframeManager().GetAllSnapshots()

for tf, snapshot := range snapshots {
    fmt.Println(snapshot.String())
}

// Output:
// [5m] Price: 0.00001015 | RSI: 45.32 | MACD: 0.0002/0.0001/0.0001 | ...
// [1h] Price: 0.00001012 | RSI: 28.54 | MACD: 0.0015/0.0012/0.0003 | ...
// [1d] Price: 0.00001020 | RSI: 62.18 | MACD: 0.0045/0.0038/0.0007 | ...
```

### 2. Log Signal Reasons

```go
signal := mts.GenerateSignal(ctx)
reason := mts.GetSignalReason()

fmt.Printf("Signal: %s\n", signal)
fmt.Printf("Reason: %s\n", reason)

// Output:
// Signal: BUY
// Reason: Daily Trend: BULLISH | 1h BUY: RSI oversold (28.45), MACD bullish crossover | 5-min entry confirmed | Volatility OK (3.21%)
```

### 3. Monitor Position Metrics

```go
if position.InPosition {
    summary := riskManager.GetPositionSummary(
        position.EntryPrice,
        currentPrice,
        position.Quantity,
        stopLoss,
        takeProfit,
    )

    fmt.Printf("Unrealized P/L: $%.2f (%.2f%%)\n",
        summary.UnrealizedPL, summary.UnrealizedPLPercent)
    fmt.Printf("Stop: %.8f | Target: %.8f\n",
        summary.StopLossPrice, summary.TakeProfitPrice)
}
```

## Common Errors and Solutions

### Error: "Insufficient data on 1-hour timeframe"

**Cause:** Not enough price data collected yet.

**Solution:** Wait for warm-up period. Strategy needs ~200 data points per timeframe.

```go
if !mts.IsReady() {
    fmt.Println("Warming up... waiting for data")
    continue
}
```

### Error: "Volatility outside acceptable range"

**Cause:** Market too quiet or too volatile.

**Solution:** Adjust volatility thresholds or wait for better conditions.

```go
config.BBandsMinWidth = 0.5  // Lower minimum if market is quiet
config.BBandsMaxWidth = 15.0 // Raise maximum if you want to trade volatile markets
```

### Error: "Trend misalignment: Daily=BEARISH, Hourly Signal=BUY"

**Cause:** Trying to buy in a daily downtrend.

**Solution:** Either disable daily filter or wait for trend alignment.

```go
config.RequireDailyTrendConfirmation = false  // Allow counter-trend trades
```

## Performance Tuning

### Reduce False Signals

```go
// Require stronger conditions
config.RSIOversold = 25.0      // More oversold
config.RSIOverbought = 75.0    // More overbought

// Stricter timeframe requirements
config.RequireDailyTrendConfirmation = true
config.Require5MinuteEntry = true
```

### Increase Trade Frequency

```go
// Relax conditions
config.RSIOversold = 35.0
config.RSIOverbought = 65.0

// Optional: Disable some filters
config.Require5MinuteEntry = false
config.BBandsMinWidth = 0.5  // Lower volatility threshold
```

### Optimize Risk/Reward

```go
// Conservative (lower risk, lower reward)
riskConfig.RiskPerTradePercent = 1.0
riskConfig.RiskRewardRatio = 3.0  // 3:1 ratio

// Aggressive (higher risk, higher reward)
riskConfig.RiskPerTradePercent = 3.0
riskConfig.RiskRewardRatio = 1.5  // 1.5:1 ratio
```

## Next Steps

1. **Read full documentation:** `MULTI_TIMEFRAME_STRATEGY.md`
2. **Run the example:** `go run examples/multitimeframe_example.go`
3. **Backtest with historical data:** Implement backtesting module
4. **Paper trade:** Test with real-time data, simulated orders
5. **Live trading:** Start with very small positions

## Quick Reference: Key Functions

| Function | Purpose |
|----------|---------|
| `mts.Update(price, volume, timestamp)` | Update all timeframes with new data |
| `mts.IsReady()` | Check if enough data collected |
| `mts.GenerateSignal(ctx)` | Generate BUY/SELL/NONE signal |
| `mts.GetSignalReason()` | Get human-readable explanation |
| `riskManager.CalculatePositionSize()` | Calculate safe position size |
| `riskManager.ShouldExit()` | Check stop-loss / take-profit |
| `mcAnalyzer.AnalyzeMarketConditions()` | Check if market is tradeable |
| `trailingStop.Update(price)` | Update trailing stop, returns true if triggered |

## Support

For questions or issues:
1. Check `MULTI_TIMEFRAME_STRATEGY.md` for detailed explanations
2. Review `examples/multitimeframe_example.go` for working code
3. See `pkg/strategy/*.go` for implementation details

**Happy Trading!** 📈
