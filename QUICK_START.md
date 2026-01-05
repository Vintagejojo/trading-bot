# Quick Start Guide - Trading Bot

## Starting the Application

```bash
cd trading-bot-ui
wails dev
```

The application window will open automatically.

---

## What to Expect

### First Launch (No Setup)

You should see the **Setup Wizard** screen:

1. Welcome screen with instructions
2. Form to enter Binance API keys
3. Security recommendations

### Testing Without Real API Keys

For testing purposes, you can enter dummy API keys:

- **API Key**: `test1234567890123456789`
- **Secret**: `secret1234567890123456789`

**Note**: The bot won't connect to Binance with fake keys, but the UI will work.

---

## Common Issues & Fixes

### App Crashes When Clicking "Start Bot"

**Fixed!** The app now handles missing API credentials gracefully instead of crashing.

### Error: "bot not properly initialized: missing API credentials"

This is expected if you entered dummy API keys. The UI will still work, but the bot won't trade.

### To Reset Everything

Delete the setup files and restart:

```bash
# Windows
del %APPDATA%\trading-bot\.env
del %APPDATA%\trading-bot\auth.pin

# Linux/Mac
rm ~/.config/trading-bot/.env
rm ~/.config/trading-bot/auth.pin
```

---

## Testing Flow

### 1. Test Setup Wizard

1. App starts → Shows Setup Wizard
2. Enter API keys (real or fake)
3. Click "Save API Keys"
4. Should transition to PIN setup or main app

### 2. Test PIN Lock (Optional)

1. After setup, you may see PIN setup screen
2. Enter a 4+ character PIN
3. Confirm the PIN
4. Click "Set PIN" or "Skip"

###  3. Test Main Application

1. Should see the trading interface with 3 panels:
   - Bot Controls (left)
   - Performance Stats & Position (left)
   - Trade History (right)

2. Try selecting different strategies:
   - RSI
   - MACD
   - Bollinger Bands

3. Notice how parameters change based on strategy

4. Enter a symbol (e.g., "BTCUSDT")

5. Enter quantity (e.g., "1000")

6. Toggle Paper Trading ON/OFF

7. Click **START BOT**

**Expected behavior**:
- If using fake keys: Shows error "bot not properly initialized" but **doesn't crash**
- If using real testnet keys: Bot connects and starts trading

---

## Key Commands

**Start Wails**:
```bash
cd trading-bot-ui
wails dev
```

**Stop Wails**:
- Close the application window, or
- Press `Ctrl+C` in terminal

**View logs**:
The terminal shows all bot activity in real-time.

---

## Files Created by the App

### API Keys
```
Windows: %APPDATA%\trading-bot\.env
macOS:   ~/Library/Application Support/trading-bot/.env
Linux:   ~/.config/trading-bot/.env
```

### PIN Hash (if set)
```
Windows: %APPDATA%\trading-bot\auth.pin
macOS:   ~/Library/Application Support/trading-bot/auth.pin
Linux:   ~/.config/trading-bot/auth.pin
```

### Database
```
trading-bot-ui/trading_bot.db
```

---

## Next Steps

Once the basic UI is working, you can:

1. **Use Real Testnet Keys**:
   - Go to https://testnet.binance.vision
   - Create testnet API keys
   - Update keys in the app or in `.env` file

2. **Test Paper Trading**:
   - Keep "Paper Trading" ON
   - Start bot and watch it simulate trades

3. **Test Live Trading** (⚠️ Use testnet first!):
   - Turn OFF "Paper Trading"
   - Bot will execute real orders on testnet

4. **Review Trade History**:
   - After some trades, check the table
   - View profit/loss calculations
   - See indicator values at trade time

---

## Troubleshooting

### "Module not found" errors

```bash
cd trading-bot-ui
go mod tidy
```

### Frontend build errors

```bash
cd trading-bot-ui/frontend
npm install
```

### Wails command not found

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Make sure `$GOPATH/bin` is in your PATH.

### Database locked errors

Close the app completely and restart.

---

## Current Status

✅ **Fixed**: App no longer crashes when clicking "Start Bot" without API keys
✅ **Working**: Setup Wizard, PIN Lock, Main UI
✅ **Working**: Strategy selection, parameter customization
⚠️ **Requires Real Keys**: Actual trading needs valid Binance API keys

---

## For Full Testing Guide

See `TESTING_GUIDE.md` for comprehensive testing instructions.
