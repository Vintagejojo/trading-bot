<template>
  <div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
    <h2 class="text-xl font-semibold mb-4 text-blue-400">Performance</h2>

    <div v-if="stats.total_trades === 0" class="text-center py-4 text-gray-400">
      <p class="text-sm">No trading data yet</p>
    </div>

    <div v-else class="space-y-4">
      <!-- Total P/L -->
      <div class="bg-gray-700 rounded p-4">
        <div class="text-sm text-gray-400 mb-1">Total Profit/Loss</div>
        <div :class="stats.total_profit_loss >= 0 ? 'text-green-400' : 'text-red-400'"
             class="text-2xl font-bold">
          {{ stats.total_profit_loss >= 0 ? '+' : '' }}${{ stats.total_profit_loss?.toFixed(2) || '0.00' }}
        </div>
      </div>

      <!-- Stats Grid -->
      <div class="grid grid-cols-2 gap-3">
        <div class="bg-gray-700 rounded p-3">
          <div class="text-xs text-gray-400 mb-1">Total Trades</div>
          <div class="text-lg font-semibold text-gray-100">
            {{ stats.total_trades || 0 }}
          </div>
        </div>

        <div class="bg-gray-700 rounded p-3">
          <div class="text-xs text-gray-400 mb-1">Win Rate</div>
          <div class="text-lg font-semibold"
               :class="(stats.win_rate || 0) >= 50 ? 'text-green-400' : 'text-red-400'">
            {{ stats.win_rate?.toFixed(1) || '0.0' }}%
          </div>
        </div>

        <div class="bg-gray-700 rounded p-3">
          <div class="text-xs text-gray-400 mb-1">Avg P/L</div>
          <div class="text-lg font-semibold"
               :class="(stats.average_profit_loss || 0) >= 0 ? 'text-green-400' : 'text-red-400'">
            {{ (stats.average_profit_loss || 0) >= 0 ? '+' : '' }}${{ stats.average_profit_loss?.toFixed(2) || '0.00' }}
          </div>
        </div>

        <div class="bg-gray-700 rounded p-3">
          <div class="text-xs text-gray-400 mb-1">Buys / Sells</div>
          <div class="text-lg font-semibold text-gray-100">
            {{ stats.total_buys || 0 }} / {{ stats.total_sells || 0 }}
          </div>
        </div>
      </div>

      <!-- Best/Worst Trade -->
      <div class="space-y-2">
        <div class="flex justify-between items-center text-sm">
          <span class="text-gray-400">Largest Win</span>
          <span class="text-green-400 font-semibold">
            +${{ stats.largest_win?.toFixed(2) || '0.00' }}
          </span>
        </div>
        <div class="flex justify-between items-center text-sm">
          <span class="text-gray-400">Largest Loss</span>
          <span class="text-red-400 font-semibold">
            ${{ stats.largest_loss?.toFixed(2) || '0.00' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PerformanceStats',
  props: {
    stats: {
      type: Object,
      default: () => ({})
    }
  }
}
</script>
