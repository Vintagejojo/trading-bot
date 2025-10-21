<template>
  <v-card>
    <v-card-title class="d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <v-icon icon="mdi-history" class="mr-2" color="info"></v-icon>
        Trade History
      </div>
      <v-chip v-if="trades.length > 0" size="small" variant="outlined">
        {{ trades.length }} trades
      </v-chip>
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
          <tr v-for="trade in trades" :key="trade.id">
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

      <v-divider v-if="trades.length > 0" class="my-4"></v-divider>

      <div v-if="trades.length > 0" class="d-flex justify-space-between align-center">
        <span class="text-caption text-grey">Total P/L:</span>
        <v-chip
          :color="totalProfitLoss >= 0 ? 'success' : 'error'"
          variant="flat"
        >
          {{ totalProfitLoss >= 0 ? '+' : '' }}${{ totalProfitLoss.toFixed(2) }}
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
  setup(props) {
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
      if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
      if (num >= 1000) return (num / 1000).toFixed(2) + 'K'
      return num.toFixed(2)
    }

    const totalProfitLoss = computed(() => {
      return props.trades
        .filter(t => t.side === 'SELL')
        .reduce((sum, t) => sum + (t.profit_loss || 0), 0)
    })

    return { formatTime, formatNumber, totalProfitLoss }
  }
}
</script>
