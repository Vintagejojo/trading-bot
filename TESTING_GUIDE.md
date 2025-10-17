# Testing Guide - Trading Bot

This guide will walk you through testing all features of the trading bot application.

---

## Prerequisites

Before testing, ensure you have:

- [ ] Go 1.21+ installed
- [ ] Node.js 16+ and npm installed
- [ ] Wails CLI installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- [ ] Git Bash or similar terminal (for Windows)

---

## Step 1: Install Frontend Dependencies

First, let's install all required frontend dependencies including Tailwind CSS.

```bash
# Navigate to frontend directory
cd trading-bot-ui/frontend

# Install all dependencies
npm install

# Install Tailwind CSS and related packages
npm install -D tailwindcss postcss autoprefixer

# Initialize Tailwind
npx tailwindcss init -p
```

This creates:
- `tailwind.config.js` - Tailwind configuration
- `postcss.config.js` - PostCSS configuration

---

## Step 2: Configure Tailwind CSS

### 2.1 Update `tailwind.config.js`

Edit `trading-bot-ui/frontend/tailwind.config.js`:

```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
```

### 2.2 Create `src/style.css`

Create `trading-bot-ui/frontend/src/style.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

### 2.3 Import styles in `main.js`

Edit `trading-bot-ui/frontend/src/main.js` and add at the top:

```javascript
import './style.css'
```

---

## Step 3: Generate Wails Bindings

Return to the Wails project root and generate TypeScript/JavaScript bindings:

```bash
cd ..  # Back to trading-bot-ui directory
wails generate module
```

This creates bindings in `frontend/wailsjs/` for all your Go backend functions.

---

## Step 4: Start Development Server

Start the Wails development server with hot-reload:

```bash
wails dev
```

**What this does:**
- Compiles Go backend
- Starts Vue.js frontend with Vite
- Opens application window
- Enables hot-reload (changes auto-update)

**First launch:** The app should show the Setup Wizard since no API keys are configured yet.

---

## Test Sequence 1: Fresh Installation (No Setup)

### 4.1 Test Setup Wizard

**Expected behavior:**
1. ✅ App opens to Setup Wizard screen
2. ✅ Shows welcome message with rocket emoji
3. ✅ "Show instructions" button visible

**Test the instructions panel:**
1. Click through the instructions
2. Verify all steps are visible:
   - Log in to Binance
   - Go to API Management
   - Create API
   - Complete 2FA
   - Copy keys
3. Security warnings displayed (disable withdrawals)
4. Testnet link visible

**Test the form:**
1. Click "Got it, continue to setup"
2. Form appears with two input fields
3. Try submitting empty form → Should show validation error
4. Enter short API key (< 20 chars) → Should show error
5. Enter valid-looking API key and secret:
   ```
   API Key: test1234567890123456789
   Secret:  secret1234567890123456789
   ```
6. Click "Save API Keys"
7. Should show success message
8. After 1 second, should transition to next screen

**Verify backend:**
```bash
# Check if .env file was created
cat ~/.config/trading-bot/.env

# Should contain:
# BINANCE_API_KEY=test1234567890123456789
# BINANCE_API_SECRET=secret1234567890123456789
```

### 4.2 Test PIN Setup (After Setup Wizard)

**Expected behavior:**
1. ✅ After setup wizard completes, PIN lock screen appears
2. ✅ Shows "Set Up PIN Protection" header
3. ✅ Two input fields: PIN and Confirm PIN

**Test PIN creation:**
1. Enter short PIN (< 4 chars) → Should show error
2. Enter mismatched PINs → Should show error
3. Enter matching PINs (e.g., "1234")
4. Click "Set PIN"
5. Should show success and lock app

**Test skipping PIN:**
1. Refresh app or close and reopen
2. On PIN setup screen, click "Skip"
3. Should go directly to main app
4. Close and reopen app
5. Should go directly to main app (no PIN lock)

**Verify PIN file:**
```bash
# If you set a PIN, check if hash was saved
cat ~/.config/trading-bot/auth.pin

# Should contain SHA-256 hash (64 hex characters)
```

### 4.3 Test PIN Lock (If PIN Was Set)

**Close and reopen the app:**
1. ✅ Should show PIN lock screen immediately
2. ✅ Shows "Enter PIN to Unlock Trading Bot"

**Test unlock:**
1. Enter wrong PIN → Should show error
2. Clear and try again
3. Enter correct PIN (e.g., "1234")
4. Click "Unlock"
5. Should unlock and show main application

---

## Test Sequence 2: Main Application Features

Once you're past setup and PIN lock, you should see the main trading interface.

### 5.1 Test Bot Controls

**Verify UI elements:**
- ✅ Strategy dropdown (RSI, MACD, Bollinger Bands)
- ✅ Symbol input field
- ✅ Quantity input field
- ✅ Dynamic parameter fields
- ✅ Paper Trading toggle (should be ON by default)
- ✅ START BOT button (green)

**Test strategy parameter loading:**
1. Select "RSI" from dropdown
2. Should show: Period, Overbought Level, Oversold Level
3. Select "MACD" from dropdown
4. Should show: Fast Period, Slow Period, Signal Period
5. Select "Bollinger Bands" from dropdown
6. Should show: Period, Std Deviation Multiplier

**Test validation:**
1. Leave symbol empty → Try to start bot → Should show error
2. Enter invalid symbol (e.g., "FAKECOIN")
3. Enter quantity: 1000
4. Click START BOT

**Expected behavior:**
- If using test API keys: Will fail to connect to Binance (expected)
- If using real testnet keys: Should connect successfully

### 5.2 Test Paper Trading Toggle

1. Toggle "Paper Trading" OFF
2. Should show warning dialog about live trading
3. Confirm or cancel
4. Toggle should reflect state

### 5.3 Test Performance Stats

**Verify card displays:**
- Total Profit/Loss: $0.00 (0.00%)
- Win Rate: 0%
- Total Trades: 0
- Average P/L: $0.00
- Largest Win: $0.00
- Largest Loss: $0.00

**After bot runs and makes trades**, these should update.

### 5.4 Test Current Position

**Initial state:**
- Should show "No open position"

**After bot buys:**
- Should show position details
- Entry price, quantity, strategy
- Time held (updates in real-time)
- Profit/loss (live updates)

### 5.5 Test Trade History

**Initial state:**
- Empty table or message: "No trades yet"

**After trades execute:**
- Should show rows with:
  - 🟢 BUY or 🔴 SELL icon
  - Symbol, price, quantity
  - Profit/loss (for sells)
  - Strategy name
  - Indicator value
  - Timestamp
  - Mode badge (📝 PAPER or 🔴 LIVE)

### 5.6 Test Auto-Refresh

The app auto-refreshes every 5 seconds.

**To verify:**
1. Start the bot
2. Watch the browser console (F12)
3. Should see refresh logs every 5 seconds
4. Trade history should update automatically
5. Stats should update automatically

---

## Test Sequence 3: Bot Trading Logic

### 6.1 Test with Paper Trading (Safe)

**Setup:**
1. Ensure Paper Trading is ON
2. Select RSI strategy
3. Symbol: SHIBUSDT (or any active Binance pair)
4. Quantity: 10000
5. RSI parameters: Period 14, Overbought 70, Oversold 30

**Start the bot:**
```bash
# You should see WebSocket connection in terminal
# Bot will connect to Binance and start receiving price data
```

**Monitor the terminal output:**
- Should see RSI calculations
- Should see "RSI: 45.2" type logs
- When RSI <= 30: Should execute BUY (paper trade)
- When RSI >= 70: Should execute SELL (paper trade)

**Check database:**
```bash
# After some trades, check the database
sqlite3 trading-bot-ui/trading_bot.db

# Run queries
SELECT * FROM trades;
SELECT * FROM positions;
```

### 6.2 Test Stop Bot

1. Click "STOP BOT" button
2. Should show confirmation dialog
3. Confirm stop
4. Bot status should change to "Stopped"
5. Running indicator should turn gray
6. WebSocket connection should close (check terminal)

---

## Test Sequence 4: Edge Cases

### 7.1 Test Network Disconnection

1. Start bot
2. Disconnect internet
3. Bot should attempt reconnection
4. Check terminal for reconnection logs
5. Reconnect internet
6. Bot should resume automatically

### 7.2 Test Invalid API Keys

1. Stop the bot
2. Edit `~/.config/trading-bot/.env`
3. Change API key to invalid value
4. Try to start bot
5. Should show error message

### 7.3 Test Multiple Starts

1. Try clicking START BOT multiple times rapidly
2. Should show error: "bot is already running"
3. Only one instance should run

### 7.4 Test App Close While Bot Running

1. Start bot
2. Close app window
3. Reopen app
4. Bot should be stopped (not auto-resume)

---

## Test Sequence 5: Real Trading (CAUTION!)

⚠️ **ONLY proceed if you:**
- Have tested thoroughly with paper trading
- Are using Binance Testnet (https://testnet.binance.vision)
- Have created testnet API keys
- Are comfortable risking the test funds

### 8.1 Create Testnet Account

1. Go to https://testnet.binance.vision
2. Register for free testnet account
3. Get free test BNB from faucet
4. Create API keys:
   - Enable Trading: YES
   - Enable Withdrawals: NO
   - IP restriction: Optional

### 8.2 Update API Keys

**Option 1: Edit .env file directly**
```bash
nano ~/.config/trading-bot/.env
# Replace with testnet keys
```

**Option 2: Delete setup and redo**
```bash
rm ~/.config/trading-bot/.env
# Restart app, Setup Wizard will appear
# Enter testnet keys
```

### 8.3 Test Live Trading on Testnet

1. Turn OFF "Paper Trading"
2. Confirm the warning
3. Start bot with small quantity
4. Monitor trades in:
   - App trade history
   - Binance Testnet account
5. Verify trades actually execute on Binance

---

## Common Issues & Solutions

### Issue: "command not found: wails"

**Solution:**
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Add to PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: "npm install" fails

**Solution:**
```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

### Issue: Tailwind styles not applying

**Solution:**
```bash
# Verify tailwind.config.js has correct content paths
# Verify style.css is imported in main.js
# Restart wails dev
```

### Issue: "Cannot find module './wailsjs/go/main/App'"

**Solution:**
```bash
# Generate bindings
wails generate module

# Restart wails dev
```

### Issue: WebSocket connection fails

**Solution:**
1. Check internet connection
2. Verify symbol exists on Binance (e.g., BTCUSDT)
3. Try different WebSocket endpoint
4. Check firewall/antivirus

### Issue: Bot doesn't execute trades

**Solution:**
1. Check if Paper Trading is ON (no real trades)
2. Check if indicator values trigger signals (RSI <= 30 or >= 70)
3. Check terminal logs for errors
4. Verify API keys are valid

### Issue: Database errors

**Solution:**
```bash
# Check if database exists
ls trading-bot-ui/trading_bot.db

# If corrupted, delete and restart bot (will recreate)
rm trading-bot-ui/trading_bot.db
```

---

## Build for Production

After testing, build the production executable:

```bash
cd trading-bot-ui
wails build
```

**Output locations:**
- Windows: `build/bin/trading-bot.exe`
- macOS: `build/bin/trading-bot.app`
- Linux: `build/bin/trading-bot`

**Test the production build:**
1. Run the executable
2. Complete setup wizard
3. Verify all features work
4. Check file locations for .env and auth.pin

---

## Testing Checklist

Use this checklist to ensure complete testing:

### Setup Flow
- [ ] Setup Wizard appears on first launch
- [ ] Instructions panel displays correctly
- [ ] Form validation works (empty, too short, invalid)
- [ ] API keys save to correct location
- [ ] Transition to PIN setup after wizard completes

### PIN Protection
- [ ] PIN setup screen appears after API setup
- [ ] Can set PIN with validation
- [ ] Can skip PIN setup
- [ ] PIN lock appears on app relaunch (if PIN set)
- [ ] Correct PIN unlocks app
- [ ] Wrong PIN shows error
- [ ] Forgot PIN recovery works (delete auth.pin)

### Main Application
- [ ] App loads after unlocking
- [ ] All three panels visible (controls, stats, position)
- [ ] Strategy dropdown works
- [ ] Parameter fields change based on strategy
- [ ] Can enter symbol and quantity
- [ ] Paper Trading toggle works
- [ ] START BOT button works
- [ ] STOP BOT button works

### Bot Functionality
- [ ] WebSocket connects to Binance
- [ ] Receives real-time price data
- [ ] RSI calculates correctly
- [ ] Buy signal triggers at oversold level
- [ ] Sell signal triggers at overbought level
- [ ] Trades logged to database
- [ ] Trade history displays in table
- [ ] Stats update after trades
- [ ] Current position shows when holding

### Database
- [ ] Database file created on first trade
- [ ] Trades table populated
- [ ] Positions table populated
- [ ] Can query database directly
- [ ] App loads historical trades on restart

### Auto-Refresh
- [ ] Data refreshes every 5 seconds
- [ ] Event listeners work (bot:started, bot:stopped)
- [ ] UI updates automatically
- [ ] No memory leaks after extended running

### Error Handling
- [ ] Invalid API keys show error
- [ ] Invalid symbol shows error
- [ ] Network disconnection handled gracefully
- [ ] Bot auto-reconnects after network restored
- [ ] Database errors show user-friendly messages

### Production Build
- [ ] Build completes without errors
- [ ] Executable runs standalone
- [ ] Setup wizard works in built version
- [ ] API keys and PIN save correctly
- [ ] Trading functionality works in production build

---

## Performance Testing

### Memory Usage

Monitor the app's memory usage:

```bash
# While app is running, check memory
# Windows Task Manager / macOS Activity Monitor / Linux htop

# Should stay under 200MB for normal operation
```

### CPU Usage

- Idle (bot stopped): < 1% CPU
- Running (bot active): 2-5% CPU
- Spikes during trades: 10-20% CPU (brief)

### WebSocket Stability

Run the bot for extended periods:
- 1 hour: Should stay connected
- 24 hours: Should handle reconnections
- 1 week: Should be stable with occasional reconnects

---

## Next Steps After Testing

Once all tests pass:

1. **Document any bugs found** → Create issues
2. **Optimize performance** → Profile slow operations
3. **Add missing features** → Settings page, manual trading, etc.
4. **Prepare for distribution** → Create installer, write user docs
5. **Set up CI/CD** → Automated builds and tests

---

## Summary

Complete testing process:

1. Install dependencies (Node, Tailwind)
2. Configure Tailwind CSS
3. Generate Wails bindings
4. Test Setup Wizard flow
5. Test PIN lock flow
6. Test main application features
7. Test bot trading logic (paper trading)
8. Test edge cases and error handling
9. Build production executable
10. Test production build

**Estimated testing time:** 2-4 hours for thorough testing

Good luck! 🚀
