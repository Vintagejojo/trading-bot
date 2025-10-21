# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a cryptocurrency trading bot for Binance that uses RSI (Relative Strength Index) technical analysis to generate automated buy/sell signals. The project consists of two main components:
1. **CLI trading bot** (`cmd/rsi-bot`) - Core trading logic with real-time WebSocket data processing
2. **Wails UI application** (`trading-bot-ui`) - Desktop GUI interface (in early development)

## Architecture

### Standard Go Project Layout

```
trading-bot/
├── cmd/rsi-bot/           # Main CLI executable
├── internal/              # Private application packages
│   ├── bot/              # Core trading bot logic and WebSocket handling
│   ├── indicators/       # Technical indicator implementations (RSI, MACD, etc.)
│   ├── config/           # Configuration loading (Viper)
│   └── models/           # Shared data structures (Config, Position, KlineEvent)
├── configs/              # YAML configuration files
└── trading-bot-ui/       # Wails desktop application (separate module)
```

### Key Components

**Bot (`internal/bot/bot.go`)**:
- Manages WebSocket connection to Binance (tries multiple endpoints for reliability)
- Processes real-time 1-minute kline/candlestick data
- Implements RSI-based trading strategy with position tracking
- Handles automatic reconnection with ping/pong keep-alive mechanism
- Executes buy/sell orders via Binance API (go-binance/v2 library)

**Indicators (`internal/indicators/`)**:
- **Interface** (`indicator.go`): Common interface for all technical indicators
- **RSI** (`rsi.go`): Relative Strength Index implementation
  - Maintains rolling window of price data (period + 20 buffer)
  - Calculates RSI using standard formula: RSI = 100 - (100 / (1 + RS))
  - Requires (period + 1) data points for valid calculation
  - Implements Indicator interface for extensibility

**Configuration (`internal/config/config.go`)**:
- Loads YAML config using Viper with sensible defaults
- API credentials loaded from `.env` file via godotenv
- Supports paper trading mode (`trading_enabled: false`)

### Trading Logic

- **BUY Signal**: RSI <= oversold_level (default 30) AND no existing position
- **SELL Signal**: RSI >= overbought_level (default 70) AND holding position
- Position tracking includes entry price, quantity, and profit/loss calculation
- All trades logged with emoji indicators for easy monitoring

## Common Commands

### Running the Bot

```bash
# CLI bot (from project root)
go run cmd/rsi-bot/main.go

# Build executable
go build -o bin/rsi-bot cmd/rsi-bot/main.go
./bin/rsi-bot
```

### Wails UI Application

```bash
cd trading-bot-ui

# Development mode (hot-reload)
wails dev

# Build for production
wails build

# Install frontend dependencies
cd frontend && npm install
```

### Testing & Development

```bash
# Install/update dependencies
go mod tidy

# Format code
go fmt ./...

# Run with specific config
# (Edit configs/config.yaml before running)
```

## Configuration

### Main Config (`configs/config.yaml`)
```yaml
symbol: "SHIBUSDT"           # Trading pair
rsi_period: 14               # RSI calculation period
overbought_level: 70.0       # Sell signal threshold
oversold_level: 30.0         # Buy signal threshold
quantity: 150000.0           # Order size
trading_enabled: false       # ALWAYS start with false (paper trading)
```

### Environment Variables (`.env`)
```
BINANCE_API_KEY=your_key_here
BINANCE_API_SECRET=your_secret_here
```

**CRITICAL**: The bot uses Binance Testnet (`https://testnet.binance.vision`) for API calls. API credentials are loaded from environment and NEVER committed to git.

## Important Implementation Details

### WebSocket Connection Strategy
The bot attempts multiple Binance WebSocket endpoints for redundancy:
1. `wss://stream.binance.com:9443/ws/{symbol}@kline_1m`
2. `wss://stream.binance.com/ws/{symbol}@kline_1m`
3. `wss://data-stream.binance.vision/ws/{symbol}@kline_1m`

Connection includes proper headers (User-Agent, Origin) and implements ping/pong keep-alive every 30 seconds with 60-second read deadline.

### Position Management
- Single position tracking (not portfolio-based)
- Thread-safe position updates on successful trades
- Profit/loss calculated as percentage of entry price
- Position state: InPosition (bool), Quantity, EntryPrice, LastUpdate

### Order Execution
Buy/sell orders use Binance Market orders (`OrderTypeMarket`) via the `go-binance/v2` client. The `executeBuyOrder` and `executeSellOrder` functions in `internal/bot/bot.go` need rigorous testing before live trading (see TODO comments in code).

## Development Guidelines

### Phased Development Plan
The README.md contains a detailed phased growth plan (PHASE 2-7) for adding features incrementally:
- Phase 2: Current state (multi-file structure)
- Phase 3: Add database (SQLite)
- Phase 4: Real trading (currently on testnet)
- Phase 5: Web interface
- Phase 6: Strategy pattern (multiple strategies)
- Phase 7: Vue.js frontend

**Critical Rule**: Only advance phases when previous phase runs stable for 24+ hours without crashes.

### Safety Considerations
- Always test with `trading_enabled: false` (paper trading) first
- Bot currently configured for Binance Testnet
- RSI signals should be verified against independent sources before trusting
- Small quantities recommended for initial live testing
- No stop-loss or take-profit limits currently implemented

### Module Structure
The project uses a single Go module (`rsi-bot`) at the root. The `trading-bot-ui` directory has its own `go.mod` and is a separate Wails application that will eventually integrate with the trading bot logic.

## Known Limitations & TODOs

1. **Order Execution**: Buy/sell functions marked for rigorous testing (see `internal/bot/bot.go:287-322`)
2. **Wails UI**: Currently minimal boilerplate, needs integration with bot logic
3. **No database**: Position history and trade logs not persisted
4. **Single strategy**: Only RSI implemented, no support for multiple strategies
5. **No risk management**: No stop-loss, take-profit, or position sizing logic

## Dependencies

**Core**:
- `gorilla/websocket` - WebSocket client
- `adshao/go-binance/v2` - Binance API client
- `spf13/viper` - Configuration management
- `joho/godotenv` - Environment variable loading

**UI**:
- `wailsapp/wails/v2` - Desktop application framework
- Vue.js frontend (in `trading-bot-ui/frontend`)
