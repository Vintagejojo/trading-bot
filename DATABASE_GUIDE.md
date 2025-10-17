# Database Integration Guide

## Overview

The trading bot now includes **SQLite database integration** for persistent storage of trades, positions, and performance statistics. This enables trade history tracking, profit/loss analysis, and recovery from bot restarts.

## Database Features

### ✅ What's Implemented

1. **Trade Logging**
   - Every buy and sell is recorded with full details
   - Indicator values at time of trade (JSON)
   - Strategy name and signal reason
   - Paper trading vs. live trading flag
   - Binance order IDs (for real trades)
   - Profit/loss calculations (for sell trades)

2. **Position Tracking**
   - Open and closed positions stored
   - Entry/exit prices and times
   - Profit/loss per position
   - Links to related buy/sell trades

3. **Performance Statistics**
   - Total trades (buys/sells)
   - Win rate percentage
   - Average profit/loss
   - Largest win/loss
   - Date range of activity

4. **Position Recovery**
   - Bot automatically restores open positions on restart
   - No need to manually track what you're holding

## Database Schema

### `trades` Table
```sql
- id (PRIMARY KEY)
- symbol (e.g., "SHIBUSDT")
- side ("BUY" or "SELL")
- quantity
- price
- total (quantity * price)
- strategy ("RSI", "MACD", "BBands")
- indicator_values (JSON string)
- signal_reason (human-readable)
- paper_trade (boolean)
- timestamp
- binance_order_id
- profit_loss (for SELL trades)
- profit_loss_percent (for SELL trades)
- related_buy_id (links SELL to BUY)
```

### `positions` Table
```sql
- id (PRIMARY KEY)
- symbol
- quantity
- entry_price
- entry_time
- exit_price
- exit_time
- strategy
- is_open (boolean)
- profit_loss
- profit_loss_percent
- buy_trade_id (FK to trades)
- sell_trade_id (FK to trades)
```

## Database File

**Location**: `trading_bot.db` (created automatically in project root)

**Driver**: Pure-Go SQLite (`modernc.org/sqlite`) - no CGO required

**Backup**: Copy `trading_bot.db` to backup your entire trade history

## How It Works

### 1. On Bot Startup
```go
// Bot checks for existing open position
dbPosition, err := db.GetOpenPosition(symbol)
if dbPosition != nil {
    // Restore position from database
    position.InPosition = true
    position.Quantity = dbPosition.Quantity
    position.EntryPrice = dbPosition.EntryPrice
}
```

### 2. On BUY Signal
```go
// Execute order (if trading_enabled)
orderID := executeBuyOrder(price)

// Log trade to database
trade := &Trade{
    Symbol: "SHIBUSDT",
    Side: "BUY",
    Quantity: 150000,
    Price: 0.00001234,
    Strategy: "RSI",
    IndicatorValues: `{"rsi": 28.5}`,
    SignalReason: "RSI 28.5 <= 30.0 (OVERSOLD)",
    PaperTrade: true,
    Timestamp: now,
}
db.InsertTrade(trade)

// Create new position
position := &Position{
    Symbol: "SHIBUSDT",
    Quantity: 150000,
    EntryPrice: 0.00001234,
    Strategy: "RSI",
    IsOpen: true,
    BuyTradeID: tradeID,
}
db.InsertPosition(position)
```

### 3. On SELL Signal
```go
// Calculate profit/loss
profitLoss = (exitPrice - entryPrice) * quantity
profitPercent = ((exitPrice - entryPrice) / entryPrice) * 100

// Execute order
orderID := executeSellOrder(price)

// Log trade
trade := &Trade{
    Side: "SELL",
    ProfitLoss: 12.50,
    ProfitLossPercent: 5.2,
    // ... other fields
}
db.InsertTrade(trade)

// Close position
db.UpdatePosition(positionID, exitPrice, exitTime, profitLoss, profitPercent, sellTradeID)
```

## Querying Trade History

### From Command Line (SQLite CLI)

```bash
sqlite3 trading_bot.db

# View recent trades
SELECT * FROM trades ORDER BY timestamp DESC LIMIT 10;

# View all positions
SELECT * FROM positions ORDER BY entry_time DESC;

# Calculate total profit/loss
SELECT SUM(profit_loss) FROM trades WHERE side = 'SELL';

# View profitable trades only
SELECT * FROM trades WHERE profit_loss > 0 ORDER BY profit_loss DESC;
```

### From Code (Bot Methods)

```go
// Get recent trades
trades, err := bot.GetRecentTrades(20)

// Get trades by date range
start := time.Now().AddDate(0, 0, -7) // Last 7 days
end := time.Now()
trades, err := bot.GetTradesByDateRange(start, end)

// Get summary statistics
summary, err := bot.GetTradeSummary()
fmt.Printf("Total Trades: %d\n", summary.TotalTrades)
fmt.Printf("Win Rate: %.2f%%\n", summary.WinRate)
fmt.Printf("Total P/L: $%.2f\n", summary.TotalProfitLoss)
```

## Integration with Wails GUI

The database is ready for your GUI. Here's how to expose it:

### Backend (Wails app.go)

```go
type App struct {
    bot *bot.Bot
}

// Get trade history for display
func (a *App) GetTradeHistory(limit int) []database.Trade {
    trades, _ := a.bot.GetRecentTrades(limit)
    return trades
}

// Get performance statistics
func (a *App) GetStats() database.TradeSummary {
    summary, _ := a.bot.GetTradeSummary()
    return *summary
}

// Get current position
func (a *App) GetCurrentPosition() *database.Position {
    pos, _ := a.bot.GetOpenPosition()
    return pos
}
```

### Frontend (Vue.js)

```vue
<template>
  <div class="dashboard">
    <!-- Statistics Card -->
    <div class="stats">
      <h3>Performance</h3>
      <p>Total Trades: {{ stats.total_trades }}</p>
      <p>Win Rate: {{ stats.win_rate.toFixed(2) }}%</p>
      <p>Total P/L: ${{ stats.total_profit_loss.toFixed(2) }}</p>
    </div>

    <!-- Trade History Table -->
    <table class="trades">
      <thead>
        <tr>
          <th>Time</th>
          <th>Side</th>
          <th>Price</th>
          <th>Quantity</th>
          <th>P/L</th>
          <th>Strategy</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="trade in trades" :key="trade.id">
          <td>{{ formatTime(trade.timestamp) }}</td>
          <td :class="trade.side.toLowerCase()">{{ trade.side }}</td>
          <td>{{ trade.price.toFixed(8) }}</td>
          <td>{{ trade.quantity }}</td>
          <td :class="trade.profit_loss > 0 ? 'profit' : 'loss'">
            {{ trade.profit_loss ? trade.profit_loss.toFixed(2) : '-' }}
          </td>
          <td>{{ trade.strategy }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
export default {
  data() {
    return {
      trades: [],
      stats: {}
    }
  },
  mounted() {
    this.loadData()
  },
  methods: {
    async loadData() {
      this.trades = await window.go.main.App.GetTradeHistory(50)
      this.stats = await window.go.main.App.GetStats()
    }
  }
}
</script>
```

## Data Visualization Ideas

Once you have the Wails GUI, you can add:

1. **Equity Curve Chart**
   - X-axis: Time
   - Y-axis: Cumulative profit/loss
   - Shows trading performance over time

2. **Win/Loss Distribution**
   - Pie chart showing % wins vs. losses
   - Bar chart of profit/loss per trade

3. **Strategy Comparison**
   - Compare RSI vs. MACD vs. BBands performance
   - Group trades by strategy and show stats

4. **Indicator Chart**
   - Show historical indicator values
   - Overlay buy/sell points on price chart

## Database Maintenance

### Backup

```bash
# Simple copy
cp trading_bot.db trading_bot_backup_$(date +%Y%m%d).db

# Or use SQLite backup command
sqlite3 trading_bot.db ".backup trading_bot_backup.db"
```

### Reset

```bash
# Delete database to start fresh
rm trading_bot.db

# Bot will recreate it on next run
```

### Export to CSV

```bash
sqlite3 trading_bot.db

.mode csv
.output trades.csv
SELECT * FROM trades;
.output stdout
```

## Performance Considerations

- Database file size: ~1KB per trade (very small)
- 10,000 trades = ~10MB database file
- Queries are indexed for fast lookups
- No performance impact on bot execution

## Troubleshooting

### "database is locked" error
- Close any SQLite GUI tools viewing the database
- Only one process should write at a time

### Missing trades
- Check that `trading_bot.db` exists in project root
- Check bot logs for database errors
- Verify `go build` completed successfully

### Incorrect profit/loss
- Ensure bot is tracking positions correctly
- Check that buy and sell are properly linked (`related_buy_id`)
- Verify quantity matches between buy and sell

## Next Steps

- ✅ Database implemented and tested
- ⬜ Add Wails GUI with trade history display
- ⬜ Implement real-time equity curve chart
- ⬜ Add CSV export functionality in GUI
- ⬜ Create strategy backtesting using historical data
