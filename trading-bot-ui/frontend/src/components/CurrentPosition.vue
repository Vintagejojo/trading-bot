<template>
  <div class="bg-gray-800 rounded-lg p-6 border border-gray-700">
    <h2 class="text-xl font-semibold mb-4 text-blue-400">Current Position</h2>

    <div v-if="!position || !position.is_open" class="text-center py-6 text-gray-400">
      <div class="text-4xl mb-2">💼</div>
      <p class="text-sm">No open position</p>
    </div>

    <div v-else class="space-y-4">
      <!-- Position Info -->
      <div class="bg-gray-700 rounded p-4 space-y-3">
        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-400">Symbol</span>
          <span class="text-lg font-semibold text-gray-100">{{ position.symbol }}</span>
        </div>

        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-400">Quantity</span>
          <span class="text-lg font-mono text-gray-100">
            {{ formatNumber(position.quantity) }}
          </span>
        </div>

        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-400">Entry Price</span>
          <span class="text-lg font-mono text-gray-100">
            {{ position.entry_price?.toFixed(8) }}
          </span>
        </div>

        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-400">Entry Time</span>
          <span class="text-sm text-gray-300">
            {{ formatDateTime(position.entry_time) }}
          </span>
        </div>

        <div class="flex justify-between items-center">
          <span class="text-sm text-gray-400">Strategy</span>
          <span class="text-xs bg-blue-900 text-blue-300 px-2 py-1 rounded uppercase">
            {{ position.strategy }}
          </span>
        </div>
      </div>

      <!-- Status Badge -->
      <div class="flex items-center justify-center space-x-2 text-green-400">
        <span class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></span>
        <span class="text-sm font-medium">Position Open</span>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CurrentPosition',
  props: {
    position: {
      type: Object,
      default: null
    }
  },
  setup() {
    const formatNumber = (num) => {
      if (!num) return '0'
      if (num >= 1000000) {
        return (num / 1000000).toFixed(2) + 'M'
      } else if (num >= 1000) {
        return (num / 1000).toFixed(2) + 'K'
      }
      return num.toLocaleString()
    }

    const formatDateTime = (timestamp) => {
      if (!timestamp) return '-'
      const date = new Date(timestamp)
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      })
    }

    return {
      formatNumber,
      formatDateTime
    }
  }
}
</script>
