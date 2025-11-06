# Multi-Timeframe Trading Strategy Documentation

## Overview

This document provides a comprehensive guide to the multi-timeframe trading strategy implementation. The strategy uses three timeframes (Daily, 1-Hour, 5-Minute) with three technical indicators (MACD, Bollinger Bands, RSI) to generate high-probability trading signals.

---

## Table of Contents

1. [Strategy Architecture](#strategy-architecture)
2. [How Each Indicator Behaves on Each Timeframe](#how-each-indicator-behaves-on-each-timeframe)
3. [Multi-Timeframe Confirmation Logic](#multi-timeframe-confirmation-logic)
4. [Entry and Exit Conditions](#entry-and-exit-conditions)
5. [Risk Management](#risk-management)
6. [Market Condition Filters](#market-condition-filters)
7. [Configuration Guide](#configuration-guide)
8. [Debugging and Logging](#debugging-and-logging)
9. [Example Scenarios](#example-scenarios)

---

## Strategy Architecture

### Module Structure

```
pkg/strategy/
├── timeframe.go              # Timeframe data structures and OHLCV aggregation
├── multitimeframe.go         # Multi-timeframe indicator manager
├── multitimeframe_strategy.go # Main strategy logic with signal generation
├── risk_management.go        # Position sizing, stop-loss, take-profit
└── market_conditions.go      # Volatility and liquidity checks
```

### Data Flow

```
Raw Price Data (1-minute klines)
         ↓
┌────────────────────────────────────────┐
│  MultiTimeframeManager                 │
│  ┌──────────────────────────────────┐ │
│  │ 5-Minute Timeframe               │ │
│  │  • Aggregates 1m → 5m candles    │ │
│  │  • RSI, MACD, BBands indicators  │ │
│  └──────────────────────────────────┘ │
│  ┌──────────────────────────────────┐ │
│  │ 1-Hour Timeframe                 │ │
│  │  • Aggregates 1m → 1h candles    │ │
│  │  • RSI, MACD, BBands indicators  │ │
│  └──────────────────────────────────┘ │
│  ┌──────────────────────────────────┐ │
│  │ Daily Timeframe                  │ │
│  │  • Aggregates 1m → 1d candles    │ │
│  │  • RSI, MACD, BBands indicators  │ │
│  └──────────────────────────────────┘ │
└────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────┐
│  MultiTimeframeStrategy                │
│  1. Daily Trend Bias Filter            │
│  2. 1-Hour Signal Generation           │
│  3. Daily-Hourly Alignment Check       │
│  4. 5-Minute Entry Precision           │
│  5. Market Condition Validation        │
└────────────────────────────────────────┘
         ↓
┌────────────────────────────────────────┐
│  RiskManager                           │
│  • Position sizing                     │
│  • Stop-loss calculation               │
│  • Take-profit calculation             │
│  • Trailing stop management            │
└────────────────────────────────────────┘
         ↓
    BUY/SELL Signal
```

---

## How Each Indicator Behaves on Each Timeframe

### Daily Timeframe (1d) - Trend Bias Filter

**Purpose:** Identify the dominant trend to filter trades in the trend's direction.

#### RSI (Relative Strength Index)
- **Period:** 14 days
- **Bullish Signal:** RSI > 60 (price has upward momentum)
- **Bearish Signal:** RSI < 40 (price has downward momentum)
- **Neutral:** RSI between 40-60 (no clear trend)
- **Use Case:** Filter out counter-trend trades

#### MACD (Moving Average Convergence Divergence)
- **Parameters:** 12/26/9 (fast/slow/signal)
- **Bullish Signal:** MACD Histogram > 0 (bullish momentum)
- **Bearish Signal:** MACD Histogram < 0 (bearish momentum)
- **Use Case:** Confirm trend direction and momentum strength

#### Bollinger Bands
- **Period:** 20 days, 2 standard deviations
- **Bullish Signal:** Price > Middle Band (SMA)
- **Bearish Signal:** Price < Middle Band (SMA)
- **Band Width:** Indicates long-term volatility (wider = more volatile)
- **Use Case:** Determine if asset is in an uptrend or downtrend

**Daily Trend Determination Logic:**
```
BullishSignals = 0
BearishSignals = 0

IF RSI > 60:       BullishSignals++
IF RSI < 40:       BearishSignals++

IF MACD Histogram > 0:  BullishSignals++
IF MACD Histogram < 0:  BearishSignals++

IF Price > BBands Middle:  BullishSignals++
IF Price < BBands Middle:  BearishSignals++

IF BullishSignals > BearishSignals:   Trend = BULLISH
IF BearishSignals > BullishSignals:   Trend = BEARISH
ELSE:                                  Trend = NEUTRAL
```

---

### 1-Hour Timeframe (1h) - Signal Generation

**Purpose:** Generate the primary buy/sell signals based on indicator confluence.

#### RSI
- **Period:** 14 hours
- **Oversold (BUY):** RSI ≤ 30
- **Overbought (SELL):** RSI ≥ 70
- **Use Case:** Identify short-term reversal opportunities

#### MACD
- **Parameters:** 12/26/9
- **Bullish Crossover (BUY):** MACD line crosses above Signal line (Histogram > 0)
- **Bearish Crossover (SELL):** MACD line crosses below Signal line (Histogram < 0)
- **Use Case:** Confirm momentum direction

#### Bollinger Bands
- **Period:** 20 hours, 2 std dev
- **Buy Zone:** Price ≤ Lower Band (oversold)
- **Sell Zone:** Price ≥ Upper Band (overbought)
- **Band Width:** Indicates hourly volatility for trade viability
- **Use Case:** Identify price extremes

**1-Hour Signal Generation Logic:**
```
BUY CONDITIONS (need 2 out of 3):
1. RSI ≤ 30 (oversold)
2. MACD Histogram > 0 AND MACD > Signal (bullish crossover)
3. Price ≤ Lower BBand * 1.01 (within 1% of lower band)

SELL CONDITIONS (need 2 out of 3):
1. RSI ≥ 70 (overbought)
2. MACD Histogram < 0 AND MACD < Signal (bearish crossover)
3. Price ≥ Upper BBand * 0.99 (within 1% of upper band)
```

---

### 5-Minute Timeframe (5m) - Entry Precision

**Purpose:** Fine-tune entry timing to avoid premature entries.

#### RSI
- **Period:** 14 (5-minute periods)
- **For BUY confirmation:** RSI < 60 (not yet overbought)
- **For SELL confirmation:** RSI > 40 (not yet oversold)
- **Use Case:** Ensure momentum hasn't reversed

#### MACD
- **Parameters:** 12/26/9
- **For BUY confirmation:** MACD Histogram ≥ 0 (momentum turning up)
- **For SELL confirmation:** MACD Histogram ≤ 0 (momentum turning down)
- **Use Case:** Confirm short-term momentum alignment

#### Bollinger Bands
- **Period:** 20 (5-minute periods)
- **Use Case:** Monitor immediate price action volatility
- **Note:** Primarily used for volatility context, not direct signals

**5-Minute Entry Precision Logic:**
```
IF 1-Hour Signal = BUY:
  Confirm: RSI < 60 AND MACD Histogram ≥ 0

IF 1-Hour Signal = SELL:
  Confirm: RSI > 40 AND MACD Histogram ≤ 0
```

---

## Multi-Timeframe Confirmation Logic

### Signal Generation Pipeline

```
┌─────────────────────────────────────────────────────────────┐
│ PHASE 1: Daily Trend Bias Filter                            │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Analyze Daily Indicators:                            │   │
│ │  • RSI trend direction                               │   │
│ │  • MACD histogram direction                          │   │
│ │  • Price vs BBands middle                            │   │
│ │                                                      │   │
│ │ Result: BULLISH / BEARISH / NEUTRAL                  │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ IF RequireDailyTrendConfirmation = true:                    │
│   → Must have BULLISH or BEARISH trend (not NEUTRAL)        │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ PHASE 2: 1-Hour Signal Generation                           │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ BUY Signal (if NOT in position):                     │   │
│ │  • RSI ≤ 30 (oversold)               ✓ or ✗         │   │
│ │  • MACD bullish crossover            ✓ or ✗         │   │
│ │  • Price at lower BBand              ✓ or ✗         │   │
│ │                                                      │   │
│ │  Requirement: At least 2 out of 3 conditions        │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ SELL Signal (if IN position):                        │   │
│ │  • RSI ≥ 70 (overbought)             ✓ or ✗         │   │
│ │  • MACD bearish crossover            ✓ or ✗         │   │
│ │  • Price at upper BBand              ✓ or ✗         │   │
│ │                                                      │   │
│ │  Requirement: At least 2 out of 3 conditions        │   │
│ └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ PHASE 3: Daily-Hourly Alignment Check                       │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ IF Daily Trend = BULLISH:                            │   │
│ │   → ONLY allow BUY signals                           │   │
│ │                                                      │   │
│ │ IF Daily Trend = BEARISH:                            │   │
│ │   → ONLY allow SELL signals                          │   │
│ │                                                      │   │
│ │ IF Daily Trend = NEUTRAL:                            │   │
│ │   → Allow both (or reject if strict filtering)       │   │
│ └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ PHASE 4: 5-Minute Entry Precision                           │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ IF 1h Signal = BUY:                                  │   │
│ │   Confirm: 5m RSI < 60 AND 5m MACD Histogram ≥ 0    │   │
│ │                                                      │   │
│ │ IF 1h Signal = SELL:                                 │   │
│ │   Confirm: 5m RSI > 40 AND 5m MACD Histogram ≤ 0    │   │
│ └──────────────────────────────────────────────────────┘   │
│                                                              │
│ Purpose: Avoid entering just as momentum reverses           │
└─────────────────────────────────────────────────────────────┘
         ↓
┌─────────────────────────────────────────────────────────────┐
│ PHASE 5: Market Condition Validation                        │
│                                                              │
│ ┌──────────────────────────────────────────────────────┐   │
│ │ Volatility Check (using 1h BBands width):           │   │
│ │  • Minimum: 1% (enough movement for profit)          │   │
│ │  • Maximum: 10% (avoid extreme volatility)           │   │
│ │                                                      │   │
│ │ Liquidity Check (optional):                          │   │
│ │  • Volume vs 20-period average ≥ 0.5x               │   │
│ │  • Bid-ask spread < 0.5%                             │   │
│ └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
         ↓
    FINAL SIGNAL: BUY / SELL / NONE
```

---

## Entry and Exit Conditions

### Entry Conditions Summary

**Long Entry (BUY):**
1. ✅ Daily trend is BULLISH (or neutral if daily filter disabled)
2. ✅ 1-Hour: 2 out of 3 conditions met (RSI oversold, MACD bullish, Price at lower BBand)
3. ✅ 5-Minute: RSI < 60 AND MACD Histogram ≥ 0
4. ✅ Volatility: 1% ≤ BBands width ≤ 10%
5. ✅ Not currently in a position

**Short Entry (SELL) - Exit Long Position:**
1. ✅ Daily trend is BEARISH (or neutral if daily filter disabled)
2. ✅ 1-Hour: 2 out of 3 conditions met (RSI overbought, MACD bearish, Price at upper BBand)
3. ✅ 5-Minute: RSI > 40 AND MACD Histogram ≤ 0
4. ✅ Volatility: 1% ≤ BBands width ≤ 10%
5. ✅ Currently holding a position

### Exit Conditions Summary

Positions are closed when:

**Strategy-Based Exit:**
- SELL signal generated (as above) while in position

**Risk Management Exit:**
- Stop-loss triggered (price drops below stop-loss level)
- Take-profit reached (price hits profit target)
- Trailing stop triggered (price retraces after peak)

---

## Risk Management

### Position Sizing

The `RiskManager` calculates position size based on:

```go
// Key parameters
MaxPositionSizePercent: 10%   // Max 10% of portfolio per trade
RiskPerTradePercent:    2%    // Risk max 2% per trade

// Position sizing formula
RiskAmount = Portfolio * RiskPerTradePercent / 100
StopLossDistance = EntryPrice - StopLossPrice
QuantityByRisk = RiskAmount / StopLossDistance

MaxPositionValue = Portfolio * MaxPositionSizePercent / 100
QuantityByValue = MaxPositionValue / EntryPrice

FinalQuantity = MIN(QuantityByRisk, QuantityByValue)
```

**Example:**
- Portfolio: $10,000
- Entry Price: $0.00001000
- Stop-Loss: 3% below entry = $0.00000970

```
RiskAmount = $10,000 * 2% = $200
StopLossDistance = $0.00001000 - $0.00000970 = $0.00000030
QuantityByRisk = $200 / $0.00000030 = 666,666,667

MaxPositionValue = $10,000 * 10% = $1,000
QuantityByValue = $1,000 / $0.00001000 = 100,000,000

FinalQuantity = MIN(666,666,667, 100,000,000) = 100,000,000
```

### Stop-Loss Strategies

**Fixed Percentage Stop-Loss (Default):**
```go
StopLossPrice = EntryPrice * (1 - StopLossPercent/100)
// Example: Entry $0.00001000, 3% stop = $0.00000970
```

**ATR-Based Dynamic Stop-Loss:**
```go
ATR = AverageTrueRange(14 periods)
StopLossPrice = EntryPrice - (ATR * ATRMultiplier)
// Example: ATR = $0.00000015, Multiplier = 2.0
// Stop = $0.00001000 - ($0.00000015 * 2) = $0.00000970
```

### Take-Profit Strategies

**Fixed Percentage:**
```go
TakeProfitPrice = EntryPrice * (1 + TakeProfitPercent/100)
// Example: Entry $0.00001000, 6% profit = $0.00001060
```

**Risk/Reward Ratio (Recommended):**
```go
RiskAmount = EntryPrice - StopLossPrice
RewardAmount = RiskAmount * RiskRewardRatio
TakeProfitPrice = EntryPrice + RewardAmount

// Example: Risk = $0.00000030, Ratio = 2:1
// Reward = $0.00000030 * 2 = $0.00000060
// TP = $0.00001000 + $0.00000060 = $0.00001060
```

### Trailing Stop

**Configuration:**
```go
TrailingStopPercent:  4%   // Activate after 4% profit
TrailingStopDistance: 2%   // Trail 2% below peak
```

**Behavior:**
1. Entry: $0.00001000, Initial Stop: $0.00000970
2. Price rises to $0.00001040 (+4%) → Trailing stop ACTIVATES
3. New stop: $0.00001040 * 0.98 = $0.00001019 (locked in profit)
4. Price rises to $0.00001080 → Stop moves to $0.00001058
5. Price drops to $0.00001058 → Exit triggered (locked in ~5.8% profit)

---

## Market Condition Filters

### Volatility Filter (Bollinger Band Width)

**Purpose:** Ensure sufficient price movement for profitable trades, but avoid extreme volatility.

```go
BandWidth = ((UpperBand - LowerBand) / MiddleBand) * 100

MinVolatility: 1%   // Too low = choppy/ranging market
MaxVolatility: 10%  // Too high = extreme risk
```

**Trading Implications:**
- **< 1%:** Market too quiet, likely consolidating. Avoid trading.
- **1% - 5%:** Normal volatility, ideal for trading.
- **5% - 10%:** High volatility, proceed with caution.
- **> 10%:** Extreme volatility, avoid trading (news events, crashes).

### Liquidity Filter (Volume)

**Purpose:** Ensure sufficient market activity for order execution.

```go
VolumeAverage = Average(Volume, 20 periods)
CurrentVolumeRatio = CurrentVolume / VolumeAverage

MinVolumeMultiplier: 0.5   // At least 50% of average volume
```

**Trading Implications:**
- **< 0.5x:** Low liquidity, risk of slippage and poor fills. Avoid.
- **0.5x - 1.0x:** Adequate liquidity.
- **> 1.0x:** Good liquidity, favorable for trading.
- **> 2.0x:** High volume spike (potential breakout or news).

### Spread Filter (Bid-Ask Spread)

**Purpose:** Ensure tight spreads for cost-effective entry/exit.

```go
SpreadPercent = ((AskPrice - BidPrice) / MidPrice) * 100

MaxSpreadPercent: 0.5%
```

**Trading Implications:**
- **< 0.1%:** Very tight spread (ideal).
- **0.1% - 0.5%:** Acceptable spread.
- **> 0.5%:** Wide spread, poor liquidity. Avoid trading.

---

## Configuration Guide

### Example: Multi-Timeframe Strategy Configuration

```go
import "rsi-bot/pkg/strategy"

// Multi-timeframe manager config
mtfConfig := strategy.DefaultMultiTimeframeConfig()
mtfConfig.Timeframes = []strategy.Timeframe{
    strategy.Timeframe5m,  // 5-minute for entry precision
    strategy.Timeframe1h,  // 1-hour for signals
    strategy.Timeframe1d,  // Daily for trend bias
}
mtfConfig.MaxCandles = 200

// Indicator parameters (applied to all timeframes)
mtfConfig.RSIPeriod = 14
mtfConfig.MACDFast = 12
mtfConfig.MACDSlow = 26
mtfConfig.MACDSignal = 9
mtfConfig.BBandsPeriod = 20
mtfConfig.BBandsStdDev = 2.0

// Create manager
mtfManager, err := strategy.NewMultiTimeframeManager(mtfConfig)

// Strategy configuration
strategyConfig := strategy.DefaultMultiTimeframeStrategyConfig()
strategyConfig.RSIOversold = 30.0
strategyConfig.RSIOverbought = 70.0
strategyConfig.RequireDailyTrendConfirmation = true
strategyConfig.RequireHourlySignal = true
strategyConfig.Require5MinuteEntry = true

// Create strategy
mts, err := strategy.NewMultiTimeframeStrategy(strategyConfig)
```

### Example: Risk Management Configuration

```go
// Risk management config
riskConfig := strategy.DefaultRiskConfig()
riskConfig.MaxPositionSizePercent = 10.0  // Max 10% per trade
riskConfig.RiskPerTradePercent = 2.0      // Risk 2% per trade

// Stop-loss
riskConfig.StopLossPercent = 3.0          // 3% fixed stop
riskConfig.UseATRStopLoss = false         // Or use ATR-based

// Take-profit
riskConfig.UseRiskRewardRatio = true
riskConfig.RiskRewardRatio = 2.0          // 2:1 reward/risk

// Trailing stop
riskConfig.UseTrailingStop = true
riskConfig.TrailingStopPercent = 4.0      // Activate at 4% profit
riskConfig.TrailingStopDistance = 2.0     // Trail 2% below peak

// Create risk manager
riskManager := strategy.NewRiskManager(riskConfig)

// Calculate position size
positionSize, err := riskManager.CalculatePositionSize(
    portfolioValue,  // e.g., 10000.0
    entryPrice,      // e.g., 0.00001000
    volatility,      // ATR or 0 if not using ATR stop
)

fmt.Printf("Quantity: %.0f\n", positionSize.Quantity)
fmt.Printf("Stop-Loss: %.8f\n", positionSize.StopLossPrice)
fmt.Printf("Take-Profit: %.8f\n", positionSize.TakeProfitPrice)
fmt.Printf("Risk Amount: $%.2f (%.2f%%)\n",
    positionSize.RiskAmount, positionSize.MaxLossPercent)
```

### Example: Market Condition Configuration

```go
// Market condition analyzer config
mcConfig := strategy.DefaultMarketConditionConfig()
mcConfig.MinVolatilityPercent = 1.0    // 1% minimum
mcConfig.MaxVolatilityPercent = 10.0   // 10% maximum

mcConfig.UseVolumeFilter = true
mcConfig.MinVolumeMultiplier = 0.5     // 50% of average
mcConfig.VolumeAveragePeriod = 20

mcConfig.MaxSpreadPercent = 0.5        // 0.5% max spread

// Create analyzer
mcAnalyzer := strategy.NewMarketConditionAnalyzer(mcConfig)

// Analyze conditions
condition := mcAnalyzer.AnalyzeMarketConditions(
    volatility,      // e.g., BBands width %
    currentVolume,   // Current volume
    volumeHistory,   // []float64 of past volumes
    bidPrice,        // Bid price
    askPrice,        // Ask price
)

if condition.IsTradeableMarket {
    fmt.Println("Market is tradeable!")
    fmt.Println(condition.String())
} else {
    fmt.Println("Market NOT tradeable:", condition.Reasons)
}
```

---

## Debugging and Logging

### Signal Debugging

Log all timeframe snapshots before generating signals:

```go
snapshots := mts.GetMultiTimeframeManager().GetAllSnapshots()

for tf, snapshot := range snapshots {
    fmt.Println(snapshot.String())
    // Output example:
    // [1d] Price: 0.00001023 | RSI: 65.42 | MACD: 0.0012/0.0008/0.0004 |
    //      BBands: 0.00001150/0.00001020/0.00000890 (width: 2.55%)
}
```

### Signal Reason Logging

The strategy provides detailed reasoning for each signal:

```go
signal := mts.GenerateSignal(ctx)
reason := mts.GetSignalReason()

fmt.Printf("Signal: %s\n", signal)
fmt.Printf("Reason: %s\n", reason)

// Example output:
// Signal: BUY
// Reason: Daily Trend: BULLISH | 1h BUY: RSI oversold (28.45), Price at lower BB |
//         5-min entry confirmed | Volatility OK (3.21%)
```

### Position Tracking

Track position performance in real-time:

```go
if position.InPosition {
    summary := riskManager.GetPositionSummary(
        position.EntryPrice,
        currentPrice,
        position.Quantity,
        stopLossPrice,
        takeProfitPrice,
    )

    fmt.Printf("Position P/L: $%.2f (%.2f%%)\n",
        summary.UnrealizedPL, summary.UnrealizedPLPercent)
    fmt.Printf("Current Stop: %.8f | Target: %.8f\n",
        summary.StopLossPrice, summary.TakeProfitPrice)
}
```

### Recommended Logging Structure

```go
type TradeLog struct {
    Timestamp       time.Time
    Signal          string        // "BUY" / "SELL" / "NONE"
    Reason          string

    // Timeframe snapshots
    DailyRSI        float64
    DailyTrend      string
    HourlyRSI       float64
    HourlyMACD      float64
    FiveMinRSI      float64

    // Market conditions
    Volatility      float64
    VolumeRatio     float64

    // Position details (if applicable)
    EntryPrice      float64
    Quantity        float64
    StopLoss        float64
    TakeProfit      float64
}

// Log to database or file
logEntry := TradeLog{
    Timestamp:  time.Now(),
    Signal:     signal.String(),
    Reason:     mts.GetSignalReason(),
    // ... populate other fields
}
```

---

## Example Scenarios

### Scenario 1: Successful Long Trade

**Setup:**
- Symbol: SHIBUSDT
- Portfolio: $10,000
- Daily trend: BULLISH (RSI: 62, MACD: +0.0015, Price above SMA)

**Entry (1-Hour Signal):**
- Time: 10:00 AM
- Price: $0.00001000
- 1h RSI: 28 (oversold ✓)
- 1h MACD: Histogram turning positive, crossover ✓
- 1h BBands: Price at $0.00000998, lower band at $0.00000995 ✓
- **2 out of 3 conditions met → BUY signal**

**5-Minute Confirmation:**
- 5m RSI: 45 (< 60 ✓)
- 5m MACD Histogram: +0.00005 (≥ 0 ✓)
- **Entry confirmed**

**Volatility Check:**
- BBands width: 3.2% (within 1%-10% ✓)

**Risk Calculation:**
- Entry: $0.00001000
- Stop-Loss (3%): $0.00000970
- Take-Profit (2:1 ratio): $0.00001060
- Position size: 100,000,000 SHIB
- Risk: $200 (2% of portfolio)
- Potential profit: $400

**Trade Execution:**
- BUY 100,000,000 SHIB @ $0.00001000

**Exit:**
- Time: 2:30 PM (4.5 hours later)
- Peak price: $0.00001075
- Trailing stop activated at $0.00001040 (4% profit)
- Trailing stop triggered at $0.00001053 (2% trail from peak)
- **SELL 100,000,000 SHIB @ $0.00001053**
- **Profit: $353 (+3.5%)**

---

### Scenario 2: Stopped Out Trade

**Setup:**
- Daily trend: NEUTRAL
- 1h signals align for BUY

**Entry:**
- Price: $0.00001000
- Stop-Loss: $0.00000970 (3%)

**What Happened:**
- Price drops to $0.00000975
- Then reverses to $0.00000985
- But stop-loss already triggered at $0.00000970

**Exit:**
- SELL @ $0.00000970
- **Loss: $200 (-3.0%, exactly as planned)**

**Lesson:** Stop-loss protected the account from larger drawdown. The 2% risk rule limited the loss.

---

### Scenario 3: No Trade (Volatility Too High)

**Setup:**
- Daily trend: BULLISH
- 1h RSI: 25 (oversold)
- 1h MACD: Bullish crossover
- 1h Price: At lower BBand
- **All conditions met for BUY**

**Market Conditions:**
- BBands width: 12.5% (> 10% max ✓)
- Reason: Major news event causing extreme volatility

**Decision:**
- **NO TRADE - Volatility filter rejected signal**
- Reason: "Volatility too high (12.5% > 10%)"

**Outcome:** Avoided high-risk trade. Price spiked +15% then crashed -20% within hours.

---

### Scenario 4: Trend Misalignment

**Setup:**
- Daily trend: BEARISH (RSI: 35, MACD: -0.002)
- 1h signals: BUY (RSI oversold, MACD crossover)

**Decision:**
- **NO TRADE - Daily-hourly misalignment**
- Reason: "Trend misalignment: Daily=BEARISH, Hourly Signal=BUY"

**Outcome:** Daily downtrend continued, 1h bounce was temporary. Avoided counter-trend trap.

---

## Summary

This multi-timeframe strategy provides:

1. **Layered Confirmation:** Multiple timeframes reduce false signals
2. **Trend Alignment:** Daily filter keeps trades with the main trend
3. **Precise Entries:** 5-minute confirmation improves entry timing
4. **Risk Management:** Position sizing and stop-loss protect capital
5. **Market Filters:** Volatility and liquidity checks ensure tradeable conditions

**Key Principles:**
- Trade WITH the daily trend (or be aware when going against it)
- Require multiple indicator confluence (not just one signal)
- Size positions based on risk, not arbitrary amounts
- Use stop-losses religiously
- Monitor market conditions before entering

**Next Steps:**
1. Backtest this strategy on historical data
2. Paper trade for at least 30 days
3. Adjust thresholds based on your asset's characteristics
4. Add logging and monitoring dashboards
5. Gradually introduce real capital with small position sizes

---

**Created:** 2025-10-27
**Version:** 1.0
**Author:** Claude Code Assistant
