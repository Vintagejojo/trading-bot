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
          label="Quantity"
          type="number"
          step="0.00000001"
          variant="outlined"
          density="comfortable"
          class="mt-3"
        ></v-text-field>

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
      </div>
    </v-card-text>
  </v-card>
</template>

<script>
import { ref } from 'vue'
import { GetDefaultStrategyParams } from '../../wailsjs/go/main/App'

export default {
  name: 'BotControls',
  props: {
    botStatus: {
      type: Object,
      required: true
    }
  },
  emits: ['start-bot', 'stop-bot'],
  setup(props, { emit }) {
    const strategies = [
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

    const config = ref({
      strategy: 'rsi',
      symbol: 'BTCUSDT',
      quantity: 0.001,
      paperTrading: true,
      params: {
        period: 14,
        overbought_level: 70,
        oversold_level: 30
      }
    })

    const loadStrategyParams = async () => {
      try {
        const params = await GetDefaultStrategyParams(config.value.strategy)
        config.value.params = params
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

    loadStrategyParams()

    return {
      strategies,
      tradingPairs,
      config,
      loadStrategyParams,
      handleStart,
      handleStop
    }
  }
}
</script>
