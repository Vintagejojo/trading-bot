<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <v-icon icon="mdi-wallet" class="mr-2" color="warning"></v-icon>
      Current Position
    </v-card-title>

    <v-card-text>
      <div v-if="!position || !position.is_open" class="text-center py-8 text-grey">
        <v-icon icon="mdi-briefcase-outline" size="64" class="mb-2"></v-icon>
        <p>No open position</p>
        <p class="text-caption">Waiting for trade signal...</p>
      </div>

      <div v-else>
        <v-list>
          <v-list-item>
            <v-list-item-title>Symbol</v-list-item-title>
            <template v-slot:append>
              <strong>{{ position.symbol }}</strong>
            </template>
          </v-list-item>

          <v-list-item>
            <v-list-item-title>Quantity</v-list-item-title>
            <template v-slot:append>
              <span class="font-mono">{{ formatNumber(position.quantity) }}</span>
            </template>
          </v-list-item>

          <v-list-item>
            <v-list-item-title>Entry Price</v-list-item-title>
            <template v-slot:append>
              <span class="font-mono">${{ position.entry_price?.toFixed(8) }}</span>
            </template>
          </v-list-item>

          <v-list-item>
            <v-list-item-title>Entry Time</v-list-item-title>
            <template v-slot:append>
              <span class="text-caption">{{ formatDateTime(position.entry_time) }}</span>
            </template>
          </v-list-item>

          <v-list-item>
            <v-list-item-title>Strategy</v-list-item-title>
            <template v-slot:append>
              <v-chip size="small" color="primary">{{ position.strategy }}</v-chip>
            </template>
          </v-list-item>
        </v-list>

        <v-alert color="success" variant="tonal" class="mt-4">
          <v-icon icon="mdi-check-circle" start></v-icon>
          Position Open
        </v-alert>
      </div>
    </v-card-text>
  </v-card>
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
      if (num >= 1000000) return (num / 1000000).toFixed(2) + 'M'
      if (num >= 1000) return (num / 1000).toFixed(2) + 'K'
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

    return { formatNumber, formatDateTime }
  }
}
</script>
