<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <v-icon icon="mdi-cog" class="mr-2" color="primary"></v-icon>
      Bot Controls
    </v-card-title>

    <v-card-text>
      <!-- Bot Not Running - Show Config -->
      <div v-if="!botStatus.running">
        <v-select
          v-model="config.strategy"
          :items="strategies"
          label="Strategy"
          variant="outlined"
          density="comfortable"
          @update:model-value="loadStrategyParams"
        ></v-select>

        <v-select
          v-model="config.symbol"
          :items="tradingPairs"
          item-title="label"
          item-value="symbol"
          label="Trading Pair"
          variant="outlined"
          density="comfortable"
          class="mt-3"
        >
          <template v-slot:item="{ props, item }">
            <v-list-item v-bind="props">
              <v-list-item-subtitle>{{ item.raw.info }}</v-list-item-subtitle>
            </v-list-item>
          </template>
        </v-select>

        <v-text-field
          v-model.number="config.quantity"
          :label="config.strategy === 'dca' ? 'Amount (USD)' : 'Quantity'"
          type="number"
          :step="config.strategy === 'dca' ? '1' : '0.00000001'"
          variant="outlined"
          density="comfortable"
          class="mt-3"
          :prefix="config.strategy === 'dca' ? '$' : ''"
          :hint="quantityHint"
          persistent-hint
        >
          <template v-slot:append-inner v-if="config.strategy === 'dca' && currentPrice > 0">
            <v-chip size="x-small" color="info" variant="flat">
              ≈ {{ (config.quantity / currentPrice).toFixed(8) }} BTC
            </v-chip>
          </template>
        </v-text-field>

        <!-- RSI Parameters -->
        <v-expand-transition>
          <v-card v-if="config.strategy === 'rsi'" variant="outlined" class="mt-3">
            <v-card-subtitle>RSI Parameters</v-card-subtitle>
            <v-card-text>
              <v-text-field
                v-model.number="config.params.period"
                label="RSI Period"
                type="number"
                min="2"
                max="100"
                variant="outlined"
                density="compact"
              ></v-text-field>

              <v-text-field
                v-model.number="config.params.overbought_level"
                label="Overbought Level"
                type="number"
                min="50"
                max="100"
                variant="outlined"
                density="compact"
                class="mt-2"
              ></v-text-field>

              <v-text-field
                v-model.number="config.params.oversold_level"
                label="Oversold Level"
                type="number"
                min="0"
                max="50"
                variant="outlined"
                density="compact"
                class="mt-2"
              ></v-text-field>
            </v-card-text>
          </v-card>
        </v-expand-transition>

        <!-- MACD Parameters -->
        <v-expand-transition>
          <v-card v-if="config.strategy === 'macd'" variant="outlined" class="mt-3">
            <v-card-subtitle>MACD Parameters</v-card-subtitle>
            <v-card-text>
              <v-text-field
                v-model.number="config.params.fast_period"
                label="Fast Period"
                type="number"
                min="2"
                variant="outlined"
                density="compact"
              ></v-text-field>

              <v-text-field
                v-model.number="config.params.slow_period"
                label="Slow Period"
                type="number"
                min="2"
                variant="outlined"
                density="compact"
                class="mt-2"
              ></v-text-field>

              <v-text-field
                v-model.number="config.params.signal_period"
                label="Signal Period"
                type="number"
                min="2"
                variant="outlined"
                density="compact"
                class="mt-2"
              ></v-text-field>
            </v-card-text>
          </v-card>
        </v-expand-transition>

        <!-- Bollinger Bands Parameters -->
        <v-expand-transition>
          <v-card v-if="config.strategy === 'bbands'" variant="outlined" class="mt-3">
            <v-card-subtitle>Bollinger Bands Parameters</v-card-subtitle>
            <v-card-text>
              <v-text-field
                v-model.number="config.params.period"
                label="Period"
                type="number"
                min="2"
                variant="outlined"
                density="compact"
              ></v-text-field>

              <v-text-field
                v-model.number="config.params.std_dev"
                label="Std Dev Multiplier"
                type="number"
                step="0.1"
                min="0.5"
                variant="outlined"
                density="compact"
                class="mt-2"
              ></v-text-field>
            </v-card-text>
          </v-card>
        </v-expand-transition>

        <!-- DCA Parameters -->
        <v-expand-transition>
          <v-card v-if="config.strategy === 'dca'" variant="outlined" class="mt-3">
            <v-card-subtitle>DCA Schedule</v-card-subtitle>
            <v-card-text>
              <v-select
                v-model.number="config.params.day_of_week"
                :items="daysOfWeek"
                label="Day of Week"
                variant="outlined"
                density="compact"
              ></v-select>

              <v-text-field
                v-model.number="config.params.hour_of_day"
                label="Hour (UTC 0-23)"
                type="number"
                min="0"
                max="23"
                variant="outlined"
                density="compact"
                class="mt-2"
                hint="9 = 9am UTC, 14 = 2pm UTC"
              ></v-text-field>

              <v-divider class="my-3"></v-divider>

              <v-checkbox
                v-model="config.params.buy_the_dip"
                label="Enable Buy-the-Dip"
                color="primary"
                density="compact"
                hint="Buy extra when price drops significantly"
              ></v-checkbox>

              <v-expand-transition>
                <div v-if="config.params.buy_the_dip">
                  <v-text-field
                    v-model.number="config.params.dip_threshold"
                    label="Dip Threshold (%)"
                    type="number"
                    step="0.1"
                    min="1"
                    max="20"
                    variant="outlined"
                    density="compact"
                    class="mt-2"
                    hint="Buy extra when price drops this % from 24h high"
                  ></v-text-field>

                  <v-text-field
                    v-model.number="config.params.dip_multiplier"
                    label="Dip Buy Multiplier"
                    type="number"
                    step="0.1"
                    min="1"
                    max="5"
                    variant="outlined"
                    density="compact"
                    class="mt-2"
                    hint="1.5 = buy 1.5x normal amount on dips"
                  ></v-text-field>
                </div>
              </v-expand-transition>
            </v-card-text>
          </v-card>
        </v-expand-transition>

        <v-checkbox
          v-model="config.paperTrading"
          label="Paper Trading (Recommended)"
          color="info"
          class="mt-3"
        ></v-checkbox>

        <v-btn
          block
          size="large"
          color="success"
          prepend-icon="mdi-play"
          @click="handleStart"
          class="mt-4"
        >
          Start Bot
        </v-btn>
      </div>

      <!-- Bot Running - Show Status -->
      <div v-else>
        <v-list>
          <v-list-item>
            <v-list-item-title>Strategy</v-list-item-title>
            <template v-slot:append>
              <v-chip color="primary" size="small">{{ botStatus.strategy.toUpperCase() }}</v-chip>
            </template>
          </v-list-item>

          <v-list-item>
            <v-list-item-title>Symbol</v-list-item-title>
            <template v-slot:append>
              <strong>{{ botStatus.symbol }}</strong>
            </template>
          </v-list-item>

          <v-list-item>
            <v-list-item-title>Mode</v-list-item-title>
            <template v-slot:append>
              <v-chip
                :color="botStatus.trading_mode === 'live' ? 'error' : 'info'"
                size="small"
              >
                {{ botStatus.trading_mode.toUpperCase() }}
              </v-chip>
            </template>
          </v-list-item>
        </v-list>

        <v-btn
          block
          size="large"
          color="error"
          prepend-icon="mdi-stop"
          @click="handleStop"
          class="mt-4"
        >
          Stop Bot
        </v-btn>

        <!-- Demo Data Controls (for screenshots/testing) -->
        <v-divider class="my-4"></v-divider>

        <v-card variant="tonal" color="info" class="mt-4">
          <v-card-subtitle>Demo/Testing</v-card-subtitle>
          <v-card-text>
            <v-btn
              block
              size="small"
              color="info"
              prepend-icon="mdi-database-plus"
              @click="generateDemo"
              class="mb-2"
            >
              Generate Demo Trades
            </v-btn>
            <v-btn
              block
              size="small"
              variant="outlined"
              prepend-icon="mdi-delete"
              @click="clearDemo"
            >
              Clear Demo Data
            </v-btn>
            <div class="text-caption text-grey mt-2">
              For screenshots and UI testing
            </div>
          </v-card-text>
        </v-card>
      </div>
    </v-card-text>
  </v-card>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { GetDefaultStrategyParams, GenerateDemoTrades, ClearDemoTrades } from '../../wailsjs/go/main/App'

export default {
  name: 'BotControls',
  props: {
    botStatus: {
      type: Object,
      required: true
    }
  },
  emits: ['start-bot', 'stop-bot', 'refresh-data'],
  setup(props, { emit }) {
    const strategies = [
      { title: 'DCA - Dollar Cost Averaging (Recommended)', value: 'dca' },
      { title: 'RSI - Mean Reversion', value: 'rsi' },
      { title: 'MACD - Trend Following', value: 'macd' },
      { title: 'Bollinger Bands - Volatility', value: 'bbands' },
      { title: 'Multi-Timeframe - Advanced (Daily/1h/5m)', value: 'multitimeframe' }
    ]

    const tradingPairs = [
      { symbol: 'BTCUSDT', label: 'BTCUSDT - Bitcoin/USDT', info: 'Unmatched liquidity, ideal for momentum and RSI-based strategies' },
      { symbol: 'BTCETH', label: 'BTCETH - Bitcoin/Ethereum', info: 'Track relative strength between top two cryptocurrencies' },
      { symbol: 'BTCBNB', label: 'BTCBNB - Bitcoin/BNB', info: 'Binance ecosystem influence with high volatility opportunities' },
      { symbol: 'BTCSOL', label: 'BTCSOL - Bitcoin/Solana', info: 'Capitalize on Solana growth with swing trading opportunities' },
      { symbol: 'BTCXRP', label: 'BTCXRP - Bitcoin/Ripple', info: 'High volume breakout and reversal strategies' },
      { symbol: 'BTCADA', label: 'BTCADA - Bitcoin/Cardano', info: 'Smoother price action, compatible with BB and Stochastic setups' },
      { symbol: 'BTCDOGE', label: 'BTCDOGE - Bitcoin/Dogecoin', info: 'Meme coin volatility perfect for short-term scalping bots' },
      { symbol: 'BTCLINK', label: 'BTCLINK - Bitcoin/Chainlink', info: 'DeFi relevance with solid mid-cap behavior' },
      { symbol: 'BTCMATIC', label: 'BTCMATIC - Bitcoin/Polygon', info: 'Follow Polygon scaling narrative and ecosystem-driven trends' },
      { symbol: 'BTCAVAX', label: 'BTCAVAX - Bitcoin/Avalanche', info: 'Aligned with Avalanche expanding ecosystem' },
      { symbol: 'SHIBUSDT', label: 'SHIBUSDT - Shiba Inu/USDT', info: 'Extreme volatility with massive trading volume, ideal for aggressive scalping' }
    ]

    const daysOfWeek = [
      { title: 'Sunday', value: 0 },
      { title: 'Monday', value: 1 },
      { title: 'Tuesday', value: 2 },
      { title: 'Wednesday', value: 3 },
      { title: 'Thursday', value: 4 },
      { title: 'Friday', value: 5 },
      { title: 'Saturday', value: 6 }
    ]

    const currentPrice = ref(0)

    const config = ref({
      strategy: 'dca',
      symbol: 'BTCUSDT',
      quantity: 100,
      paperTrading: true,
      params: {
        day_of_week: 1,
        hour_of_day: 9,
        buy_the_dip: false,
        dip_threshold: 5.0,
        dip_multiplier: 1.5
      }
    })

    const quantityHint = computed(() => {
      if (config.value.strategy === 'dca') {
        return 'Dollar amount to invest per purchase'
      }
      return 'Amount of crypto to buy/sell per trade'
    })

    const fetchCurrentPrice = async () => {
      try {
        // Fetch BTC price from Binance public API
        const response = await fetch('https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT')
        const data = await response.json()
        currentPrice.value = parseFloat(data.price)
      } catch (error) {
        console.error('Failed to fetch BTC price:', error)
        currentPrice.value = 0
      }
    }

    const loadStrategyParams = async () => {
      try {
        const params = await GetDefaultStrategyParams(config.value.strategy)
        config.value.params = params

        // Adjust default quantity for DCA
        if (config.value.strategy === 'dca') {
          config.value.quantity = 100 // $100 default for DCA
        } else if (config.value.quantity === 100) {
          config.value.quantity = 0.001 // Reset to BTC amount if switching from DCA
        }
      } catch (error) {
        console.error('Failed to load strategy params:', error)
      }
    }

    const handleStart = () => {
      if (!config.value.symbol) {
        alert('Please enter a symbol')
        return
      }
      if (config.value.quantity <= 0) {
        alert('Please enter a valid quantity')
        return
      }
      emit('start-bot', config.value)
    }

    const handleStop = () => {
      if (confirm('Are you sure you want to stop the bot?')) {
        emit('stop-bot')
      }
    }

    const generateDemo = async () => {
      try {
        await GenerateDemoTrades()
        // Trigger immediate refresh of all data
        emit('refresh-data')
        // Small delay to ensure DB commits, then show success
        setTimeout(() => {
          alert('✅ Demo trades generated! Check Performance and Current Position sections.')
        }, 100)
      } catch (error) {
        console.error('Failed to generate demo trades:', error)
        alert('Failed to generate demo trades: ' + error)
      }
    }

    const clearDemo = async () => {
      if (!confirm('Clear all demo/paper trades?')) {
        return
      }
      try {
        await ClearDemoTrades()
        // Trigger immediate refresh
        emit('refresh-data')
        setTimeout(() => {
          alert('✅ Demo trades cleared!')
        }, 100)
      } catch (error) {
        console.error('Failed to clear demo trades:', error)
        alert('Failed to clear demo trades: ' + error)
      }
    }

    // Load initial params and fetch price on mount
    onMounted(() => {
      loadStrategyParams()
      fetchCurrentPrice()
      // Refresh price every 30 seconds
      setInterval(fetchCurrentPrice, 30000)
    })

    return {
      strategies,
      tradingPairs,
      daysOfWeek,
      config,
      currentPrice,
      quantityHint,
      loadStrategyParams,
      handleStart,
      handleStop,
      generateDemo,
      clearDemo
    }
  }
}
</script>
