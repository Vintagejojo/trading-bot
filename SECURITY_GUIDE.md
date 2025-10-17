# Security Guide - Trading Bot

## Overview

Your trading bot implements **multi-layer security** to protect against unauthorized access and ensure safe operation.

---

## Security Layers

### 1. **Operating System Protection** (Default)

**File System Permissions**:
- `.env` file: Contains API keys, only readable by your OS user
- `trading_bot.db`: Database file, protected by OS file permissions
- Configuration files: Secured by OS user account

**Effectiveness**:
- ✅ Prevents other users on the same computer from accessing files
- ✅ No additional configuration needed
- ✅ Works on Windows, macOS, and Linux

**Recommendation**: This is sufficient for single-user machines.

---

### 2. **Optional PIN Protection** (NEW!)

**When to Use**:
- Shared computers
- Extra security layer for peace of mind
- Protection against someone using your unlocked desktop

**How It Works**:

1. **First Launch**: App asks if you want to set a PIN (can skip)
2. **With PIN**: Every app launch requires PIN entry
3. **Secure Storage**: PIN is hashed with SHA-256, stored in:
   - Windows: `C:\Users\<username>\AppData\Roaming\trading-bot\auth.pin`
   - macOS: `~/Library/Application Support/trading-bot/auth.pin`
   - Linux: `~/.config/trading-bot/auth.pin`

**PIN Features**:
- ✅ Minimum 4 characters
- ✅ SHA-256 hashing (never stored in plain text)
- ✅ File permissions: `0600` (owner read/write only)
- ✅ Can be changed or removed anytime
- ✅ Optional - can skip on first run

---

## Using PIN Protection

### Setting Up PIN (First Time)

1. Launch the app
2. See "Set PIN" screen
3. Enter desired PIN (min 4 characters)
4. Confirm PIN
5. Click "Set PIN" or "Skip"

### Unlocking App

1. Launch the app
2. Enter your PIN
3. Click "Unlock"

### Managing PIN

**From GUI** (Future enhancement):
- Settings → Security → Change PIN
- Settings → Security → Remove PIN

**From Code** (Currently):
```javascript
// Change PIN
await ChangePIN(oldPIN, newPIN)

// Remove PIN
await RemovePIN()
```

### Forgot PIN?

**Option 1**: Delete the PIN file
```bash
# Windows
del "%APPDATA%\trading-bot\auth.pin"

# macOS/Linux
rm ~/.config/trading-bot/auth.pin
```

**Option 2**: Reinstall app (preserves database and trades)

---

## API Key Security

### Best Practices

1. **Use Binance API Restrictions**:
   - Enable "Enable Withdrawals" = **OFF**
   - Enable "Enable Trading" = **ON** (if needed)
   - Restrict to your IP address (optional)
   - Set withdrawal whitelist

2. **Environment Variables** (`.env`):
   ```
   BINANCE_API_KEY=your_key_here
   BINANCE_API_SECRET=your_secret_here
   ```

3. **File Permissions**:
   ```bash
   # Make .env readable only by you
   chmod 600 .env
   ```

4. **Never Commit to Git**:
   - `.env` is in `.gitignore`
   - Never share API keys publicly

---

## Desktop App Security

### What's Protected

✅ **Database Access**: SQLite file only accessible by your OS user
✅ **API Keys**: Stored in `.env`, not hardcoded
✅ **PIN Authentication**: Optional extra layer
✅ **Auto-Lock**: Stops bot when app closes

### What's NOT Protected

❌ **Process Memory**: API keys are in memory while bot runs
❌ **Network Traffic**: HTTPS, but logs might contain data
❌ **Screen Recording**: No protection against screen capture

---

## Threat Model

### Protected Against

✅ **Other users on shared computer** → OS file permissions + PIN
✅ **Unauthorized app access** → PIN protection
✅ **Accidental changes** → Confirmation dialogs
✅ **API key exposure in code** → Environment variables

### NOT Protected Against

❌ **Malware on your computer** → OS-level compromise
❌ **Physical access to unlocked machine** → Use PIN + lock screen
❌ **Stolen database file** → Trade history is readable (no sensitive API data)
❌ **Network monitoring** → Use VPN if concerned

---

## Recommendations by Use Case

### Home PC (Solo User)
- **Minimal**: OS file permissions (default)
- **Optional**: Set PIN if you want extra protection
- **No additional configuration needed**

### Shared Computer
- **Required**: Set PIN on first launch
- **Recommended**: Lock Windows/macOS when away
- **Optional**: Use separate Windows user account

### Public/Untrusted Environment
- **NOT RECOMMENDED**: Don't run trading bot on public computers
- **Alternative**: Use VPS/cloud server instead

---

## Advanced Security (Optional)

### 1. Encrypt Database at Rest

```bash
# Install SQLCipher (encrypted SQLite)
go get github.com/mutecomm/go-sqlcipher/v4

# Update database.go to use sqlcipher driver
```

### 2. Hardware Security Key

- Use YubiKey or similar for 2FA
- Requires additional implementation

### 3. Network Restrictions

```go
// In app.go, add IP whitelist check
func (a *App) StartBot(...) error {
    if !isAllowedIP() {
        return fmt.Errorf("unauthorized network")
    }
    // ... rest of code
}
```

### 4. Audit Logging

```go
// Log all sensitive operations
log.Printf("User unlocked app from IP: %s", getIP())
log.Printf("Bot started with strategy: %s", strategy)
```

---

## Security Checklist

### Before First Run
- [ ] Set up `.env` file with API keys
- [ ] Set Binance API restrictions (disable withdrawals!)
- [ ] Decide if you want PIN protection
- [ ] Review file permissions on `.env`

### Ongoing
- [ ] Keep API keys secure (never share)
- [ ] Lock computer when away from desk
- [ ] Regularly review trade history for anomalies
- [ ] Update Binance API key periodically
- [ ] Monitor bot logs for suspicious activity

### If Compromised
- [ ] Immediately disable Binance API key
- [ ] Generate new API key with restrictions
- [ ] Review trade history for unauthorized trades
- [ ] Change PIN (if using)
- [ ] Check for malware on computer

---

## PIN Implementation Details

### Hash Algorithm
- SHA-256 (one-way cryptographic hash)
- No salt (acceptable for desktop app, not web service)
- 64-character hex output

### File Storage
```
~/.config/trading-bot/auth.pin
File permissions: 0600 (rw-------)
Contents: SHA-256 hash only (not plaintext PIN)
```

### Security Properties
- ✅ Cannot reverse hash to get original PIN
- ✅ File only readable by your user account
- ✅ No network transmission
- ✅ Wiped from memory after verification

### Limitations
- ⚠️ No salt means rainbow table attacks possible (mitigated by file permissions)
- ⚠️ No rate limiting on PIN attempts (desktop app, not public-facing)
- ⚠️ No account lockout after failed attempts

---

## Comparison: PIN vs No PIN

| Feature | No PIN | With PIN |
|---------|--------|----------|
| Protection from other OS users | ✅ | ✅ |
| Protection if desktop unlocked | ❌ | ✅ |
| Convenience | ✅ Best | ⚠️ Enter PIN each launch |
| Setup complexity | ✅ None | ⚠️ Set PIN once |
| Security level | Good | Better |
| Recovery if forgotten | N/A | Delete PIN file |

**Recommendation**:
- Home PC, solo user: **No PIN needed**
- Shared computer: **Use PIN**
- High-value trading: **Use PIN + VPS**

---

## Future Enhancements

Potential security improvements:

1. **Biometric Authentication** (fingerprint, Face ID)
2. **2FA with TOTP** (Google Authenticator)
3. **Session Timeout** (auto-lock after inactivity)
4. **Encrypted Backups** (GPG/PGP encrypted database exports)
5. **Audit Trail** (detailed log of all actions with timestamps)
6. **Remote Lock** (disable bot remotely via API)

---

## Questions?

### "Is PIN encryption strong enough?"

For a desktop app:
- **Yes**, if your computer itself is secure
- **No**, if you have malware or untrusted users with admin access

PIN is designed to prevent:
- Casual access by someone using your unlocked computer
- Unauthorized changes if you step away

It's **not** designed to protect against:
- Malware with admin privileges
- Professional attackers with physical access

### "Should I use PIN or not?"

**Use PIN if**:
- You share your computer
- You handle large amounts
- You want extra peace of mind
- You work in public spaces

**Skip PIN if**:
- Solo user on personal computer
- Already using full-disk encryption
- Want maximum convenience
- Low-value test trading

### "Can I add stronger authentication?"

Yes! The architecture supports:
- Hardware keys (YubiKey)
- Biometric (Touch ID, Windows Hello)
- Time-based OTP (Google Authenticator)

These would require additional implementation.

---

## Summary

Your trading bot has **good default security**:

1. **OS-level protection** (file permissions) ✅
2. **API keys in environment variables** ✅
3. **Optional PIN protection** ✅ (NEW!)
4. **Confirmation dialogs** ✅
5. **No hardcoded secrets** ✅

**For most users**: Default security is sufficient.

**For shared computers**: Enable PIN protection.

**For high-value trading**: Consider VPS + PIN + API restrictions.
