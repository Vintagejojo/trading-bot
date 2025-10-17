# Trading Bot Architecture - Multi-Strategy System

## Overview

This document outlines the architecture for expanding the RSI trading bot into a modular, multi-strategy system that supports:
1. **Independent strategies** - Run strategies separately
2. **Confirmation mode** - Multiple indicators must agree
3. **Weighted scoring** - Indicators vote with different weights
4. **Dynamic selection** - Automatically choose strategy based on market conditions
5. **P&L tracking** - Monitor performance over time

## Core Concepts

### Strategy vs Indicator

**Indicator**: A technical calculation (RSI, MACD, Bollinger Bands, etc.)
- Input: Price data
- Output: Numerical values (RSI value, MACD line, etc.)
- Stateless calculation logic

**Strategy**: Decision-making logic that uses one or more indicators
- Input: Indicator values + market context
- Output: Trading signals (BUY, SELL, HOLD) with confidence score
- Stateful, can track positions and conditions

### Signal System

Every strategy produces **Signals** with:
- **Action**: BUY, SELL, HOLD
- **Confidence**: 0.0 to 1.0 (how strong the signal is)
- **Reason**: Human-readable explanation
- **Metadata**: Strategy name, timestamp, indicator values

## Architectural Design

```
┌─────────────────────────────────────────────────────────────┐
│                      Market Data Stream                      │
│                    (WebSocket Klines)                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
         ┌───────────────────────────────┐
         │     Indicator Calculators     │
         │  (RSI, MACD, BBands, etc.)    │
         └───────┬───────────────────────┘
                 │ Indicator Values
                 ▼
         ┌───────────────────────────────┐
         │         Strategies            │
         │  ┌─────────────────────────┐  │
         │  │  RSI Strategy           │  │
         │  │  MACD Strategy          │  │
         │  │  BollingerBands Strategy│  │
         │  │  Composite Strategies   │  │
         │  └─────────────────────────┘  │
         └───────┬───────────────────────┘
                 │ Signals (Action + Confidence)
                 ▼
         ┌───────────────────────────────┐
         │      Execution Engine         │
         │  ┌─────────────────────────┐  │
         │  │ Mode: Independent       │  │
         │  │       Confirmation      │  │
         │  │       Weighted          │  │
         │  │       Dynamic           │  │
         │  └─────────────────────────┘  │
         └───────┬───────────────────────┘
                 │ Execute Decision
                 ▼
         ┌───────────────────────────────┐
         │      Order Executor           │
         │   (Paper/Live Trading)        │
         └───────┬───────────────────────┘
                 │
                 ▼
         ┌───────────────────────────────┐
         │      P&L Tracker              │
         │   (Performance Metrics)       │
         └───────────────────────────────┘
```

## File Structure

```
trading-bot/
├── internal/
│   ├── indicators/          # Pure calculations
│   │   ├── rsi.go
│   │   ├── macd.go
│   │   ├── bbands.go
│   │   └── stochastic_rsi.go
│   │
│   ├── strategies/          # Trading decision logic
│   │   ├── strategy.go      # Strategy interface
│   │   ├── rsi_strategy.go
│   │   ├── macd_strategy.go
│   │   ├── bbands_strategy.go
│   │   └── composite_strategy.go  # Multi-indicator strategies
│   │
│   ├── signals/             # Signal definitions
│   │   └── signal.go
│   │
│   ├── engine/              # Execution logic
│   │   ├── engine.go        # Main execution engine
│   │   ├── independent.go   # Independent mode
│   │   ├── confirmation.go  # Confirmation mode
│   │   ├── weighted.go      # Weighted scoring mode
│   │   └── dynamic.go       # Dynamic selection mode
│   │
│   ├── executor/            # Order execution
│   │   ├── executor.go
│   │   ├── paper.go         # Paper trading
│   │   └── live.go          # Live trading
│   │
│   ├── tracker/             # P&L tracking
│   │   ├── tracker.go
│   │   ├── trade.go
│   │   └── metrics.go
│   │
│   └── market/              # Market condition detection
│       └── conditions.go
```

## Interface Definitions

### 1. Indicator Interface

```go
package indicators

// Indicator represents a technical indicator calculator
type Indicator interface {
    // Name returns the indicator name (e.g., "RSI", "MACD")
    Name() string

    // Update adds new price data and recalculates
    Update(price float64, timestamp time.Time) error

    // GetValue returns the current indicator value(s)
    // For multi-value indicators (MACD), returns map with keys like "macd", "signal", "histogram"
    GetValue() (map[string]float64, bool) // value, isValid

    // IsReady returns true when enough data exists for calculation
    IsReady() bool

    // Reset clears all historical data
    Reset()
}
```

### 2. Signal Structure

```go
package signals

type Action string

const (
    ActionBuy  Action = "BUY"
    ActionSell Action = "SELL"
    ActionHold Action = "HOLD"
)

// Signal represents a trading signal from a strategy
type Signal struct {
    Action      Action             // BUY, SELL, HOLD
    Confidence  float64            // 0.0 to 1.0
    Reason      string             // Human-readable explanation
    Strategy    string             // Strategy name that generated this
    Timestamp   time.Time
    Metadata    map[string]float64 // Indicator values used

    // For position sizing (optional)
    SuggestedQuantity float64
}
```

### 3. Strategy Interface

```go
package strategies

import (
    "rsi-bot/internal/signals"
    "rsi-bot/internal/indicators"
)

// Strategy represents a trading strategy that generates signals
type Strategy interface {
    // Name returns the strategy identifier
    Name() string

    // Analyze receives indicator values and generates a signal
    Analyze(indicators map[string]indicators.Indicator, currentPrice float64) signals.Signal

    // GetRequiredIndicators returns list of indicator names needed
    GetRequiredIndicators() []string

    // Configure sets strategy parameters (thresholds, periods, etc.)
    Configure(params map[string]interface{}) error
}
```

### 4. Execution Engine Interface

```go
package engine

type ExecutionMode string

const (
    ModeIndependent  ExecutionMode = "independent"   // Run one strategy
    ModeConfirmation ExecutionMode = "confirmation"  // All must agree
    ModeWeighted     ExecutionMode = "weighted"      // Weighted voting
    ModeDynamic      ExecutionMode = "dynamic"       // Auto-select strategy
)

type Engine interface {
    // SetMode configures how strategies are combined
    SetMode(mode ExecutionMode) error

    // AddStrategy registers a strategy for execution
    AddStrategy(strategy strategies.Strategy, weight float64) error

    // Evaluate processes all strategies and decides on action
    Evaluate(indicators map[string]indicators.Indicator, currentPrice float64) (*signals.Signal, error)

    // SetThreshold for confirmation/weighted modes (0.0 to 1.0)
    SetThreshold(threshold float64)
}
```

## Execution Modes Explained

### Mode 1: Independent

Run a single strategy in isolation.

**Configuration:**
```yaml
execution:
  mode: independent
  strategy: rsi_strategy
```

**Logic:**
- Only one strategy is active
- Direct signal → execution
- Simple and clean for single-indicator trading

**Use case:** Testing individual strategies, specialized market conditions

### Mode 2: Confirmation

All strategies must agree on the same action.

**Configuration:**
```yaml
execution:
  mode: confirmation
  strategies:
    - rsi_strategy
    - macd_strategy
    - bbands_strategy
  min_confidence: 0.7  # Each strategy must be at least 70% confident
```

**Logic:**
```
RSI Signal:    BUY (0.85)
MACD Signal:   BUY (0.75)
BBands Signal: BUY (0.70)
→ Result: BUY (average confidence: 0.77)

RSI Signal:    BUY (0.85)
MACD Signal:   SELL (0.60)
BBands Signal: BUY (0.70)
→ Result: HOLD (no agreement)
```

**Use case:** Conservative trading, reducing false signals

### Mode 3: Weighted Scoring

Strategies vote with different weights, must reach threshold.

**Configuration:**
```yaml
execution:
  mode: weighted
  strategies:
    - name: rsi_strategy
      weight: 0.6
    - name: macd_strategy
      weight: 0.4
  threshold: 0.8  # 80% weighted confidence to execute
```

**Logic:**
```
RSI Signal:  BUY (confidence: 0.9) × 0.6 = 0.54
MACD Signal: BUY (confidence: 0.7) × 0.4 = 0.28
Total Score: 0.82 → EXECUTE BUY

RSI Signal:  BUY (confidence: 0.6) × 0.6 = 0.36
MACD Signal: SELL (confidence: 0.5) × 0.4 = -0.20 (opposite action = negative)
Total Score: 0.16 → HOLD
```

**Use case:** Balanced approach, prioritize proven indicators

### Mode 4: Dynamic Selection

System detects market conditions and selects best strategy.

**Configuration:**
```yaml
execution:
  mode: dynamic
  market_detection:
    volatility_period: 20
    trend_period: 50
  strategy_mapping:
    trending: macd_strategy
    ranging: rsi_strategy
    high_volatility: bbands_strategy
```

**Logic:**
1. Analyze market conditions (ATR for volatility, ADX for trend strength)
2. Select appropriate strategy based on current condition
3. Execute that strategy's signal

**Market Conditions:**
- **Trending**: Strong directional movement (use MACD)
- **Ranging**: Price oscillating in range (use RSI)
- **High Volatility**: Large price swings (use Bollinger Bands)

**Use case:** Adaptive trading, automatic optimization

## P&L Tracking System

### Trade Structure

```go
package tracker

type Trade struct {
    ID            string
    Symbol        string
    Side          string    // "BUY" or "SELL"
    Quantity      float64
    EntryPrice    float64
    ExitPrice     float64
    EntryTime     time.Time
    ExitTime      time.Time
    Strategy      string
    ProfitLoss    float64   // In quote currency (USDT)
    ProfitPercent float64   // Percentage return
    Status        string    // "OPEN", "CLOSED"
    IsPaperTrade  bool
}

type Metrics struct {
    TotalTrades      int
    WinningTrades    int
    LosingTrades     int
    WinRate          float64
    TotalProfit      float64
    TotalLoss        float64
    NetProfit        float64
    AverageProfit    float64
    AverageLoss      float64
    LargestWin       float64
    LargestLoss      float64
    ProfitFactor     float64  // Total profit / Total loss

    // By Strategy
    StrategyMetrics  map[string]*Metrics
}
```

### Tracker Interface

```go
type Tracker interface {
    // RecordTrade logs a completed trade
    RecordTrade(trade *Trade) error

    // OpenPosition records a new position
    OpenPosition(symbol, strategy string, side string, quantity, price float64) (*Trade, error)

    // ClosePosition marks position as closed and calculates P&L
    ClosePosition(tradeID string, exitPrice float64, exitTime time.Time) error

    // GetMetrics returns performance statistics
    GetMetrics() *Metrics

    // GetStrategyMetrics returns metrics for a specific strategy
    GetStrategyMetrics(strategyName string) *Metrics

    // GetOpenPositions returns all currently open trades
    GetOpenPositions() []*Trade

    // Export saves trades to CSV/JSON
    Export(filepath string, format string) error
}
```

### Storage Options

**Option 1: In-Memory** (simplest, current phase)
```go
type MemoryTracker struct {
    trades       []*Trade
    openTrades   map[string]*Trade
    metrics      *Metrics
}
```

**Option 2: SQLite** (Phase 3 - persistent)
```sql
CREATE TABLE trades (
    id TEXT PRIMARY KEY,
    symbol TEXT,
    side TEXT,
    quantity REAL,
    entry_price REAL,
    exit_price REAL,
    entry_time DATETIME,
    exit_time DATETIME,
    strategy TEXT,
    profit_loss REAL,
    profit_percent REAL,
    status TEXT,
    is_paper_trade BOOLEAN
);
```

## Implementation Phases

### Phase 2.5: Extract Indicators (Current Priority)

**Goal**: Separate RSI calculation from strategy logic

**Tasks**:
1. Move `calculator/rsi.go` → `indicators/rsi.go` and implement Indicator interface
2. Update `bot/bot.go` to use indicator interface
3. Test that existing RSI bot still works

**Files to modify**:
- `internal/calculator/rsi.go` → `internal/indicators/rsi.go`
- `internal/bot/bot.go`

### Phase 3: Add Strategy Layer

**Goal**: Create strategy abstraction and RSI strategy implementation

**Tasks**:
1. Create `signals/signal.go` with Signal struct
2. Create `strategies/strategy.go` interface
3. Implement `strategies/rsi_strategy.go` using RSI indicator
4. Update bot to use strategy instead of direct RSI calculation

**New files**:
- `internal/signals/signal.go`
- `internal/strategies/strategy.go`
- `internal/strategies/rsi_strategy.go`

### Phase 4: Add More Indicators

**Goal**: Implement MACD, Bollinger Bands, Stochastic RSI

**Tasks**:
1. Implement `indicators/macd.go`
2. Implement `indicators/bbands.go`
3. Implement `indicators/stochastic_rsi.go`
4. Create corresponding strategy implementations

**New files**:
- `internal/indicators/macd.go`
- `internal/indicators/bbands.go`
- `internal/indicators/stochastic_rsi.go`
- `internal/strategies/macd_strategy.go`
- `internal/strategies/bbands_strategy.go`

### Phase 5: Execution Engine

**Goal**: Support multiple execution modes

**Tasks**:
1. Create `engine/engine.go` with mode support
2. Implement independent mode (default)
3. Implement confirmation mode
4. Implement weighted mode
5. Implement dynamic mode with market condition detection

**New files**:
- `internal/engine/engine.go`
- `internal/engine/independent.go`
- `internal/engine/confirmation.go`
- `internal/engine/weighted.go`
- `internal/engine/dynamic.go`
- `internal/market/conditions.go`

### Phase 6: P&L Tracking

**Goal**: Track performance metrics

**Tasks**:
1. Create `tracker/tracker.go` with in-memory storage
2. Create `tracker/metrics.go` for calculations
3. Integrate with order executor
4. Add export to CSV/JSON
5. (Optional) Migrate to SQLite for persistence

**New files**:
- `internal/tracker/tracker.go`
- `internal/tracker/trade.go`
- `internal/tracker/metrics.go`

## Configuration Examples

### Example 1: Independent RSI Strategy
```yaml
execution:
  mode: independent
  strategy: rsi_strategy

strategies:
  rsi_strategy:
    period: 14
    overbought: 70
    oversold: 30
    confidence_multiplier: 1.0  # Full confidence when thresholds hit

symbol: "BTCUSDT"
quantity: 0.001
```

### Example 2: Confirmation Mode (Conservative)
```yaml
execution:
  mode: confirmation
  strategies:
    - rsi_strategy
    - macd_strategy
    - bbands_strategy
  min_confidence: 0.7

strategies:
  rsi_strategy:
    period: 14
    overbought: 70
    oversold: 30

  macd_strategy:
    fast_period: 12
    slow_period: 26
    signal_period: 9

  bbands_strategy:
    period: 20
    std_dev: 2.0
```

### Example 3: Weighted Scoring (Balanced)
```yaml
execution:
  mode: weighted
  threshold: 0.75
  strategies:
    - name: rsi_strategy
      weight: 0.5
    - name: macd_strategy
      weight: 0.3
    - name: bbands_strategy
      weight: 0.2
```

### Example 4: Dynamic Selection (Adaptive)
```yaml
execution:
  mode: dynamic
  market_detection:
    volatility_indicator: atr
    volatility_period: 14
    trend_indicator: adx
    trend_period: 14
    update_interval: 1h

  strategy_mapping:
    trending_up: macd_strategy
    trending_down: rsi_strategy
    ranging: rsi_strategy
    high_volatility: bbands_strategy
    low_volatility: rsi_strategy
```

## Next Steps

1. **Review this architecture** - Does it meet your requirements?
2. **Choose starting phase** - Recommend Phase 2.5 (extract indicators)
3. **Implementation order** - I can help implement each phase step by step
4. **Testing strategy** - Each phase should be tested independently before moving forward

Would you like me to start implementing Phase 2.5 (extracting indicators from the current RSI calculator)?
