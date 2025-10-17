# Trading Strategy Guide

This guide explains how to use the different trading strategies available in the bot.

## Overview

The bot now supports multiple technical indicators and trading strategies through a flexible strategy pattern. This allows you to easily switch between different trading approaches without modifying code.

## Available Strategies

### 1. RSI Strategy (Relative Strength Index)

**File**: `configs/config-rsi.yaml`

**How it works**:
- Tracks momentum oscillator (0-100 scale)
- **BUY**: When RSI drops to/below oversold level (default 30)
- **SELL**: When RSI rises to/above overbought level (default 70)

**Best for**: Range-bound markets, mean reversion

**Configuration**:
```yaml
strategy:
  type: "rsi"
  overbought_level: 70.0
  oversold_level: 30.0
  indicator:
    type: "rsi"
    params:
      period: 14
```

---

### 2. MACD Strategy (Moving Average Convergence Divergence)

**File**: `configs/config-macd.yaml`

**How it works**:
- Tracks relationship between fast and slow moving averages
- **BUY**: When MACD line crosses ABOVE signal line (bullish crossover)
- **SELL**: When MACD line crosses BELOW signal line (bearish crossover)

**Best for**: Trending markets, momentum trading

**Configuration**:
```yaml
strategy:
  type: "macd"
  indicator:
    type: "macd"
    params:
      fast_period: 12
      slow_period: 26
      signal_period: 9
```

---

### 3. Bollinger Bands Strategy

**File**: `configs/config-bbands.yaml`

**How it works**:
- Creates dynamic price channels based on volatility
- **BUY**: When price touches/crosses BELOW lower band
- **SELL**: When price touches/crosses ABOVE upper band

**Best for**: Volatility trading, identifying price extremes

**Configuration**:
```yaml
strategy:
  type: "bbands"
  indicator:
    type: "bbands"
    params:
      period: 20
      std_dev: 2.0
```

---

## How to Switch Strategies

### Method 1: Use Pre-configured Files

Copy one of the example configs to `configs/config.yaml`:

```bash
# Use RSI strategy
cp configs/config-rsi.yaml configs/config.yaml

# Use MACD strategy
cp configs/config-macd.yaml configs/config.yaml

# Use Bollinger Bands strategy
cp configs/config-bbands.yaml configs/config.yaml
```

Then run the bot:
```bash
go run cmd/rsi-bot/main.go
```

### Method 2: Custom Configuration

Create your own `configs/config.yaml`:

```yaml
symbol: "BTCUSDT"
quantity: 0.001
trading_enabled: false

strategy:
  type: "macd"  # or "rsi" or "bbands"
  # RSI-specific (only for RSI strategy):
  overbought_level: 75.0
  oversold_level: 25.0

  indicator:
    type: "macd"  # must match strategy type
    params:
      fast_period: 10
      slow_period: 20
      signal_period: 5
```

---

## Architecture for Wails GUI

The bot is now structured to easily integrate with your Wails GUI:

### Backend Methods to Expose

```go
// In your Wails app.go

// Get list of available strategies
func (a *App) GetAvailableStrategies() []string {
    factory := strategy.NewFactory()
    return factory.GetAvailableStrategies()
}

// Set strategy at runtime
func (a *App) SetStrategy(strategyType string, params map[string]interface{}) error {
    // Create new strategy
    // Update bot.strategy
    // Return error if invalid
}

// Get current position
func (a *App) GetPosition() Position {
    return bot.position
}

// Toggle paper trading
func (a *App) SetTradingEnabled(enabled bool) {
    bot.config.TradingEnabled = enabled
}

// Get trade history (once database is added)
func (a *App) GetTradeHistory() []Trade {
    // Query database
}
```

### Vue.js Frontend Components

**Strategy Selector**:
```vue
<select v-model="selectedStrategy">
  <option value="rsi">RSI (Relative Strength Index)</option>
  <option value="macd">MACD (Trend Following)</option>
  <option value="bbands">Bollinger Bands (Volatility)</option>
</select>
```

**Dynamic Parameter Editor**:
```vue
<div v-if="selectedStrategy === 'rsi'">
  <input v-model="rsiPeriod" type="number" placeholder="Period">
  <input v-model="overboughtLevel" type="number" placeholder="Overbought">
  <input v-model="oversoldLevel" type="number" placeholder="Oversold">
</div>
```

**Paper Trading Toggle**:
```vue
<toggle v-model="paperTradingEnabled" label="Paper Trading" />
```

---

## Next Steps

### Completed ✅
- [x] MACD indicator implementation
- [x] Bollinger Bands indicator implementation
- [x] Strategy pattern interface
- [x] RSI/MACD/BBands strategies
- [x] Updated bot to use strategy pattern
- [x] Example config files for each strategy

### Remaining Tasks
- [ ] Add SQLite database for trade history
- [ ] Test order execution on testnet
- [ ] Build Wails GUI with strategy selector
- [ ] Add real-time charts for indicators
- [ ] Implement profit/loss tracking
- [ ] Add more strategies (EMA crossover, Stochastic RSI, etc.)

---

## Testing Recommendations

1. **Start with paper trading** (`trading_enabled: false`)
2. **Test each strategy separately** for 24 hours
3. **Monitor indicator values** in logs
4. **Verify buy/sell signals** against TradingView
5. **Only enable live trading** after thorough testing

---

## Strategy Performance Notes

Add your observations here as you test:

- **RSI**: Works well on SHIBUSDT in ranging conditions
- **MACD**: Better for trending coins like BTC/ETH
- **Bollinger Bands**: Effective during high volatility

---

## Customization Tips

### Aggressive RSI
```yaml
overbought_level: 80.0  # Higher threshold = fewer sells
oversold_level: 20.0    # Lower threshold = fewer buys
```

### Fast MACD
```yaml
fast_period: 8
slow_period: 17
signal_period: 9
```

### Tight Bollinger Bands
```yaml
std_dev: 1.5  # Narrower bands = more signals
```
