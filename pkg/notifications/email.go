package notifications

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"
)

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost           string
	SMTPPort           string
	FromEmail          string
	FromPassword       string
	ToEmail            string
	Enabled            bool
	NotifyOnDCABuy     bool // Send email for regular DCA purchases
	NotifyOnDipBuy     bool // Send email for buy-the-dip purchases
	SendMonthlySummary bool // Send monthly portfolio summaries
}

// EmailNotifier sends email notifications
type EmailNotifier struct {
	config EmailConfig
}

// NewEmailNotifier creates a new email notifier
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		config: config,
	}
}

// LoadFromEnv loads email config from environment variables
func LoadEmailConfigFromEnv() EmailConfig {
	return EmailConfig{
		SMTPHost:           getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		FromEmail:          os.Getenv("SMTP_FROM_EMAIL"),
		FromPassword:       os.Getenv("SMTP_PASSWORD"),
		ToEmail:            os.Getenv("NOTIFICATION_EMAIL"),
		Enabled:            os.Getenv("EMAIL_NOTIFICATIONS_ENABLED") == "true",
		NotifyOnDCABuy:     getEnv("NOTIFY_ON_DCA_BUY", "true") == "true",     // Default enabled
		NotifyOnDipBuy:     getEnv("NOTIFY_ON_DIP_BUY", "true") == "true",     // Default enabled
		SendMonthlySummary: getEnv("SEND_MONTHLY_SUMMARY", "true") == "true", // Default enabled
	}
}

// TradeNotification represents a trade event
type TradeNotification struct {
	Symbol          string
	Side            string
	Quantity        float64
	Price           float64
	Total           float64
	TotalHoldings   float64
	TotalValue      float64
	AverageCost     float64
	UnrealizedGain  float64
	UnrealizedROI   float64
	NextBuyTime     time.Time
	IsDipBuy        bool
	DipPercent      float64
}

// SendTradeNotification sends an email notification for a trade
func (e *EmailNotifier) SendTradeNotification(notification TradeNotification) error {
	if !e.config.Enabled {
		log.Println("📧 Email notifications disabled, skipping...")
		return nil
	}

	if e.config.ToEmail == "" {
		log.Println("⚠️  No notification email configured, skipping...")
		return nil
	}

	// Check if user wants emails for this type of buy
	if notification.Side == "BUY" {
		if notification.IsDipBuy && !e.config.NotifyOnDipBuy {
			log.Println("📧 Dip buy email notifications disabled, skipping...")
			return nil
		}
		if !notification.IsDipBuy && !e.config.NotifyOnDCABuy {
			log.Println("📧 Regular DCA email notifications disabled, skipping...")
			return nil
		}
	}

	var subject, body string

	if notification.Side == "BUY" {
		emoji := "✅"
		buyType := "DCA Purchase"

		if notification.IsDipBuy {
			emoji = "🎯"
			buyType = fmt.Sprintf("DIP BUY (%.1f%% down)", notification.DipPercent)
		}

		subject = fmt.Sprintf("%s %s: $%.2f → %.6f %s", emoji, buyType, notification.Total, notification.Quantity, notification.Symbol)

		gainEmoji := "📈"
		if notification.UnrealizedGain < 0 {
			gainEmoji = "📉"
		}

		body = fmt.Sprintf(`Your automated Bitcoin purchase executed successfully!

💰 Purchase Details:
   Amount Invested: $%.2f
   Bitcoin Price: $%.2f
   BTC Purchased: %.8f BTC

📊 Portfolio Update:
   Total Holdings: %.8f BTC
   Current Value: $%.2f
   Average Cost: $%.2f/BTC
   %s Unrealized Gain: $%.2f (%+.1f%%)

⏰ Next Purchase:
   %s

---
Powered by Tradecraft 🤖
Intelligent Bitcoin Accumulation
`,
			notification.Total,
			notification.Price,
			notification.Quantity,
			notification.TotalHoldings,
			notification.TotalValue,
			notification.AverageCost,
			gainEmoji,
			notification.UnrealizedGain,
			notification.UnrealizedROI,
			notification.NextBuyTime.Format("Monday, January 2 at 3:04 PM"),
		)
	} else {
		subject = fmt.Sprintf("💸 Sold %.6f %s at $%.2f", notification.Quantity, notification.Symbol, notification.Price)
		body = fmt.Sprintf("Sale notification details here...")
	}

	return e.sendEmail(subject, body)
}

// MonthlySummary represents monthly portfolio summary
type MonthlySummary struct {
	MonthOf         time.Time
	NumPurchases    int
	TotalInvested   float64
	BTCAccumulated  float64
	TotalHoldings   float64
	CurrentValue    float64
	TotalCost       float64
	ProfitLoss      float64
	ROI             float64
	AverageCost     float64
	CurrentPrice    float64
	BestBuyPrice    float64
	WorstBuyPrice   float64
	NextBuyTime     time.Time
	DipBuysEnabled  bool
}

// SendMonthlySummary sends monthly portfolio summary email
func (e *EmailNotifier) SendMonthlySummary(summary MonthlySummary) error {
	if !e.config.Enabled || e.config.ToEmail == "" {
		log.Println("📧 Email notifications disabled or no email configured, skipping monthly summary...")
		return nil
	}

	if !e.config.SendMonthlySummary {
		log.Println("📧 Monthly summary emails disabled, skipping...")
		return nil
	}

	roiEmoji := "📈"
	if summary.ROI < 0 {
		roiEmoji = "📉"
	}

	subject := fmt.Sprintf("%s Monthly Bitcoin Report: %+.1f%% ROI", roiEmoji, summary.ROI)

	bestBuyDiff := ((summary.AverageCost - summary.BestBuyPrice) / summary.AverageCost) * 100
	worstBuyDiff := ((summary.WorstBuyPrice - summary.AverageCost) / summary.AverageCost) * 100

	dipStatus := "No"
	if summary.DipBuysEnabled {
		dipStatus = "Yes (>5%)"
	}

	body := fmt.Sprintf(`YOUR BITCOIN ACCUMULATION - Month of %s

💰 This Month:
   Purchases: %d
   Total Invested: $%.2f
   BTC Acquired: %.8f BTC

📊 Portfolio:
   Total Holdings: %.8f BTC
   Current Value: $%.2f
   Total Invested: $%.2f
   %s Profit/Loss: $%.2f (%+.1f%%)

📈 Performance:
   Average Buy Price: $%.2f
   Current BTC Price: $%.2f
   Best Buy: $%.2f (%.1f%% below avg)
   Worst Buy: $%.2f (%+.1f%% vs avg)

🎯 Next Actions:
   Next regular buy: %s
   Watching for dips: %s

Keep stacking sats! 🚀

---
Powered by Tradecraft 🤖
`,
		summary.MonthOf.Format("January 2006"),
		summary.NumPurchases,
		summary.TotalInvested,
		summary.BTCAccumulated,
		summary.TotalHoldings,
		summary.CurrentValue,
		summary.TotalCost,
		roiEmoji,
		summary.ProfitLoss,
		summary.ROI,
		summary.AverageCost,
		summary.CurrentPrice,
		summary.BestBuyPrice,
		bestBuyDiff,
		summary.WorstBuyPrice,
		worstBuyDiff,
		summary.NextBuyTime.Format("Monday, Jan 2 at 3:04 PM"),
		dipStatus,
	)

	return e.sendEmail(subject, body)
}

// sendEmail sends an email using SMTP
func (e *EmailNotifier) sendEmail(subject, body string) error {
	auth := smtp.PlainAuth("", e.config.FromEmail, e.config.FromPassword, e.config.SMTPHost)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", e.config.ToEmail, subject, body))

	addr := e.config.SMTPHost + ":" + e.config.SMTPPort
	err := smtp.SendMail(addr, auth, e.config.FromEmail, []string{e.config.ToEmail}, msg)

	if err != nil {
		log.Printf("❌ Failed to send email: %v", err)
		return err
	}

	log.Printf("📧 Email sent: %s", subject)
	return nil
}

// Helper function
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
