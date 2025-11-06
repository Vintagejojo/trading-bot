package main

import (
	"fmt"
	"log"
	"time"

	"rsi-bot/pkg/models"
	"rsi-bot/pkg/strategy"
)

/*
Multi-Timeframe Strategy Example

This example demonstrates how to:
1. Set up the multi-timeframe strategy
2. Process incoming price data
3. Generate trading signals
4. Calculate position sizing with risk management
5. Monitor positions with trailing stops
6. Check market conditions before trading

Usage:
  go run examples/multitimeframe_example.go
*/

func main() {
	fmt.Println("=== Multi-Timeframe Trading Strategy Example ===\n")

	// ========================================
	// Step 1: Configure the Strategy
	// ========================================
	fmt.Println("Step 1: Configuring multi-timeframe strategy...")

	// Create multi-timeframe strategy with default configuration
	strategyConfig := strategy.DefaultMultiTimeframeStrategyConfig()

	// Customize thresholds if needed
	strategyConfig.RSIOversold = 30.0
	strategyConfig.RSIOverbought = 70.0
	strategyConfig.RequireDailyTrendConfirmation = true
	strategyConfig.RequireHourlySignal = true
	strategyConfig.Require5MinuteEntry = true

	mts, err := strategy.NewMultiTimeframeStrategy(strategyConfig)
	if err != nil {
		log.Fatalf("Failed to create strategy: %v", err)
	}

	fmt.Println("✓ Strategy configured")
	fmt.Printf("  - Timeframes: Daily (1d), Hourly (1h), 5-Minute (5m)\n")
	fmt.Printf("  - Indicators: RSI, MACD, Bollinger Bands\n")
	fmt.Printf("  - RSI thresholds: Oversold=%.0f, Overbought=%.0f\n\n",
		strategyConfig.RSIOversold, strategyConfig.RSIOverbought)

	// ========================================
	// Step 2: Configure Risk Management
	// ========================================
	fmt.Println("Step 2: Configuring risk management...")

	riskConfig := strategy.DefaultRiskConfig()
	riskConfig.MaxPositionSizePercent = 10.0  // Max 10% of portfolio per trade
	riskConfig.RiskPerTradePercent = 2.0      // Risk 2% per trade
	riskConfig.StopLossPercent = 3.0          // 3% stop-loss
	riskConfig.UseRiskRewardRatio = true
	riskConfig.RiskRewardRatio = 2.0          // 2:1 reward/risk ratio
	riskConfig.UseTrailingStop = true
	riskConfig.TrailingStopPercent = 4.0      // Activate trailing at 4% profit
	riskConfig.TrailingStopDistance = 2.0     // Trail 2% below peak

	riskManager := strategy.NewRiskManager(riskConfig)

	fmt.Println("✓ Risk management configured")
	fmt.Printf("  - Max position size: %.0f%% of portfolio\n", riskConfig.MaxPositionSizePercent)
	fmt.Printf("  - Risk per trade: %.0f%%\n", riskConfig.RiskPerTradePercent)
	fmt.Printf("  - Stop-loss: %.0f%%\n", riskConfig.StopLossPercent)
	fmt.Printf("  - Risk/Reward ratio: %.0f:1\n", riskConfig.RiskRewardRatio)
	fmt.Printf("  - Trailing stop: %.0f%% activation, %.0f%% trail\n\n",
		riskConfig.TrailingStopPercent, riskConfig.TrailingStopDistance)

	// ========================================
	// Step 3: Configure Market Condition Analyzer
	// ========================================
	fmt.Println("Step 3: Configuring market condition analyzer...")

	mcConfig := strategy.DefaultMarketConditionConfig()
	mcConfig.MinVolatilityPercent = 1.0
	mcConfig.MaxVolatilityPercent = 10.0
	mcConfig.UseVolumeFilter = true
	mcConfig.MinVolumeMultiplier = 0.5

	mcAnalyzer := strategy.NewMarketConditionAnalyzer(mcConfig)

	fmt.Println("✓ Market condition analyzer configured")
	fmt.Printf("  - Volatility range: %.0f%% - %.0f%%\n",
		mcConfig.MinVolatilityPercent, mcConfig.MaxVolatilityPercent)
	fmt.Printf("  - Min volume: %.0fx average\n\n", mcConfig.MinVolumeMultiplier)

	// ========================================
	// Step 4: Simulate Price Data Stream
	// ========================================
	fmt.Println("Step 4: Simulating price data stream...\n")

	// Portfolio state
	portfolioValue := 10000.0
	position := &models.Position{
		InPosition: false,
		Quantity:   0,
		EntryPrice: 0,
	}

	// Simulate incoming 1-minute kline data
	// In production, this would come from Binance WebSocket
	simulatedPrices := generateSimulatedPrices()

	volumeTracker := strategy.NewVolumeTracker(50)
	var trailingStop *strategy.TrailingStopTracker
	var stopLossPrice, takeProfitPrice float64

	// Process each price update
	for i, priceData := range simulatedPrices {
		timestamp := time.Now().Add(time.Duration(i) * time.Minute)
		price := priceData.Price
		volume := priceData.Volume

		// Update the multi-timeframe manager
		err := mts.Update(price, volume, timestamp)
		if err != nil {
			log.Printf("Error updating strategy: %v", err)
			continue
		}

		// Track volume history
		volumeTracker.Add(volume)

		// Skip if not enough data yet
		if !mts.IsReady() {
			if i%50 == 0 {
				fmt.Printf("[%d] Warming up... Collecting data for all timeframes\n", i)
			}
			continue
		}

		// ========================================
		// Step 5: Check Market Conditions
		// ========================================

		// Get hourly snapshot for volatility
		snapshots := mts.GetMultiTimeframeManager().GetAllSnapshots()
		hourlySnapshot, hasHourly := snapshots[strategy.Timeframe1h]

		if !hasHourly || !hourlySnapshot.BBandsReady {
			continue
		}

		// Check if market is tradeable
		marketCondition := mcAnalyzer.AnalyzeMarketConditions(
			hourlySnapshot.BBandsWidth,
			volume,
			volumeTracker.GetHistory(),
			price*0.9999, // Simulated bid
			price*1.0001, // Simulated ask
		)

		// ========================================
		// Step 6: Generate Trading Signal
		// ========================================

		signalContext := strategy.SignalContext{
			CurrentPrice:  price,
			Position:      position,
			IndicatorData: make(map[string]float64),
		}

		signal := mts.GenerateSignal(signalContext)
		signalReason := mts.GetSignalReason()

		// ========================================
		// Step 7: Execute Trades Based on Signals
		// ========================================

		// BUY SIGNAL
		if signal == strategy.SignalBuy && !position.InPosition && marketCondition.IsTradeableMarket {
			fmt.Printf("\n📊 [%d] BUY SIGNAL DETECTED\n", i)
			fmt.Printf("Time: %s\n", timestamp.Format("15:04:05"))
			fmt.Printf("Price: %.8f\n", price)
			fmt.Printf("Reason: %s\n", signalReason)
			fmt.Printf("Market: %s\n\n", marketCondition.String())

			// Calculate position size
			positionSize, err := riskManager.CalculatePositionSize(
				portfolioValue,
				price,
				0, // Not using ATR-based stop in this example
			)

			if err != nil {
				log.Printf("Error calculating position size: %v", err)
				continue
			}

			// Validate risk
			err = riskManager.ValidatePositionRisk(
				portfolioValue,
				positionSize.RiskAmount,
				0, // No existing positions
				0, // No existing risk
			)

			if err != nil {
				log.Printf("Position rejected: %v", err)
				continue
			}

			// Execute buy order (simulated)
			fmt.Println("💰 EXECUTING BUY ORDER:")
			fmt.Printf("  Quantity: %.0f\n", positionSize.Quantity)
			fmt.Printf("  Entry Price: %.8f\n", positionSize.EntryPrice)
			fmt.Printf("  Stop-Loss: %.8f (%.2f%% risk)\n",
				positionSize.StopLossPrice,
				((positionSize.EntryPrice-positionSize.StopLossPrice)/positionSize.EntryPrice)*100)
			fmt.Printf("  Take-Profit: %.8f (%.2f%% target)\n",
				positionSize.TakeProfitPrice,
				((positionSize.TakeProfitPrice-positionSize.EntryPrice)/positionSize.EntryPrice)*100)
			fmt.Printf("  Position Value: $%.2f\n", positionSize.PositionValue)
			fmt.Printf("  Risk Amount: $%.2f (%.2f%% of portfolio)\n",
				positionSize.RiskAmount, positionSize.MaxLossPercent)
			fmt.Printf("  Potential Profit: $%.2f\n", positionSize.PotentialProfit)
			fmt.Printf("  Risk/Reward Ratio: %.2f:1\n\n", positionSize.RiskRewardRatio)

			// Update position
			position.InPosition = true
			position.Quantity = positionSize.Quantity
			position.EntryPrice = positionSize.EntryPrice
			position.LastUpdate = timestamp

			stopLossPrice = positionSize.StopLossPrice
			takeProfitPrice = positionSize.TakeProfitPrice

			// Initialize trailing stop
			if riskConfig.UseTrailingStop {
				trailingStop = strategy.NewTrailingStopTracker(
					positionSize.EntryPrice,
					positionSize.StopLossPrice,
					riskConfig.TrailingStopPercent,
					riskConfig.TrailingStopDistance,
				)
			}
		}

		// MANAGE OPEN POSITION
		if position.InPosition {
			// Update trailing stop
			if trailingStop != nil {
				stopTriggered := trailingStop.Update(price)
				stopLossPrice = trailingStop.GetStopLossPrice()

				if stopTriggered {
					fmt.Printf("\n🛑 [%d] TRAILING STOP TRIGGERED\n", i)
					fmt.Printf("Exit Price: %.8f\n", price)

					profit := (price - position.EntryPrice) * position.Quantity
					profitPercent := ((price - position.EntryPrice) / position.EntryPrice) * 100

					fmt.Printf("Profit: $%.2f (%.2f%%)\n", profit, profitPercent)
					fmt.Printf("New Portfolio: $%.2f\n\n", portfolioValue+profit)

					// Close position
					portfolioValue += profit
					position.InPosition = false
					position.Quantity = 0
					position.EntryPrice = 0
					trailingStop = nil
					continue
				}
			}

			// Check regular stop-loss and take-profit
			shouldExit, exitReason := riskManager.ShouldExit(
				position.EntryPrice,
				price,
				stopLossPrice,
				takeProfitPrice,
			)

			if shouldExit {
				fmt.Printf("\n🎯 [%d] EXIT TRIGGERED: %s\n", i, exitReason)

				profit := (price - position.EntryPrice) * position.Quantity
				profitPercent := ((price - position.EntryPrice) / position.EntryPrice) * 100

				fmt.Printf("Profit/Loss: $%.2f (%.2f%%)\n", profit, profitPercent)
				fmt.Printf("New Portfolio: $%.2f\n\n", portfolioValue+profit)

				// Close position
				portfolioValue += profit
				position.InPosition = false
				position.Quantity = 0
				position.EntryPrice = 0
				trailingStop = nil
				continue
			}

			// Log position status every 10 updates
			if i%10 == 0 {
				summary := riskManager.GetPositionSummary(
					position.EntryPrice,
					price,
					position.Quantity,
					stopLossPrice,
					takeProfitPrice,
				)

				fmt.Printf("[%d] Position: %.2f%% P/L | Stop: %.8f | Target: %.8f",
					i, summary.UnrealizedPLPercent, stopLossPrice, takeProfitPrice)

				if trailingStop != nil && trailingStop.TrailingActive {
					fmt.Printf(" | TRAILING ACTIVE")
				}
				fmt.Println()
			}
		}

		// SELL SIGNAL (only if in position)
		if signal == strategy.SignalSell && position.InPosition {
			fmt.Printf("\n📉 [%d] SELL SIGNAL DETECTED\n", i)
			fmt.Printf("Reason: %s\n", signalReason)

			profit := (price - position.EntryPrice) * position.Quantity
			profitPercent := ((price - position.EntryPrice) / position.EntryPrice) * 100

			fmt.Printf("Exit Price: %.8f\n", price)
			fmt.Printf("Profit/Loss: $%.2f (%.2f%%)\n", profit, profitPercent)
			fmt.Printf("New Portfolio: $%.2f\n\n", portfolioValue+profit)

			// Close position
			portfolioValue += profit
			position.InPosition = false
			position.Quantity = 0
			position.EntryPrice = 0
			trailingStop = nil
		}
	}

	// ========================================
	// Step 8: Final Results
	// ========================================
	fmt.Println("\n" + repeat("=", 60))
	fmt.Println("SIMULATION COMPLETE")
	fmt.Println(repeat("=", 60))
	fmt.Printf("Starting Portfolio: $10,000.00\n")
	fmt.Printf("Ending Portfolio:   $%.2f\n", portfolioValue)
	fmt.Printf("Total P/L:          $%.2f (%.2f%%)\n",
		portfolioValue-10000, ((portfolioValue-10000)/10000)*100)
}

// PriceData represents a single price update with volume
type PriceData struct {
	Price  float64
	Volume float64
}

// generateSimulatedPrices creates test data for the example
// In production, this data comes from Binance WebSocket
func generateSimulatedPrices() []PriceData {
	prices := make([]PriceData, 0, 500)

	// Start price
	basePrice := 0.00001000

	// Simulate a trending market with some volatility
	for i := 0; i < 500; i++ {
		// Trend component (upward trend)
		trend := float64(i) * 0.00000001

		// Volatility component (random-ish)
		volatility := (float64((i*7)%20) - 10) * 0.00000005

		price := basePrice + trend + volatility
		volume := 50000000.0 + float64((i*13)%30)*1000000

		prices = append(prices, PriceData{
			Price:  price,
			Volume: volume,
		})
	}

	return prices
}

// Helper function for string repetition
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
