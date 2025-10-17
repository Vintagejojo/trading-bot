# Trading Bot - Complete Setup Flow

This document describes the complete user experience flow from first launch to fully operational trading bot.

---

## Application Flow Diagram

```
┌─────────────────────────────────────────┐
│         First Launch                    │
└─────────────────┬───────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────┐
│  Check: IsSetupComplete()?              │
│  - Does .env file exist?                │
│  - Does it have valid API keys?         │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┐
        │                   │
        NO                 YES
        │                   │
        ▼                   ▼
┌────────────────┐   ┌─────────────────┐
│ Setup Wizard   │   │ Check: HasPIN() │
│                │   │ & IsLocked()?   │
└───────┬────────┘   └────────┬────────┘
        │                     │
        │              ┌──────┴──────┐
        │              │             │
        │            LOCKED      UNLOCKED
        │              │             │
        │              ▼             │
        │        ┌──────────┐        │
        │        │ PIN Lock │        │
        │        │  Screen  │        │
        │        └────┬─────┘        │
        │             │              │
        │          UNLOCK            │
        │             │              │
        ▼             ▼              ▼
    ┌────────────────────────────────┐
    │       Main Application         │
    │  - Bot Controls                │
    │  - Trade History               │
    │  - Performance Stats           │
    │  - Current Position            │
    └────────────────────────────────┘
```

---

## Flow 1: Brand New User (No Setup)

### Step 1: Setup Wizard Appears

**What the user sees:**
- Welcome screen with rocket emoji
- "Let's set up your Binance API connection"
- Instructions panel with step-by-step guide

**Instructions shown:**
1. Log in to Binance.com
2. Go to Profile → API Management
3. Click "Create API" and name it "Trading Bot"
4. Complete 2FA verification
5. Copy your API Key and Secret Key

**Security warnings displayed:**
- ⚠️ Enable "Enable Trading" ✓
- ⚠️ DISABLE "Enable Withdrawals" ✗ (for security!)
- ⚠️ Consider restricting to your IP address

**Testnet recommendation:**
- Link to https://testnet.binance.vision
- Suggestion to test with fake money first

**User actions:**
1. Click "Got it, continue to setup"
2. Enter API Key
3. Enter API Secret (masked, with "Show secret" checkbox)
4. Click "Save API Keys"

**What happens in backend:**
```go
SaveAPIKeys(apiKey, apiSecret string) error
├─ Validates format (min 20 chars, no spaces)
├─ Creates directory: ~/.config/trading-bot/
├─ Writes to file: ~/.config/trading-bot/.env
├─ Sets file permissions: 0600 (owner read/write only)
└─ Sets environment variables for current session
```

**Success:**
- Green checkmark: "✓ API keys saved successfully!"
- Emits `setup-complete` event
- Transitions to next screen (PIN setup or main app)

### Step 2: PIN Setup (Optional)

After API keys are saved, user sees PIN lock screen:

**First-time PIN setup:**
- "Set Up PIN Protection" header
- Explanation: "Add an extra layer of security..."
- PIN input field (minimum 4 characters)
- Confirmation input field
- "Set PIN" button
- "Skip (Not Recommended)" button

**If user sets PIN:**
- PIN hashed with SHA-256
- Saved to `~/.config/trading-bot/auth.pin` (0600 permissions)
- App locks on every launch
- User must enter PIN to access

**If user skips:**
- No PIN required on future launches
- Only OS-level file permissions protect data

### Step 3: Main Application

User enters the main trading interface (see below)

---

## Flow 2: Returning User (Setup Complete)

### With PIN Protection

**What happens:**
1. App checks `IsSetupComplete()` → TRUE
2. App checks `HasPIN()` → TRUE
3. App checks `IsLocked()` → TRUE (default on launch)
4. Shows PIN Lock screen

**PIN Lock Screen:**
- Lock icon
- "Enter PIN to Unlock Trading Bot"
- PIN input field (password-masked)
- "Unlock" button
- Show/hide PIN toggle

**After correct PIN:**
- Calls `UnlockApp(pin)`
- Sets `isLocked = false`
- Loads main application
- Refreshes bot status, trades, stats

**After incorrect PIN:**
- Shows error: "Invalid PIN"
- Clears input field
- User can try again (no lockout)

### Without PIN Protection

**What happens:**
1. App checks `IsSetupComplete()` → TRUE
2. App checks `HasPIN()` → FALSE
3. App checks `IsLocked()` → FALSE
4. Shows main application immediately

**No PIN screen shown** - user goes directly to trading interface

---

## Main Application Interface

Once setup is complete and user is unlocked, they see:

### Header Bar

```
┌────────────────────────────────────────────────────────┐
│ 🤖 Trading Bot    ● Running / ● Stopped                │
│                             SHIBUSDT    📝 PAPER / 🔴 LIVE │
└────────────────────────────────────────────────────────┘
```

### Three-Panel Layout

**Left Panel (1/3 width):**

1. **Bot Controls Card**
   - Strategy dropdown (RSI, MACD, Bollinger Bands)
   - Symbol input (e.g., SHIBUSDT, BTCUSDT)
   - Quantity input (number of coins)
   - Dynamic parameter inputs:
     - RSI: period, overbought, oversold
     - MACD: fast, slow, signal periods
     - BBands: period, std deviation
   - Paper Trading toggle (on by default)
   - START BOT / STOP BOT buttons

2. **Performance Stats Card**
   - Total Profit/Loss ($$$, %)
   - Win Rate (%)
   - Total Trades / Wins / Losses
   - Average P/L per trade
   - Largest Win / Loss

3. **Current Position Card**
   - Shows if holding position
   - Entry price, quantity
   - Current profit/loss
   - Time held
   - Strategy used

**Right Panel (2/3 width):**

**Trade History Table**
- Scrollable list of all trades
- Columns:
  - Type (🟢 BUY / 🔴 SELL)
  - Symbol
  - Price
  - Quantity
  - Profit/Loss ($$$, %)
  - Strategy
  - Indicator Value (RSI: 32.5, etc.)
  - Time
  - Mode (📝 PAPER / 🔴 LIVE)

---

## Auto-Refresh Behavior

**Every 5 seconds**, the app automatically refreshes:
- `GetBotStatus()` - Updates running status, position, last trade
- `GetTradeHistory(50)` - Gets latest 50 trades
- `GetTradeSummary()` - Updates aggregate statistics

**Real-time event listeners:**
- `bot:started` → Refresh all data
- `bot:stopped` → Refresh all data
- `bot:error` → Show alert dialog

---

## File Locations

### API Keys (Per User)
```
Windows: C:\Users\<username>\AppData\Roaming\trading-bot\.env
macOS:   /Users/<username>/Library/Application Support/trading-bot/.env
Linux:   /home/<username>/.config/trading-bot/.env
```

### PIN Hash (Per User)
```
Windows: C:\Users\<username>\AppData\Roaming\trading-bot\auth.pin
macOS:   /Users/<username>/Library/Application Support/trading-bot/auth.pin
Linux:   /home/<username>/.config/trading-bot/auth.pin
```

### Database (Shared - Project Directory)
```
<project-root>/trading_bot.db
```

**Note:** API keys and PIN are stored per-user, but the trading database is in the project directory and shared by all users of the same installation.

---

## Security Model

### Layer 1: OS File Permissions (Default)
- `.env` file: 0600 (owner read/write only)
- `auth.pin` file: 0600 (owner read/write only)
- Config directory: 0700 (owner read/write/execute only)

**Protection:**
- Other OS users cannot read your API keys
- Other OS users cannot read your PIN hash

**Limitations:**
- Administrator/root can still read files
- Malware running as your user can read files
- No protection if desktop is unlocked

### Layer 2: PIN Protection (Optional)
- PIN hashed with SHA-256
- Cannot reverse hash to get original PIN
- File only readable by your OS user
- Required on every app launch if enabled

**Protection:**
- Someone using your unlocked computer cannot access bot
- Cannot start/stop bot without PIN
- Cannot modify trades without PIN

**Limitations:**
- No protection against malware
- No brute-force rate limiting
- No account lockout after failed attempts

---

## Distribution to Other Users

### For Developers

**Your API keys remain private:**
- Never commit `.env` to git (already in `.gitignore`)
- Your local `.env` stays in `~/.config/trading-bot/.env`
- Not packaged with the app

**When you build and distribute:**
```bash
cd trading-bot-ui
wails build
```

The built executable (`trading-bot.exe`, `trading-bot.app`, `trading-bot`) contains:
- ✅ Go backend code
- ✅ Vue.js frontend
- ❌ No API keys
- ❌ No PIN hashes
- ❌ No user data

### For End Users

**First launch:**
1. Download and run the app
2. See Setup Wizard (no API keys detected)
3. Follow instructions to create Binance API keys
4. Enter their own API keys
5. Keys saved to their user config directory
6. Optionally set their own PIN

**Their data stays separate:**
- Each user has their own `~/.config/trading-bot/.env`
- Each user can have their own PIN
- Database is shared if using same installation (but each user's API keys are different)

---

## Recommended User Flow

### For Testing (Recommended First Steps)

1. **Use Binance Testnet:**
   - Go to https://testnet.binance.vision
   - Create testnet API keys (no real money)
   - Enter testnet keys in Setup Wizard

2. **Enable Paper Trading:**
   - Leave "Paper Trading" toggle ON
   - Bot will simulate trades without executing
   - Safe to test strategies

3. **Test with Small Amounts:**
   - Use small quantity values
   - Test RSI strategy first (simplest)
   - Monitor for 24 hours

### For Live Trading (After Testing)

1. **Create Live API Keys:**
   - Go to https://www.binance.com
   - Create API with restrictions:
     - ✓ Enable Trading
     - ✗ DISABLE Withdrawals
     - ✓ Restrict to your IP (optional)

2. **Update API Keys in App:**
   - Settings → Update API Keys (future feature)
   - Or manually edit `~/.config/trading-bot/.env`
   - Restart app

3. **Enable Live Trading:**
   - Turn OFF "Paper Trading" toggle
   - Confirm the warning dialog
   - Start bot with small quantity

4. **Monitor Closely:**
   - Watch first few trades carefully
   - Check Binance account to verify orders
   - Increase quantity gradually

---

## Troubleshooting

### "API Keys not found" error on startup

**Cause:** Setup not complete or .env file deleted

**Solution:**
1. Check if file exists: `~/.config/trading-bot/.env`
2. If missing, app will show Setup Wizard automatically
3. Re-enter API keys

### "Invalid PIN" but you're sure it's correct

**Cause:** PIN file corrupted

**Solution:**
```bash
# Delete PIN file (varies by OS)
rm ~/.config/trading-bot/auth.pin

# App will allow you to set new PIN on next launch
```

### "Failed to start bot" error

**Possible causes:**
1. Invalid API keys
2. Network connection issues
3. Binance API rate limits
4. Invalid symbol (e.g., "FAKECOIN" doesn't exist)

**Solution:**
1. Check Binance API key is active
2. Verify symbol exists on Binance (e.g., BTCUSDT)
3. Check network/firewall
4. Check bot logs for detailed error

### Want to reset everything

**Full reset:**
```bash
# Remove user config
rm -rf ~/.config/trading-bot/

# Remove database (optional - will lose trade history)
rm trading-bot-ui/trading_bot.db

# Next launch will show Setup Wizard
```

---

## Next Steps for Development

### Still TODO:

1. **Install Tailwind CSS:**
   ```bash
   cd frontend
   npm install -D tailwindcss postcss autoprefixer
   npx tailwindcss init -p
   ```

2. **Create Tailwind config:**
   ```javascript
   // tailwind.config.js
   module.exports = {
     content: ['./src/**/*.{vue,js,ts}'],
     theme: { extend: {} },
     plugins: [],
   }
   ```

3. **Create style.css:**
   ```css
   @tailwind base;
   @tailwind components;
   @tailwind utilities;
   ```

4. **Import in main.js:**
   ```javascript
   import './style.css'
   ```

5. **Generate Wails bindings:**
   ```bash
   wails generate module
   ```

6. **Build and test:**
   ```bash
   wails dev    # Development with hot-reload
   wails build  # Production build
   ```

---

## Summary

The complete setup flow is now implemented:

1. ✅ **Setup Wizard** - First-time API key configuration
2. ✅ **PIN Lock** - Optional security layer
3. ✅ **Main App** - Full trading interface

**Flow:**
```
Setup Wizard → PIN Setup (optional) → Main App
```

**Security:**
- OS-level file permissions (default)
- Optional PIN protection
- API keys never committed to code
- Each user has their own config

**Distribution:**
- Developers keep their API keys private
- End users enter their own API keys
- Clean separation of user data
- Easy to package and distribute

The app is now ready for Tailwind installation and testing!
