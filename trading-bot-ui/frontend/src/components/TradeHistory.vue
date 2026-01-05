<template>
  <v-card>
    <v-card-title class="d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <v-icon icon="mdi-history" class="mr-2" color="info"></v-icon>
        Trade History
      </div>
      <div class="d-flex align-center gap-2">
        <v-chip v-if="trades.length > 0" size="small" variant="outlined">
          {{ displayedTrades.length }} of {{ trades.length }} trades
        </v-chip>
        <v-btn
          v-if="trades.length > maxDisplay"
          icon="mdi-open-in-new"
          variant="text"
          size="small"
          @click="$emit('view-all')"
        ></v-btn>
      </div>
    </v-card-title>

    <v-card-text>
      <div v-if="trades.length === 0" class="text-center py-8 text-grey">
        <v-icon icon="mdi-file-document-outline" size="64" class="mb-2"></v-icon>
        <p>No trades yet</p>
        <p class="text-caption">Start the bot to begin trading</p>
      </div>

      <v-table v-else density="comfortable">
        <thead>
          <tr>
            <th>Time</th>
            <th>Side</th>
            <th class="text-right">Price</th>
            <th class="text-right">Quantity</th>
            <th class="text-right">P/L</th>
            <th>Strategy</th>
            <th class="text-center">Mode</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="trade in displayedTrades" :key="trade.id">
            <td class="text-caption">{{ formatTime(trade.timestamp) }}</td>
            <td>
              <v-chip
                :color="trade.side === 'BUY' ? 'success' : 'error'"
                size="small"
                variant="flat"
              >
                {{ trade.side }}
              </v-chip>
            </td>
            <td class="text-right font-mono text-caption">${{ trade.price.toFixed(8) }}</td>
            <td class="text-right font-mono text-caption">{{ formatNumber(trade.quantity) }}</td>
            <td class="text-right">
              <span
                v-if="trade.side === 'SELL'"
                :class="trade.profit_loss >= 0 ? 'text-success' : 'text-error'"
                class="font-weight-bold text-caption"
              >
                {{ trade.profit_loss >= 0 ? '+' : '' }}${{ trade.profit_loss.toFixed(2) }}
                ({{ trade.profit_loss_percent.toFixed(2) }}%)
              </span>
              <span v-else class="text-grey text-caption">-</span>
            </td>
            <td>
              <v-chip size="x-small" variant="outlined">{{ trade.strategy }}</v-chip>
            </td>
            <td class="text-center">
              <v-chip
                :color="trade.paper_trade ? 'info' : 'error'"
                size="x-small"
                variant="flat"
              >
                {{ trade.paper_trade ? 'PAPER' : 'LIVE' }}
              </v-chip>
            </td>
          </tr>
        </tbody>
      </v-table>

      <!-- Total P/L - Only show for trading strategies with SELL trades -->
      <v-divider v-if="trades.length > 0 && hasSellTrades" class="my-4"></v-divider>

      <div v-if="trades.length > 0 && hasSellTrades" class="d-flex justify-space-between align-center">
        <span class="text-caption text-grey">Total P/L:</span>
        <v-chip
          :color="totalProfitLoss >= 0 ? 'success' : 'error'"
          variant="flat"
          class="neon-hover-primary"
        >
          <span
            class="crypto-mono"
            :class="totalProfitLoss >= 0 ? 'neon-glow-success' : 'neon-glow-error'"
          >
            {{ totalProfitLoss >= 0 ? '+' : '' }}${{ totalProfitLoss.toFixed(2) }}
          </span>
        </v-chip>
      </div>

      <!-- DCA Info - Show for accumulation-only strategies -->
      <div v-if="trades.length > 0 && !hasSellTrades" class="text-center py-2">
        <v-chip variant="outlined" color="info" size="small">
          <v-icon icon="mdi-trending-up" size="small" class="mr-1"></v-icon>
          DCA Accumulation Strategy - See Performance box for unrealized gains
        </v-chip>
      </div>
    </v-card-text>
  </v-card>
</template>

<script>
import { computed } from 'vue'

export default {
  name: 'TradeHistory',
  props: {
    trades: {
      type: Array,
      required: true
    }
  },
  emits: ['view-all'],
  setup(props) {
    const maxDisplay = 50

    // Only show last 50 trades in main panel
    const displayedTrades = computed(() => {
      return props.trades.slice(0, maxDisplay)
    })
    const formatTime = (timestamp) => {
      const date = new Date(timestamp)
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    }

    const formatNumber = (num) => {
      // Format large numbers with K/M suffix
      if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
      if (num >= 1000) return (num / 1000).toFixed(2) + 'K'

      // Format crypto quantities (small decimals) with appropriate precision
      if (num < 0.01) return num.toFixed(8)  // BTC amounts: 0.00105143
      if (num < 1) return num.toFixed(6)     // Small amounts: 0.123456

      return num.toFixed(2)  // Regular amounts: 123.45
    }

    const totalProfitLoss = computed(() => {
      return props.trades
        .filter(t => t.side === 'SELL')
        .reduce((sum, t) => sum + (t.profit_loss || 0), 0)
    })

    // Check if there are any SELL trades (for showing Total P/L)
    const hasSellTrades = computed(() => {
      return props.trades.some(t => t.side === 'SELL')
    })

    return {
      formatTime,
      formatNumber,
      totalProfitLoss,
      hasSellTrades,
      displayedTrades,
      maxDisplay
    }
  }
}
</script>
