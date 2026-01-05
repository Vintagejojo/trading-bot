# DCA Strategy with Email Notifications & Buy-the-Dip

## 🎯 What's New

Your trading bot now has THREE killer features:

1. **Email Notifications** - Get notified after every trade with full portfolio stats
2. **Buy-the-Dip** - Automatically buy extra when Bitcoin drops ≥5%
3. **Weekly Summaries** - Coming soon! (infrastructure ready)

---

## 📧 Step 1: Set Up Email Notifications

### Option A: Gmail (Recommended)

1. **Enable 2-Factor Authentication** on your Gmail account
   - Go to https://myaccount.google.com/security
   - Enable 2-Step Verification

2. **Generate App Password**
   - Go to https://myaccount.google.com/apppasswords
   - Select "Mail" and your device
   - Copy the 16-character password (e.g., `abcd efgh ijkl mnop`)

3. **Update your `.env` file:**

```bash
# Email Notifications
EMAIL_NOTIFICATIONS_ENABLED=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_FROM_EMAIL=your.email@gmail.com
SMTP_PASSWORD=abcdefghijklmnop  # Your 16-char app password (no spaces!)
NOTIFICATION_EMAIL=client.email@example.com  # Where to send alerts
```

### Option B: Other Email Providers

**Outlook/Hotmail:**
```bash
SMTP_HOST=smtp-mail.outlook.com
SMTP_PORT=587
SMTP_FROM_EMAIL=your.email@outlook.com
SMTP_PASSWORD=your_password
```

**Yahoo:**
```bash
SMTP_HOST=smtp.mail.yahoo.com
SMTP_PORT=587
SMTP_FROM_EMAIL=your.email@yahoo.com
SMTP_PASSWORD=your_password
```

---

## 🚀 Step 2: Configure DCA Strategy

### Basic DCA (Weekly Buys Only)

**File:** `configs/config-dca.yaml`

```yaml
symbol: "BTCUSDT"
quantity: 100.0  # $100 per purchase
trading_enabled: true

strategy:
  type: "dca"
  indicator:
    type: "dca"
    params:
      day_of_week: 1  # Monday
      hour_of_day: 9  # 9am UTC
```

### Advanced DCA (With Buy-the-Dip)

```yaml
symbol: "BTCUSDT"
quantity: 100.0
trading_enabled: true

strategy:
  type: "dca"
  indicator:
    type: "dca"
    params:
      day_of_week: 1
      hour_of_day: 9

      # Buy-the-Dip Magic! 🎯
      buy_the_dip: true
      dip_threshold: 5.0      # Buy extra when BTC drops ≥5%
      dip_multiplier: 1.5     # Buy 1.5x normal amount on dips
```

**What this does:**
- Regular buy: $100 every Monday at 9am
- Dip buy: Extra $150 when BTC drops 5%+ in 24h
- Example: BTC at $100k drops to $95k → automatic $150 buy!

---

## 📬 Step 3: Test Email Notifications

### Quick Test (Don't Wait Until Monday!)

1. **Edit `configs/config-dca.yaml`** temporarily:

```yaml
params:
  day_of_week: 3  # Today's day (0=Sun, 1=Mon, ... 6=Sat)
  hour_of_day: 15  # Current hour + 1 (e.g., if it's 2pm, use 15)
```

2. **Start the bot:**
```bash
cd trading-bot-ui
wails dev
```

3. **Select DCA strategy** from dropdown
4. **Click Start Bot**
5. **Wait ~1 hour** - you'll get an email when it buys!

### What the Email Looks Like

```
Subject: ✅ DCA Purchase: $100.00 → 0.00107 BTC

Your automated Bitcoin purchase executed successfully!

💰 Purchase Details:
   Amount Invested: $100.00
   Bitcoin Price: $93,450.00
   BTC Purchased: 0.00107021 BTC

📊 Portfolio Update:
   Total Holdings: 0.05234 BTC
   Current Value: $4,890.23
   Average Cost: $89,234.50/BTC
   📈 Unrealized Gain: $220.45 (+4.7%)

⏰ Next Purchase:
   Monday, December 16 at 9:00 AM

---
Powered by Tradecraft 🤖
Intelligent Bitcoin Accumulation
```

### Dip Buy Email

```
Subject: 🎯 DIP BUY (5.3% down): $150.00 → 0.00161 BTC

Bitcoin dropped 5.3% - buying extra!

[Same format as above, but with dip indicator]
```

---

## 🧪 Step 4: Test Buy-the-Dip

### Simulate a Dip (For Testing)

The buy-the-dip logic triggers when:
1. Price drops ≥5% from 24h high
2. Haven't bought a dip in last 24h

**To test without waiting for real dip:**

Option 1: Lower the threshold temporarily
```yaml
dip_threshold: 0.5  # Trigger on any -0.5% move
```

Option 2: Wait for natural market volatility (crypto moves fast!)

---

## 📊 Step 5: Monitor Your DCA

### Check Email After Each Buy

You'll receive emails for:
- ✅ Scheduled weekly buys
- 🎯 Automatic dip buys
- 📊 Weekly summaries (coming soon)

### View in UI

- **Trade History**: See all purchases
- **Performance Stats**: Track total ROI
- **Current Position**: See holdings and unrealized gains

---

## 🛠️ Troubleshooting

### Email Not Sending?

**Check 1: Test email credentials**
```bash
# Add this to pkg/notifications/email.go temporarily:
func (e *EmailNotifier) TestConnection() error {
    return e.sendEmail("Test Email", "If you get this, it works!")
}
```

**Check 2: Check logs**
```bash
# Look for these logs:
📧 Email sent: ✅ DCA Purchase...  # Success
❌ Failed to send email: ...       # Error (check SMTP settings)
```

**Common Issues:**
- Gmail: Make sure you used App Password, not regular password
- Firewall: Make sure port 587 is open
- Wrong email: Check SMTP_FROM_EMAIL matches your provider

### Dip Buy Not Triggering?

**Check:**
1. `buy_the_dip: true` in config?
2. Price actually dropped ≥5% from 24h high?
3. No dip buy in last 24h? (prevents multiple per day)

**Debug logs:**
```bash
# You'll see:
🎯 Dip detected! 5.3% down from 24h high
```

### Bot Not Starting?

**Check:**
1. `BINANCE_API_KEY` and `BINANCE_API_SECRET` in `.env`?
2. `EMAIL_NOTIFICATIONS_ENABLED=true`?
3. All email fields filled in `.env`?

---

## 💡 Pro Tips

### Optimize Your Schedule

**For US users (EST/PST):**
- UTC 9am = 4am EST / 1am PST (early morning)
- UTC 14:00 = 9am EST / 6am PST (morning)
- UTC 17:00 = 12pm EST / 9am PST (lunch)

**Best times:**
- Monday 9am UTC: Start of trading week
- Wednesday 14:00 UTC: Mid-week (avoid weekend volatility)

### Adjust Dip Settings

**Conservative (fewer dip buys):**
```yaml
dip_threshold: 10.0     # Only on big -10% crashes
dip_multiplier: 2.0     # Buy 2x on major dips
```

**Aggressive (more dip buys):**
```yaml
dip_threshold: 3.0      # Buy on -3% dips
dip_multiplier: 1.2     # Just 20% extra
```

**Whale Mode (big money):**
```yaml
quantity: 1000.0        # $1000 weekly
dip_threshold: 5.0
dip_multiplier: 3.0     # $3000 on dips!
```

### Tax Tracking

Your email receipts serve as:
- ✅ Purchase confirmations
- ✅ Cost basis records
- ✅ Trade history for taxes

Save them in a folder: "Bitcoin DCA 2025"

---

## 🎓 Understanding the Stats

### Email Metrics Explained

**Total Holdings:** How much BTC you own
**Current Value:** Holdings × current price
**Average Cost:** Your cost basis (for taxes)
**Unrealized Gain:** Paper profit (not sold yet)
**Unrealized ROI:** Percentage gain

**Example:**
```
Total Invested: $1,000
Current Value: $1,200
Unrealized Gain: $200 (20% ROI)
```

If you sell now → $200 profit (minus fees)
If you hold → who knows! 🚀

---

## 🚀 Next Steps

1. **Run for 4 weeks** with small amounts ($10-50)
2. **Verify emails work** perfectly
3. **Check trade history** matches emails
4. **Increase quantity** when confident

**Coming Soon:**
- Weekly summary emails (portfolio performance)
- Multi-asset DCA (BTC + ETH + SOL)
- Tax report generation
- Mobile push notifications

---

## 📞 Need Help?

Check logs for errors:
```bash
# Look for these:
📧 Email sent: ...
🎯 Dip detected: ...
✅ DCA Purchase: ...
❌ Failed to send email: ...
```

**Common questions:**
- "How do I test without real money?" → Use `trading_enabled: false` (paper trading)
- "Can I DCA multiple coins?" → Not yet, coming in Phase 8
- "How do I stop it?" → Click "Stop Bot" in UI

---

**You're all set! Your bot will now:**
1. ✅ Buy $100 BTC every Monday at 9am
2. 🎯 Buy extra $150 when BTC dips ≥5%
3. 📧 Email you after every purchase
4. 📊 Track your portfolio and ROI

Happy stacking! 🚀
