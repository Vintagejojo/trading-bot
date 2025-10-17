<template>
  <div class="bg-gray-800 rounded-lg p-4 border border-gray-700">
    <h2 class="text-lg font-semibold text-gray-100 mb-3">🔔 Live Activity</h2>

    <div class="bg-gray-900 rounded p-3 h-64 overflow-y-auto font-mono text-sm">
      <div v-if="logs.length === 0" class="text-gray-500 text-center py-8">
        Waiting for activity...
      </div>

      <div
        v-for="(log, index) in logs"
        :key="index"
        class="mb-2 pb-2 border-b border-gray-800 last:border-0"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <span :class="getLogColor(log.type)" class="mr-2">{{ getLogIcon(log.type) }}</span>
            <span class="text-gray-300">{{ log.message }}</span>
          </div>
          <span class="text-gray-600 text-xs ml-2">{{ formatTime(log.time) }}</span>
        </div>

        <!-- Additional data for trade events -->
        <div v-if="log.data && log.data.side" class="ml-6 mt-1 text-xs">
          <span class="text-gray-500">
            {{ log.data.side }}: {{ log.data.quantity }} @ ${{ log.data.price?.toFixed(8) }}
            <span v-if="log.data.profitLoss" :class="log.data.profitLoss > 0 ? 'text-green-400' : 'text-red-400'">
              ({{ log.data.profitPercent?.toFixed(2) }}%, ${{ log.data.profitLoss?.toFixed(2) }})
            </span>
          </span>
        </div>

        <!-- Additional data for indicator events -->
        <div v-if="log.data && log.data.values" class="ml-6 mt-1 text-xs text-gray-500">
          {{ JSON.stringify(log.data.values) }}
        </div>
      </div>
    </div>

    <div class="mt-2 flex justify-between items-center text-xs text-gray-500">
      <span>{{ logs.length }} events</span>
      <button
        @click="clearLogs"
        class="text-blue-400 hover:text-blue-300"
      >
        Clear
      </button>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'

export default {
  name: 'ActivityLog',
  setup() {
    const logs = ref([])
    const maxLogs = 100

    const addLog = (type, message, data = {}) => {
      logs.value.unshift({
        type,
        message,
        data,
        time: new Date()
      })

      // Keep only last maxLogs entries
      if (logs.value.length > maxLogs) {
        logs.value = logs.value.slice(0, maxLogs)
      }
    }

    const getLogColor = (type) => {
      const colors = {
        'bot:connected': 'text-green-400',
        'bot:candle': 'text-blue-400',
        'bot:indicator': 'text-purple-400',
        'bot:trade': 'text-yellow-400',
        'bot:status': 'text-gray-400',
        'bot:error': 'text-red-400',
      }
      return colors[type] || 'text-gray-400'
    }

    const getLogIcon = (type) => {
      const icons = {
        'bot:connected': '✅',
        'bot:candle': '📊',
        'bot:indicator': '📈',
        'bot:trade': '💰',
        'bot:status': '⏳',
        'bot:error': '❌',
      }
      return icons[type] || '•'
    }

    const formatTime = (date) => {
      return date.toLocaleTimeString()
    }

    const clearLogs = () => {
      logs.value = []
    }

    onMounted(() => {
      // Listen for all bot events
      EventsOn('bot:connected', (data) => {
        addLog('bot:connected', data.message, data.data)
      })

      EventsOn('bot:candle', (data) => {
        addLog('bot:candle', data.message, data.data)
      })

      EventsOn('bot:indicator', (data) => {
        addLog('bot:indicator', data.message, data.data)
      })

      EventsOn('bot:trade', (data) => {
        addLog('bot:trade', data.message, data.data)
      })

      EventsOn('bot:status', (data) => {
        addLog('bot:status', data.message, data.data)
      })

      EventsOn('bot:error', (data) => {
        addLog('bot:error', typeof data === 'string' ? data : data.message, typeof data === 'object' ? data.data : {})
      })
    })

    return {
      logs,
      getLogColor,
      getLogIcon,
      formatTime,
      clearLogs
    }
  }
}
</script>
