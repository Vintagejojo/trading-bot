<template>
  <div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
    <h2 class="text-xl font-semibold mb-4 text-blue-400">Trade History</h2>

    <div v-if="trades.length === 0" class="text-center py-8 text-gray-400">
      <p>No trades yet</p>
      <p class="text-sm mt-2">Start the bot to begin trading</p>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-700">
            <th class="text-left py-3 px-2 text-gray-400 font-medium">Time</th>
            <th class="text-left py-3 px-2 text-gray-400 font-medium">Side</th>
            <th class="text-right py-3 px-2 text-gray-400 font-medium">Price</th>
            <th class="text-right py-3 px-2 text-gray-400 font-medium">Quantity</th>
            <th class="text-right py-3 px-2 text-gray-400 font-medium">P/L</th>
            <th class="text-left py-3 px-2 text-gray-400 font-medium">Strategy</th>
            <th class="text-center py-3 px-2 text-gray-400 font-medium">Mode</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="trade in trades" :key="trade.id"
              class="border-b border-gray-700 hover:bg-gray-750 transition-colors">
            <td class="py-3 px-2 text-gray-300">
              {{ formatTime(trade.timestamp) }}
            </td>
            <td class="py-3 px-2">
              <span :class="trade.side === 'BUY' ? 'text-green-400' : 'text-red-400'"
                    class="font-semibold">
                {{ trade.side }}
              </span>
            </td>
            <td class="py-3 px-2 text-right text-gray-300 font-mono">
              {{ trade.price.toFixed(8) }}
            </td>
            <td class="py-3 px-2 text-right text-gray-300 font-mono">
              {{ formatNumber(trade.quantity) }}
            </td>
            <td class="py-3 px-2 text-right font-mono">
              <span v-if="trade.side === 'SELL'"
                    :class="trade.profit_loss >= 0 ? 'text-green-400' : 'text-red-400'">
                {{ trade.profit_loss >= 0 ? '+' : '' }}{{ trade.profit_loss.toFixed(2) }}
                ({{ trade.profit_loss_percent.toFixed(2) }}%)
              </span>
              <span v-else class="text-gray-500">-</span>
            </td>
            <td class="py-3 px-2 text-gray-300">
              <span class="text-xs bg-gray-700 px-2 py-1 rounded uppercase">
                {{ trade.strategy }}
              </span>
            </td>
            <td class="py-3 px-2 text-center">
              <span v-if="trade.paper_trade"
                    class="text-xs bg-blue-900 text-blue-300 px-2 py-1 rounded">
                PAPER
              </span>
              <span v-else
                    class="text-xs bg-red-900 text-red-300 px-2 py-1 rounded">
                LIVE
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Summary Row -->
    <div v-if="trades.length > 0" class="mt-4 pt-4 border-t border-gray-700">
      <div class="flex justify-between items-center text-sm">
        <span class="text-gray-400">Showing {{ trades.length }} trades</span>
        <div class="flex space-x-4">
          <span class="text-gray-400">
            Total P/L:
            <span :class="totalProfitLoss >= 0 ? 'text-green-400' : 'text-red-400'"
                  class="font-semibold ml-1">
              {{ totalProfitLoss >= 0 ? '+' : '' }}{{ totalProfitLoss.toFixed(2) }}
            </span>
          </span>
        </div>
      </div>
    </div>
  </div>
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
      if (num >= 1000000) {
        return (num / 1000000).toFixed(2) + 'M'
      } else if (num >= 1000) {
        return (num / 1000).toFixed(2) + 'K'
      }
      return num.toFixed(2)
    }

    const totalProfitLoss = computed(() => {
      return props.trades
        .filter(t => t.side === 'SELL')
        .reduce((sum, t) => sum + (t.profit_loss || 0), 0)
    })

    return {
      formatTime,
      formatNumber,
      totalProfitLoss
    }
  }
}
</script>
