<template>
  <v-card>
    <v-card-title class="d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <v-icon icon="mdi-chart-line" class="mr-2" color="success"></v-icon>
        Performance
      </div>
      <!-- Debug indicator to show which mode is active -->
      <v-chip v-if="strategy" size="x-small" variant="outlined" color="info">
        {{ strategy === 'dca' ? 'DCA Mode' : 'Trading Mode' }}
      </v-chip>
    </v-card-title>

    <v-card-text>
      <!-- DCA Strategy View -->
      <div v-if="strategy === 'dca'">
        <div v-if="!portfolioStats || !portfolioStats.total_holdings || portfolioStats.total_holdings === 0" class="text-center py-8 text-grey">
          <v-icon icon="mdi-chart-box-outline" size="64" class="mb-2"></v-icon>
          <p>No DCA purchases yet</p>
          <p class="text-caption">Start the bot to begin accumulating</p>
        </div>

        <div v-else>
          <!-- Unrealized Gain/Loss Card with Neon Glow -->
          <v-card
            :color="portfolioStats.unrealized_gain >= 0 ? 'success' : 'error'"
            variant="flat"
            :class="portfolioStats.unrealized_gain >= 0 ? 'neon-card-success' : 'neon-card-error'"
            class="mb-4"
          >
            <v-card-text>
              <div class="text-caption">Unrealized Gain/Loss</div>
              <div
                class="text-h4 font-weight-bold crypto-mono"
                :class="portfolioStats.unrealized_gain >= 0 ? 'neon-glow-success' : 'neon-glow-error'"
              >
                {{ portfolioStats.unrealized_gain >= 0 ? '+' : '' }}${{ portfolioStats.unrealized_gain?.toFixed(2) || '0.00' }}
              </div>
              <div class="text-caption mt-1">
                {{ portfolioStats.unrealized_roi >= 0 ? '+' : '' }}{{ portfolioStats.unrealized_roi?.toFixed(2) || '0.00' }}% ROI
              </div>
            </v-card-text>
          </v-card>

          <!-- Current Price Row (only show when price is available) -->
          <v-card v-if="currentPrice > 0" variant="outlined" class="mb-4 neon-card-info">
            <v-card-text>
              <div class="d-flex justify-space-between align-center">
                <div>
                  <div class="text-caption text-grey">Current Price</div>
                  <div
                    class="text-h5 font-weight-bold crypto-mono neon-glow-info"
                    :class="isPricePulsing ? 'neon-pulse' : ''"
                  >
                    ${{ currentPrice.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
                  </div>
                </div>
                <v-chip
                  v-if="priceChangePercent24h !== 0"
                  :color="priceChangePercent24h >= 0 ? 'success' : 'error'"
                  variant="flat"
                  class="neon-card-info"
                >
                  <v-icon
                    :icon="priceChangePercent24h >= 0 ? 'mdi-arrow-up' : 'mdi-arrow-down'"
                    size="small"
                    class="mr-1"
                  ></v-icon>
                  <span class="crypto-mono">
                    {{ priceChangePercent24h >= 0 ? '+' : '' }}{{ priceChangePercent24h.toFixed(2) }}% (24h)
                  </span>
                </v-chip>
              </div>
            </v-card-text>
          </v-card>

          <v-row>
            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Total Holdings</div>
                  <div class="text-h6">{{ portfolioStats.total_holdings?.toFixed(6) || '0.000000' }} {{ portfolioStats.symbol?.replace('USDT', '') || 'BTC' }}</div>
                </v-card-text>
              </v-card>
            </v-col>

            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Current Value</div>
                  <div class="text-h6">${{ portfolioStats.current_value?.toFixed(2) || '0.00' }}</div>
                </v-card-text>
              </v-card>
            </v-col>

            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Total Invested</div>
                  <div class="text-h6">${{ portfolioStats.total_cost?.toFixed(2) || '0.00' }}</div>
                </v-card-text>
              </v-card>
            </v-col>

            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Average Cost</div>
                  <div class="text-h6">${{ portfolioStats.average_cost?.toFixed(2) || '0.00' }}</div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>

          <v-list density="compact" class="mt-4">
            <v-list-item>
              <v-list-item-title>Total Purchases</v-list-item-title>
              <template v-slot:append>
                <span class="font-weight-bold">{{ portfolioStats.total_buys || 0 }}</span>
              </template>
            </v-list-item>
            <v-list-item>
              <v-list-item-title>Current Price</v-list-item-title>
              <template v-slot:append>
                <span class="font-weight-bold">${{ portfolioStats.current_price?.toFixed(2) || '0.00' }}</span>
              </template>
            </v-list-item>
          </v-list>
        </div>
      </div>

      <!-- Trading Strategy View (RSI, MACD, etc.) -->
      <div v-else>
        <div v-if="stats.total_trades === 0" class="text-center py-8 text-grey">
          <v-icon icon="mdi-chart-box-outline" size="64" class="mb-2"></v-icon>
          <p>No trading data yet</p>
          <p class="text-caption">Start the bot to see performance metrics</p>
        </div>

        <div v-else>
          <v-card
            :color="stats.total_profit_loss >= 0 ? 'success' : 'error'"
            variant="flat"
            :class="stats.total_profit_loss >= 0 ? 'neon-card-success' : 'neon-card-error'"
            class="mb-4"
          >
            <v-card-text>
              <div class="text-caption">Total Profit/Loss</div>
              <div
                class="text-h4 font-weight-bold crypto-mono"
                :class="stats.total_profit_loss >= 0 ? 'neon-glow-success' : 'neon-glow-error'"
              >
                {{ stats.total_profit_loss >= 0 ? '+' : '' }}${{ stats.total_profit_loss?.toFixed(2) || '0.00' }}
              </div>
            </v-card-text>
          </v-card>

          <v-row>
            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Total Trades</div>
                  <div class="text-h6">{{ stats.total_trades || 0 }}</div>
                </v-card-text>
              </v-card>
            </v-col>

            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Win Rate</div>
                  <div class="text-h6" :class="(stats.win_rate || 0) >= 50 ? 'text-success' : 'text-error'">
                    {{ stats.win_rate?.toFixed(1) || '0.0' }}%
                  </div>
                </v-card-text>
              </v-card>
            </v-col>

            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Avg P/L</div>
                  <div class="text-h6" :class="(stats.average_profit_loss || 0) >= 0 ? 'text-success' : 'text-error'">
                    {{ (stats.average_profit_loss || 0) >= 0 ? '+' : '' }}${{ stats.average_profit_loss?.toFixed(2) || '0.00' }}
                  </div>
                </v-card-text>
              </v-card>
            </v-col>

            <v-col cols="6">
              <v-card variant="outlined">
                <v-card-text>
                  <div class="text-caption text-grey">Buys / Sells</div>
                  <div class="text-h6">
                    <span class="text-success">{{ stats.total_buys || 0 }}</span> /
                    <span class="text-error">{{ stats.total_sells || 0 }}</span>
                  </div>
                </v-card-text>
              </v-card>
            </v-col>
          </v-row>

          <v-list density="compact" class="mt-4">
            <v-list-item>
              <v-list-item-title>Largest Win</v-list-item-title>
              <template v-slot:append>
                <span class="text-success font-weight-bold">+${{ stats.largest_win?.toFixed(2) || '0.00' }}</span>
              </template>
            </v-list-item>
            <v-list-item>
              <v-list-item-title>Largest Loss</v-list-item-title>
              <template v-slot:append>
                <span class="text-error font-weight-bold">-${{ stats.largest_loss?.toFixed(2) || '0.00' }}</span>
              </template>
            </v-list-item>
          </v-list>
        </div>
      </div>
    </v-card-text>
  </v-card>
</template>

<script>
import { watch, onMounted } from 'vue'

export default {
  name: 'PerformanceStats',
  props: {
    stats: {
      type: Object,
      default: () => ({})
    },
    portfolioStats: {
      type: Object,
      default: () => ({})
    },
    strategy: {
      type: String,
      default: ''
    },
    currentPrice: {
      type: Number,
      default: 0
    },
    priceChange24h: {
      type: Number,
      default: 0
    },
    priceChangePercent24h: {
      type: Number,
      default: 0
    },
    isPriceUp: {
      type: Boolean,
      default: true
    },
    isPricePulsing: {
      type: Boolean,
      default: false
    }
  },
  setup(props) {
    // Debug logging to verify strategy detection
    onMounted(() => {
      console.log('📊 PerformanceStats mounted - strategy:', props.strategy)
      console.log('📊 Portfolio stats:', props.portfolioStats)
      console.log('📊 Trade stats:', props.stats)
    })

    watch(() => props.strategy, (newStrategy) => {
      console.log('📊 Strategy changed to:', newStrategy)
    })

    watch(() => props.portfolioStats, (newStats) => {
      console.log('📊 Portfolio stats updated:', newStats)
      if (newStats && newStats.total_holdings) {
        console.log('  💰 Holdings:', newStats.total_holdings, 'BTC')
        console.log('  💵 Total Cost:', '$' + newStats.total_cost?.toFixed(2))
        console.log('  📈 Current Price:', '$' + newStats.current_price?.toFixed(2))
        console.log('  💸 Current Value:', '$' + newStats.current_value?.toFixed(2))
        console.log('  🎯 Unrealized Gain:', '$' + newStats.unrealized_gain?.toFixed(2))
        console.log('  📊 ROI:', newStats.unrealized_roi?.toFixed(2) + '%')
      }
    }, { deep: true })

    return {}
  }
}
</script>
