package indicators

import (
	"math/rand"
	"testing"
	"time"
)

// BenchmarkRSI_Update benchmarks the Update method with different scenarios
func BenchmarkRSI_Update(b *testing.B) {
	rsi, _ := NewRSI(14)
	price := 100.0
	timestamp := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsi.Update(price+float64(i%10), timestamp.Add(time.Duration(i)*time.Minute))
	}
}

// BenchmarkRSI_UpdateWithCalculation benchmarks Update when RSI calculation occurs
func BenchmarkRSI_UpdateWithCalculation(b *testing.B) {
	rsi, _ := NewRSI(14)
	timestamp := time.Now()

	// Warm up with 15 data points so calculation happens
	for i := 0; i < 15; i++ {
		rsi.Update(100.0+float64(i), timestamp.Add(time.Duration(i)*time.Minute))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsi.Update(100.0+float64(i%20), timestamp.Add(time.Duration(i+15)*time.Minute))
	}
}

// BenchmarkRSI_Calculate benchmarks the internal calculate method
func BenchmarkRSI_Calculate(b *testing.B) {
	rsi, _ := NewRSI(14)
	timestamp := time.Now()

	// Populate with realistic price data
	for i := 0; i < 30; i++ {
		rsi.Update(100.0+float64(i%20), timestamp.Add(time.Duration(i)*time.Minute))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsi.calculate()
	}
}

// BenchmarkRSI_GetValue benchmarks retrieving the RSI value
func BenchmarkRSI_GetValue(b *testing.B) {
	rsi, _ := NewRSI(14)
	timestamp := time.Now()

	// Populate with data
	for i := 0; i < 30; i++ {
		rsi.Update(100.0+float64(i), timestamp.Add(time.Duration(i)*time.Minute))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsi.GetValue()
	}
}

// BenchmarkRSI_NewRSI benchmarks RSI creation
func BenchmarkRSI_NewRSI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRSI(14)
	}
}

// BenchmarkRSI_DifferentPeriods benchmarks RSI with different periods
func BenchmarkRSI_DifferentPeriods(b *testing.B) {
	periods := []int{7, 14, 21, 50}

	for _, period := range periods {
		b.Run(string(rune(period)), func(b *testing.B) {
			rsi, _ := NewRSI(period)
			timestamp := time.Now()

			// Warm up
			for i := 0; i < period+20; i++ {
				rsi.Update(100.0+float64(i), timestamp.Add(time.Duration(i)*time.Minute))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rsi.Update(100.0+float64(i%50), timestamp.Add(time.Duration(i)*time.Minute))
			}
		})
	}
}

// BenchmarkRSI_MemoryAllocation benchmarks memory allocation patterns
func BenchmarkRSI_MemoryAllocation(b *testing.B) {
	timestamp := time.Now()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rsi, _ := NewRSI(14)
		for j := 0; j < 30; j++ {
			rsi.Update(100.0+float64(j), timestamp.Add(time.Duration(j)*time.Minute))
		}
		rsi.GetValue()
	}
}

// BenchmarkRSI_HighFrequencyUpdates simulates rapid market updates
func BenchmarkRSI_HighFrequencyUpdates(b *testing.B) {
	rsi, _ := NewRSI(14)
	timestamp := time.Now()
	rand.Seed(time.Now().UnixNano())

	// Warm up
	for i := 0; i < 30; i++ {
		rsi.Update(100.0+rand.Float64()*10, timestamp.Add(time.Duration(i)*time.Minute))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		price := 100.0 + rand.Float64()*10 // Random price fluctuation
		rsi.Update(price, timestamp.Add(time.Duration(i+30)*time.Minute))
	}
}

// BenchmarkRSI_SliceManagement benchmarks the slice trimming logic
func BenchmarkRSI_SliceManagement(b *testing.B) {
	rsi, _ := NewRSI(14)
	timestamp := time.Now()

	// Fill to capacity (period + 20)
	for i := 0; i < 34; i++ {
		rsi.Update(100.0+float64(i), timestamp.Add(time.Duration(i)*time.Minute))
	}

	b.ResetTimer()
	// This should trigger slice trimming on each update
	for i := 0; i < b.N; i++ {
		rsi.Update(100.0+float64(i%10), timestamp.Add(time.Duration(i+34)*time.Minute))
	}
}

// BenchmarkRSI_ConcurrentReads benchmarks concurrent read access
func BenchmarkRSI_ConcurrentReads(b *testing.B) {
	rsi, _ := NewRSI(14)
	timestamp := time.Now()

	// Populate with data
	for i := 0; i < 30; i++ {
		rsi.Update(100.0+float64(i), timestamp.Add(time.Duration(i)*time.Minute))
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rsi.GetValue()
			rsi.IsReady()
			rsi.GetDataCount()
		}
	})
}

// BenchmarkRSI_Reset benchmarks the Reset operation
func BenchmarkRSI_Reset(b *testing.B) {
	timestamp := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsi, _ := NewRSI(14)
		for j := 0; j < 30; j++ {
			rsi.Update(100.0+float64(j), timestamp.Add(time.Duration(j)*time.Minute))
		}
		rsi.Reset()
	}
}

// BenchmarkRSI_RealisticTrading simulates realistic trading scenario
func BenchmarkRSI_RealisticTrading(b *testing.B) {
	rsi, _ := NewRSI(14)
	timestamp := time.Now()
	basePrice := 100.0
	rand.Seed(time.Now().UnixNano())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate price movement with trend and noise
		change := (rand.Float64() - 0.5) * 2 // Random change between -1 and 1
		basePrice += change

		rsi.Update(basePrice, timestamp.Add(time.Duration(i)*time.Minute))

		// Simulate checking conditions every update
		if rsi.IsReady() {
			values, _ := rsi.GetValue()
			_ = values[ValueKeyRSI] // Simulate reading the value
		}
	}
}
