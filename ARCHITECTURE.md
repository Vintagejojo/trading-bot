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

### Phase 7: Safety & Resilience

**Goal**: Implement critical safety mechanisms and resilience features to protect against catastrophic losses and ensure system stability

**Tasks**:

#### 🚨 Emergency Lockouts
1. **Circuit Breaker Logic**
   - Halt all trading if drawdown exceeds 10% in 1 hour
   - Implement rolling window P&L monitoring
   - Auto-disable trading until manual override or cooldown period expires
   - Send emergency alerts via configured channels (email, SMS, webhook)

2. **Max Daily Loss Threshold**
   - Configure daily loss cap (e.g., -5% of portfolio)
   - Disable trading immediately when threshold is breached
   - Reset counter at market open or configurable time
   - Log breach events with full context

3. **Trade Frequency Guardrails**
   - Lockout if trade count exceeds X trades in Y minutes
   - Prevent rapid-fire trading loops from bugs or market anomalies
   - Configurable sliding window (e.g., max 10 trades per 15 minutes)
   - Implement exponential backoff on repeated lockouts

4. **API Health Checks**
   - Monitor API latency and response times
   - Pause trading on latency spikes (>2s response time)
   - Detect failed endpoints and rotate to backup if available
   - Track API error rates and auto-pause on elevated errors (>5%)

#### 💧 Liquidity Safeguards
1. **Minimum Order Book Depth**
   - Verify sufficient liquidity before placing orders
   - Check bid/ask depth at multiple price levels (e.g., 1%, 2%, 5%)
   - Skip trades if order book is too thin (risk of slippage)
   - Configurable minimum depth thresholds per symbol

2. **Spread Thresholds**
   - Calculate bid-ask spread percentage
   - Skip trading pairs with excessive spreads (e.g., >0.5%)
   - Monitor spread widening as early warning signal
   - Log spread violations for analysis

3. **Volume Filters**
   - Require minimum 24-hour volume before trading
   - Check recent candle volume for sudden drops
   - Configurable volume thresholds (e.g., $1M daily minimum)
   - Detect volume anomalies that may indicate manipulation

4. **Slippage Simulation**
   - Simulate order execution against order book
   - Calculate expected slippage for order size
   - Abort trades if simulated slippage exceeds threshold (e.g., >1%)
   - Use VWAP for better slippage estimation

#### 🧠 Smart Recovery Logic
1. **Graceful Degradation**
   - Rotate to backup trading pairs if primary pair fails liquidity checks
   - Switch to passive strategy (wider stops, lower frequency) during high volatility
   - Reduce position sizes automatically during adverse conditions
   - Maintain fallback strategy queue

2. **Manual Override Dashboard**
   - Web UI controls to pause/resume trading
   - Require reason codes for manual interventions
   - Override emergency lockouts with authentication
   - Display system health status and active safety triggers

3. **Logging & Alerting**
   - Log all lockout events with:
     - Timestamp
     - Trigger condition (which safety rule fired)
     - System state snapshot (positions, P&L, market conditions)
     - Recovery actions taken
   - Alert channels:
     - Critical: Emergency lockouts, API failures
     - Warning: Spread violations, volume drops
     - Info: Trade frequency approaching limits
   - Configurable alert destinations (email, Slack, Telegram, webhook)

**New files**:
- `internal/safety/circuit_breaker.go` - Drawdown and loss limit monitoring
- `internal/safety/frequency_limiter.go` - Trade frequency guardrails
- `internal/safety/api_health.go` - API monitoring and health checks
- `internal/safety/liquidity_guard.go` - Order book depth and spread checks
- `internal/safety/slippage_simulator.go` - Pre-trade slippage estimation
- `internal/safety/recovery.go` - Graceful degradation and failover logic
- `internal/safety/override.go` - Manual control and authentication
- `internal/safety/alerts.go` - Multi-channel alerting system
- `internal/safety/config.go` - Safety configuration and thresholds

**Configuration Example**:
```yaml
safety:
  circuit_breaker:
    enabled: true
    max_drawdown_1h: 10.0      # Halt if -10% in 1 hour
    max_daily_loss: 5.0         # Halt if -5% in one day
    cooldown_minutes: 60        # Wait 1 hour before auto-resume

  trade_frequency:
    enabled: true
    max_trades_per_15min: 10
    max_trades_per_hour: 30
    lockout_minutes: 30

  api_health:
    enabled: true
    max_latency_ms: 2000
    max_error_rate: 0.05        # 5%
    check_interval_seconds: 30

  liquidity:
    min_order_book_depth_usd: 50000
    max_spread_percent: 0.5
    min_24h_volume_usd: 1000000
    min_candle_volume_usd: 10000

  slippage:
    enabled: true
    max_slippage_percent: 1.0
    use_vwap: true

  recovery:
    backup_symbols:
      - "ETHUSDT"
      - "BNBUSDT"
    passive_strategy: "conservative_rsi"
    position_size_multiplier: 0.5  # Reduce size by 50% during recovery

  alerts:
    email:
      enabled: true
      smtp_server: "smtp.gmail.com:587"
      from: "bot@example.com"
      to: ["trader@example.com"]
    webhook:
      enabled: true
      url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    levels:
      critical: ["circuit_breaker", "api_failure"]
      warning: ["spread_violation", "volume_drop"]
      info: ["frequency_warning"]
```

**Critical Safety Notes**:
- All safety features should default to **ENABLED** in production
- Safety checks run BEFORE any order execution
- Failed safety checks must log detailed context
- Manual overrides require authentication and are logged
- Test safety features thoroughly in paper trading mode

### Phase 8: Multi-Timeframe Analysis

**Goal**: Implement multi-timeframe chart support for better signal reliability and reduced false positives through cross-timeframe confirmation

**Tasks**:

#### 📊 Multi-Timeframe Chart Integration
1. **Timeframe Management**
   - Support for multiple chart intervals: 1m, 5m, 15m, 1h, 4h, daily
   - Concurrent WebSocket subscriptions for each active timeframe
   - Efficient data storage and synchronization across timeframes
   - Automatic candle aggregation (e.g., build 15m candles from 1m data)

2. **Indicator Calculations Per Timeframe**
   - Run RSI, MACD, Bollinger Bands on each configured timeframe
   - Independent indicator state per timeframe
   - Timestamped indicator values for alignment
   - Example: `RSI_1m`, `RSI_5m`, `RSI_15m`, `RSI_1h` all calculated separately

3. **Data Structures**
   - Extend indicator interface to support timeframe parameter
   - Store historical candles per timeframe
   - Efficient lookback window management
   - Memory-conscious data retention policies

#### 🔄 Signal Layering & Confirmation
1. **Cross-Timeframe Signal Logic**
   - **Higher Timeframe Trend Filter**: Only trade in direction of higher timeframe trend
     - Example: Only BUY signals if 1h MACD is bullish
   - **Lower Timeframe Entry**: Use faster timeframes for precise entry
     - Example: 15m RSI oversold + 1h trend confirmation = BUY
   - **Multi-Layer Confirmation**: Require agreement across multiple timeframes
     - Example: 5m RSI oversold + 15m RSI oversold + 1h trend up = Strong BUY

2. **Timeframe Priority Levels**
   - **Primary**: Main decision timeframe (e.g., 15m)
   - **Confirmation**: Higher timeframe for trend validation (e.g., 1h)
   - **Entry**: Lower timeframe for precise timing (e.g., 5m)
   - Configurable weights per timeframe in signal scoring

3. **Signal Strength Scoring**
   - Signals gain strength when aligned across multiple timeframes
   - Penalty for conflicting signals across timeframes
   - Example scoring:
     ```
     15m RSI BUY (0.8) + 1h MACD BUY (0.9) + 5m RSI BUY (0.7) = 0.85 confidence
     15m RSI BUY (0.8) + 1h MACD SELL (0.6) = 0.4 confidence (conflict penalty)
     ```

#### ⚙️ Configurable Timeframes
1. **Per-Indicator Timeframe Settings**
   - Allow users to specify which timeframes to use per indicator
   - Example config:
     ```yaml
     indicators:
       rsi:
         timeframes: [5m, 15m, 1h]
         period: 14
       macd:
         timeframes: [15m, 1h, 4h]
         fast_period: 12
     ```

2. **Strategy-Level Timeframe Configuration**
   - Define timeframe combinations per strategy
   - Pre-configured templates (e.g., "Scalping", "Swing Trading", "Position Trading")
   - Example:
     ```yaml
     strategies:
       scalping:
         primary_timeframe: 5m
         confirmation_timeframe: 15m
         entry_timeframe: 1m

       swing_trading:
         primary_timeframe: 1h
         confirmation_timeframe: 4h
         entry_timeframe: 15m
     ```

3. **Dashboard Controls**
   - UI to enable/disable specific timeframes
   - Visual display of active timeframes per indicator
   - Real-time switching without bot restart
   - Timeframe health indicators (data quality, sync status)

#### 🧪 Backtesting Compatibility
1. **Historical Multi-Timeframe Data**
   - Load and align historical data across all timeframes
   - Handle missing data and gaps
   - Ensure proper chronological order across timeframes
   - Validate data consistency (e.g., 5m candles aggregate to 15m)

2. **Backtest Engine Updates**
   - Simulate multi-timeframe signal generation
   - Accurate timestamp alignment in backtests
   - Performance metrics per timeframe combination
   - A/B testing different timeframe configurations

3. **Trade Log Enhancements**
   - Record which timeframes contributed to each signal
   - Store indicator values from all active timeframes
   - Enable post-trade analysis by timeframe
   - Example log entry:
     ```json
     {
       "trade_id": "abc123",
       "signal": "BUY",
       "confidence": 0.85,
       "timeframes": {
         "5m": {"rsi": 28, "signal": "BUY", "confidence": 0.7},
         "15m": {"rsi": 32, "signal": "BUY", "confidence": 0.8},
         "1h": {"macd": 0.5, "signal": "BUY", "confidence": 0.9}
       }
     }
     ```

#### 🔮 Future Expansion & Scalability
1. **Additional Timeframe Support**
   - Structure code to easily add new intervals (2m, 30m, 6h, weekly)
   - Generic timeframe parsing and validation
   - Automatic candle aggregation for non-native intervals
   - No hardcoded timeframe assumptions

2. **Performance Optimization**
   - Lazy loading of timeframe data (only load what's needed)
   - Shared WebSocket connections where possible
   - Candle caching and reuse across indicators
   - Intelligent update triggering (only recalculate when new data arrives)

3. **Advanced Features**
   - Adaptive timeframe selection based on volatility
   - Automatic timeframe optimization via machine learning
   - Correlation analysis across timeframes
   - Divergence detection between timeframes

**New files**:
- `internal/timeframe/manager.go` - Timeframe management and synchronization
- `internal/timeframe/candle_store.go` - Multi-timeframe candle storage
- `internal/timeframe/aggregator.go` - Candle aggregation (1m → 5m, 15m, etc.)
- `internal/strategies/multi_timeframe_strategy.go` - Cross-timeframe strategy base
- `internal/signals/timeframe_signal.go` - Extended signal with timeframe context
- `pkg/websocket/multi_stream.go` - Multiple WebSocket stream management
- `internal/backtest/timeframe_engine.go` - Multi-timeframe backtesting

**Configuration Example**:
```yaml
timeframes:
  enabled: [1m, 5m, 15m, 1h, 4h]
  primary: 15m          # Main decision timeframe

  data_retention:
    1m: 1000            # Keep 1000 candles (~16 hours)
    5m: 500             # Keep 500 candles (~42 hours)
    15m: 300            # Keep 300 candles (~75 hours)
    1h: 200             # Keep 200 candles (~8 days)
    4h: 100             # Keep 100 candles (~16 days)

  websocket:
    aggregate_from: 1m   # Build all timeframes from 1m stream
    native_streams: [1m, 5m, 15m]  # Or subscribe to native streams

strategies:
  rsi_multi_tf:
    type: multi_timeframe
    primary_timeframe: 15m

    indicators:
      rsi:
        timeframes:
          5m:
            period: 14
            weight: 0.2
          15m:
            period: 14
            weight: 0.5
          1h:
            period: 14
            weight: 0.3

    rules:
      # Only trade if 1h trend agrees
      trend_filter:
        timeframe: 1h
        indicator: macd
        required: true

      # Primary entry on 15m
      entry:
        timeframe: 15m
        indicator: rsi
        oversold: 30
        overbought: 70

      # Additional confirmation from 5m
      confirmation:
        timeframe: 5m
        indicator: rsi
        oversold: 35
        overbought: 65
        required: false  # Optional boost

    signal_logic: weighted  # Options: all_agree, weighted, majority

backtesting:
  multi_timeframe:
    enabled: true
    validate_alignment: true
    log_all_timeframes: true
```

**Implementation Priorities**:
1. **Phase 8.1**: Basic multi-timeframe data management (1m, 5m, 15m, 1h)
2. **Phase 8.2**: Simple cross-timeframe confirmation (higher TF trend filter)
3. **Phase 8.3**: Full signal layering with configurable weights
4. **Phase 8.4**: Backtesting integration and performance analysis
5. **Phase 8.5**: Dashboard UI for timeframe management
6. **Phase 8.6**: Advanced features (adaptive selection, optimization)

**Benefits**:
- **Reduced False Signals**: Multi-timeframe confirmation filters out noise
- **Better Entries**: Use lower timeframes for precise entry points
- **Trend Alignment**: Ensure trades align with higher timeframe trends
- **Flexibility**: Adapt to different trading styles (scalping, swing, position)
- **Backtestable**: Validate timeframe combinations historically

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
