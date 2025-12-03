<template>
  <v-card>
    <v-card-title class="d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <v-icon icon="mdi-bell" class="mr-2" color="accent"></v-icon>
        Live Activity
      </div>
      <v-chip size="small" variant="outlined">{{ logs.length }} events</v-chip>
    </v-card-title>

    <v-card-text>
      <div v-if="logs.length === 0" class="text-center py-8 text-grey">
        <v-icon icon="mdi-information-outline" size="48" class="mb-2"></v-icon>
        <p>Waiting for activity...</p>
        <p class="text-caption">System ready</p>
      </div>

      <v-list v-else density="compact" max-height="400" class="overflow-y-auto">
        <v-list-item
          v-for="(log, index) in logs"
          :key="index"
          :title="log.message"
          :subtitle="formatTime(log.time)"
        >
          <template v-slot:prepend>
            <v-icon :color="getLogColor(log.type)">{{ getLogIcon(log.type) }}</v-icon>
          </template>
        </v-list-item>
      </v-list>

      <v-btn
        v-if="logs.length > 0"
        block
        size="small"
        variant="outlined"
        @click="clearLogs"
        class="mt-2"
      >
        Clear
      </v-btn>
    </v-card-text>
  </v-card>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'

export default {
  name: 'ActivityLog',
  setup() {
    const logs = ref([])
    const maxLogs = 100

    const addLog = (type, message, data = {}) => {
      logs.value.unshift({ type, message, data, time: new Date() })
      if (logs.value.length > maxLogs) {
        logs.value = logs.value.slice(0, maxLogs)
      }
    }

    const getLogColor = (type) => {
      const colors = {
        'bot:connected': 'success',
        'bot:candle': 'info',
        'bot:indicator': 'accent',
        'bot:trade': 'warning',
        'bot:status': 'grey',
        'bot:error': 'error',
      }
      return colors[type] || 'grey'
    }

    const getLogIcon = (type) => {
      const icons = {
        'bot:connected': 'mdi-check-circle',
        'bot:candle': 'mdi-chart-candlestick',
        'bot:indicator': 'mdi-chart-line',
        'bot:trade': 'mdi-currency-usd',
        'bot:status': 'mdi-clock-outline',
        'bot:error': 'mdi-alert-circle',
      }
      return icons[type] || 'mdi-circle-small'
    }

    const formatTime = (date) => {
      return date.toLocaleTimeString()
    }

    const clearLogs = () => {
      logs.value = []
    }

    onMounted(() => {
      console.log('[ActivityLog] Component mounted, using custom event bridge')

      // Listen to custom events forwarded from App.vue
      const handleActivityLog = (event) => {
        const { type, data } = event.detail
        const msg = typeof data === 'string' ? data : (data?.message || 'Update')
        const eventData = typeof data === 'object' ? data.data : {}
        addLog(type, msg, eventData)
      }

      window.addEventListener('activity-log', handleActivityLog)

      // Also listen to bot:started and bot:stopped directly
      const handleBotStarted = (event) => {
        const strategy = event.detail || 'unknown'
        addLog('bot:connected', `Bot started with ${strategy} strategy`, { strategy })
      }

      const handleBotStopped = () => {
        addLog('bot:status', 'Bot stopped', {})
      }

      window.addEventListener('bot-started', handleBotStarted)
      window.addEventListener('bot-stopped', handleBotStopped)

      // Cleanup
      onUnmounted(() => {
        window.removeEventListener('activity-log', handleActivityLog)
        window.removeEventListener('bot-started', handleBotStarted)
        window.removeEventListener('bot-stopped', handleBotStopped)
      })

      console.log('[ActivityLog] Event bridge initialized')
    })

    return { logs, getLogColor, getLogIcon, formatTime, clearLogs }
  }
}
</script>
