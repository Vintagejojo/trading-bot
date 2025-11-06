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
import { ref, onMounted } from 'vue'
import { EventsOn } from '../../wailsjs/runtime/runtime'

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
      // Handle events - data can be either {message, data} object or a plain string
      EventsOn('bot:connected', (data) => {
        const msg = typeof data === 'string' ? data : (data?.message || 'Connected')
        const eventData = typeof data === 'object' ? data.data : {}
        addLog('bot:connected', msg, eventData)
      })

      EventsOn('bot:candle', (data) => {
        const msg = typeof data === 'string' ? data : (data?.message || 'Candle updated')
        const eventData = typeof data === 'object' ? data.data : {}
        addLog('bot:candle', msg, eventData)
      })

      EventsOn('bot:indicator', (data) => {
        const msg = typeof data === 'string' ? data : (data?.message || 'Indicator updated')
        const eventData = typeof data === 'object' ? data.data : {}
        addLog('bot:indicator', msg, eventData)
      })

      EventsOn('bot:trade', (data) => {
        const msg = typeof data === 'string' ? data : (data?.message || 'Trade executed')
        const eventData = typeof data === 'object' ? data.data : {}
        addLog('bot:trade', msg, eventData)
      })

      EventsOn('bot:status', (data) => {
        const msg = typeof data === 'string' ? data : (data?.message || 'Status update')
        const eventData = typeof data === 'object' ? data.data : {}
        addLog('bot:status', msg, eventData)
      })

      EventsOn('bot:error', (data) => {
        const msg = typeof data === 'string' ? data : (data?.message || 'Error occurred')
        const eventData = typeof data === 'object' ? data.data : {}
        addLog('bot:error', msg, eventData)
      })

      EventsOn('bot:started', (strategy) => {
        addLog('bot:connected', `Bot started with ${strategy} strategy`, { strategy })
      })

      EventsOn('bot:stopped', () => {
        addLog('bot:status', 'Bot stopped', {})
      })
    })

    return { logs, getLogColor, getLogIcon, formatTime, clearLogs }
  }
}
</script>
