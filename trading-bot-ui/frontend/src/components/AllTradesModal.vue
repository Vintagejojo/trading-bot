<template>
  <v-dialog v-model="dialog" max-width="1200px" scrollable>
    <v-card>
      <v-card-title class="d-flex align-center justify-space-between bg-surface-variant">
        <div class="d-flex align-center">
          <v-icon icon="mdi-history" class="mr-2"></v-icon>
          Lifetime Trade History
        </div>
        <v-btn icon="mdi-close" variant="text" @click="dialog = false"></v-btn>
      </v-card-title>

      <!-- Filters -->
      <v-card-text class="pa-4">
        <v-row dense>
          <v-col cols="12" md="3">
            <v-select
              v-model="filters.side"
              :items="['All', 'BUY', 'SELL']"
              label="Side"
              density="compact"
              variant="outlined"
            ></v-select>
          </v-col>
          <v-col cols="12" md="3">
            <v-select
              v-model="filters.strategy"
              :items="strategyOptions"
              label="Strategy"
              density="compact"
              variant="outlined"
            ></v-select>
          </v-col>
          <v-col cols="12" md="3">
            <v-select
              v-model="filters.mode"
              :items="['All', 'PAPER', 'LIVE']"
              label="Mode"
              density="compact"
              variant="outlined"
            ></v-select>
          </v-col>
          <v-col cols="12" md="3">
            <v-btn
              block
              color="primary"
              prepend-icon="mdi-download"
              @click="exportToCSV"
              :loading="exporting"
            >
              Export CSV
            </v-btn>
          </v-col>
        </v-row>
      </v-card-text>

      <v-divider></v-divider>

      <!-- Trade Table -->
      <v-card-text style="max-height: 600px">
        <div v-if="filteredTrades.length === 0" class="text-center py-8 text-grey">
          <v-icon icon="mdi-filter-off" size="64" class="mb-2"></v-icon>
          <p>No trades match the filters</p>
        </div>

        <v-table v-else density="compact" fixed-header>
          <thead>
            <tr>
              <th>Date/Time</th>
              <th>Side</th>
              <th class="text-right">Price</th>
              <th class="text-right">Quantity</th>
              <th class="text-right">Total</th>
              <th class="text-right">P/L</th>
              <th>Strategy</th>
              <th class="text-center">Mode</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="trade in filteredTrades" :key="trade.id">
              <td class="text-caption">{{ formatTime(trade.timestamp) }}</td>
              <td>
                <v-chip
                  :color="trade.side === 'BUY' ? 'success' : 'error'"
                  size="x-small"
                  variant="flat"
                >
                  {{ trade.side }}
                </v-chip>
              </td>
              <td class="text-right font-mono text-caption">${{ trade.price.toFixed(8) }}</td>
              <td class="text-right font-mono text-caption">{{ formatNumber(trade.quantity) }}</td>
              <td class="text-right font-mono text-caption">${{ trade.total.toFixed(2) }}</td>
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
      </v-card-text>

      <!-- Summary Footer -->
      <v-divider></v-divider>
      <v-card-actions class="justify-space-between pa-4">
        <div class="text-caption">
          <strong>{{ filteredTrades.length }}</strong> of <strong>{{ trades.length }}</strong> trades shown
        </div>
        <v-chip
          :color="totalProfitLoss >= 0 ? 'success' : 'error'"
          variant="flat"
        >
          Total P/L: {{ totalProfitLoss >= 0 ? '+' : '' }}${{ totalProfitLoss.toFixed(2) }}
        </v-chip>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import { ref, computed, watch } from 'vue'
import { ExportTradesToCSV } from '../../wailsjs/go/main/App'

export default {
  name: 'AllTradesModal',
  props: {
    modelValue: {
      type: Boolean,
      required: true
    },
    trades: {
      type: Array,
      required: true
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const dialog = computed({
      get: () => props.modelValue,
      set: (val) => emit('update:modelValue', val)
    })

    const exporting = ref(false)

    const filters = ref({
      side: 'All',
      strategy: 'All',
      mode: 'All'
    })

    // Get unique strategies from trades
    const strategyOptions = computed(() => {
      const strategies = new Set(props.trades.map(t => t.strategy))
      return ['All', ...Array.from(strategies)]
    })

    // Filter trades based on selected filters
    const filteredTrades = computed(() => {
      return props.trades.filter(trade => {
        if (filters.value.side !== 'All' && trade.side !== filters.value.side) return false
        if (filters.value.strategy !== 'All' && trade.strategy !== filters.value.strategy) return false
        if (filters.value.mode !== 'All') {
          const isPaper = filters.value.mode === 'PAPER'
          if (trade.paper_trade !== isPaper) return false
        }
        return true
      })
    })

    const totalProfitLoss = computed(() => {
      return filteredTrades.value
        .filter(t => t.side === 'SELL')
        .reduce((sum, t) => sum + (t.profit_loss || 0), 0)
    })

    const formatTime = (timestamp) => {
      const date = new Date(timestamp)
      return date.toLocaleString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    }

    const formatNumber = (num) => {
      if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
      if (num >= 1000) return (num / 1000).toFixed(2) + 'K'
      return num.toFixed(4)
    }

    const exportToCSV = async () => {
      exporting.value = true
      try {
        const csv = await ExportTradesToCSV()
        // Create blob and download
        const blob = new Blob([csv], { type: 'text/csv' })
        const url = window.URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = `trades_${new Date().toISOString().split('T')[0]}.csv`
        a.click()
        window.URL.revokeObjectURL(url)
      } catch (error) {
        console.error('Failed to export trades:', error)
        alert('Failed to export trades: ' + error)
      } finally {
        exporting.value = false
      }
    }

    return {
      dialog,
      filters,
      strategyOptions,
      filteredTrades,
      totalProfitLoss,
      formatTime,
      formatNumber,
      exportToCSV,
      exporting
    }
  }
}
</script>
