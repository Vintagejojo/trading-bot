<template>
  <div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
    <h2 class="text-xl font-semibold mb-4 text-blue-400">Bot Controls</h2>

    <div v-if="!botStatus.running" class="space-y-4">
      <!-- Strategy Selection -->
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-2">Strategy</label>
        <select v-model="config.strategy" @change="loadStrategyParams"
                class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
          <option value="rsi">RSI - Mean Reversion</option>
          <option value="macd">MACD - Trend Following</option>
          <option value="bbands">Bollinger Bands - Volatility</option>
        </select>
      </div>

      <!-- Trading Pair Selection -->
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-2">Trading Pair</label>
        <select v-model="config.symbol"
                class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
          <option v-for="pair in tradingPairs" :key="pair.symbol" :value="pair.symbol">
            {{ pair.symbol }} - {{ pair.description }}
          </option>
        </select>
        <p class="mt-1 text-xs text-gray-500">{{ getSelectedPairInfo() }}</p>
      </div>

      <!-- Quantity Input -->
      <div>
        <label class="block text-sm font-medium text-gray-300 mb-2">Quantity</label>
        <input v-model.number="config.quantity" type="number" step="0.00000001"
               class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
      </div>

      <!-- Strategy Parameters -->
      <div v-if="config.strategy === 'rsi'" class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">RSI Period</label>
          <input v-model.number="config.params.period" type="number" min="2" max="100"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Overbought Level</label>
          <input v-model.number="config.params.overbought_level" type="number" min="50" max="100"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Oversold Level</label>
          <input v-model.number="config.params.oversold_level" type="number" min="0" max="50"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
      </div>

      <div v-if="config.strategy === 'macd'" class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Fast Period</label>
          <input v-model.number="config.params.fast_period" type="number" min="2"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Slow Period</label>
          <input v-model.number="config.params.slow_period" type="number" min="2"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Signal Period</label>
          <input v-model.number="config.params.signal_period" type="number" min="2"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
      </div>

      <div v-if="config.strategy === 'bbands'" class="space-y-3">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Period</label>
          <input v-model.number="config.params.period" type="number" min="2"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">Std Dev Multiplier</label>
          <input v-model.number="config.params.std_dev" type="number" step="0.1" min="0.5"
                 class="w-full bg-gray-700 border border-gray-600 rounded px-3 py-2 text-gray-100 focus:outline-none focus:border-blue-500">
        </div>
      </div>

      <!-- Paper Trading Toggle -->
      <div class="flex items-center space-x-3 py-2">
        <input v-model="config.paperTrading" type="checkbox" id="paperTrading"
               class="w-4 h-4 text-blue-600 bg-gray-700 border-gray-600 rounded focus:ring-blue-500">
        <label for="paperTrading" class="text-sm font-medium text-gray-300">
          Paper Trading (Recommended)
        </label>
      </div>

      <!-- Start Button -->
      <button @click="handleStart"
              class="w-full bg-green-600 hover:bg-green-700 text-white font-semibold py-3 px-4 rounded transition-colors">
        Start Bot
      </button>
    </div>

    <!-- Running State -->
    <div v-else class="space-y-4">
      <div class="bg-gray-700 rounded p-4 space-y-2">
        <div class="flex justify-between text-sm">
          <span class="text-gray-400">Strategy:</span>
          <span class="text-gray-100 font-medium uppercase">{{ botStatus.strategy }}</span>
        </div>
        <div class="flex justify-between text-sm">
          <span class="text-gray-400">Symbol:</span>
          <span class="text-gray-100 font-medium">{{ botStatus.symbol }}</span>
        </div>
        <div class="flex justify-between text-sm">
          <span class="text-gray-400">Mode:</span>
          <span :class="botStatus.trading_mode === 'live' ? 'text-red-400' : 'text-blue-400'"
                class="font-medium uppercase">
            {{ botStatus.trading_mode }}
          </span>
        </div>
      </div>

      <!-- Stop Button -->
      <button @click="handleStop"
              class="w-full bg-red-600 hover:bg-red-700 text-white font-semibold py-3 px-4 rounded transition-colors">
        Stop Bot
      </button>
    </div>
  </div>
</template>

<script>
import { ref, watch } from 'vue'
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
    const tradingPairs = [
      {
        symbol: 'BTCUSDT',
        description: 'Bitcoin/USDT',
        strategy: 'Momentum & RSI',
        info: 'Unmatched liquidity, ideal for momentum and RSI-based strategies'
      },
      {
        symbol: 'BTCETH',
        description: 'Bitcoin/Ethereum',
        strategy: 'Trend Following & Divergence',
        info: 'Track relative strength between top two cryptocurrencies'
      },
      {
        symbol: 'BTCBNB',
        description: 'Bitcoin/BNB',
        strategy: 'High Volatility',
        info: 'Binance ecosystem influence with high volatility opportunities'
      },
      {
        symbol: 'BTCSOL',
        description: 'Bitcoin/Solana',
        strategy: 'Swing Trading',
        info: 'Capitalize on Solana growth with swing trading opportunities'
      },
      {
        symbol: 'BTCXRP',
        description: 'Bitcoin/Ripple',
        strategy: 'Breakout & Reversal',
        info: 'High volume breakout and reversal strategies'
      },
      {
        symbol: 'BTCADA',
        description: 'Bitcoin/Cardano',
        strategy: 'Bollinger Bands & Stochastic',
        info: 'Smoother price action, compatible with BB and Stochastic setups'
      },
      {
        symbol: 'BTCDOGE',
        description: 'Bitcoin/Dogecoin',
        strategy: 'Scalping',
        info: 'Meme coin volatility perfect for short-term scalping bots'
      },
      {
        symbol: 'BTCLINK',
        description: 'Bitcoin/Chainlink',
        strategy: 'DeFi Mid-Cap',
        info: 'DeFi relevance with solid mid-cap behavior'
      },
      {
        symbol: 'BTCMATIC',
        description: 'Bitcoin/Polygon',
        strategy: 'Ecosystem Trends',
        info: 'Follow Polygon scaling narrative and ecosystem-driven trends'
      },
      {
        symbol: 'BTCAVAX',
        description: 'Bitcoin/Avalanche',
        strategy: 'Breakout & Momentum',
        info: 'Aligned with Avalanche expanding ecosystem'
      },
      {
        symbol: 'SHIBUSDT',
        description: 'Shiba Inu/USDT',
        strategy: 'High Volume Meme Coin',
        info: 'Extreme volatility with massive trading volume, ideal for aggressive scalping'
      }
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

    const getSelectedPairInfo = () => {
      const pair = tradingPairs.find(p => p.symbol === config.value.symbol)
      return pair ? `${pair.strategy} - ${pair.info}` : ''
    }

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

    // Load default params on mount
    loadStrategyParams()

    return {
      tradingPairs,
      config,
      getSelectedPairInfo,
      loadStrategyParams,
      handleStart,
      handleStop
    }
  }
}
</script>
