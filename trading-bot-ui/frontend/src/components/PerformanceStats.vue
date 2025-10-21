<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <v-icon icon="mdi-chart-line" class="mr-2" color="success"></v-icon>
      Performance
    </v-card-title>

    <v-card-text>
      <div v-if="stats.total_trades === 0" class="text-center py-8 text-grey">
        <v-icon icon="mdi-chart-box-outline" size="64" class="mb-2"></v-icon>
        <p>No trading data yet</p>
        <p class="text-caption">Start the bot to see performance metrics</p>
      </div>

      <div v-else>
        <v-card :color="stats.total_profit_loss >= 0 ? 'success' : 'error'" variant="flat" class="mb-4">
          <v-card-text>
            <div class="text-caption">Total Profit/Loss</div>
            <div class="text-h4 font-weight-bold">
              {{ stats.total_profit_loss >= 0 ? '+' : '' }}${{ stats.total_profit_loss?.toFixed(2) || '0.00' }}
            </div>
          </v-card-text>
        </v-card>

        <v-row>
          <v-col cols="6">
            <v-card variant="outlined">
              <v-card-text>
                <div class="text-caption text-grey">Total Trades</div>
                <div class="text-h6">{{ stats.total_trades || 0 }}</div>
              </v-card-text>
            </v-card>
          </v-col>

          <v-col cols="6">
            <v-card variant="outlined">
              <v-card-text>
                <div class="text-caption text-grey">Win Rate</div>
                <div class="text-h6" :class="(stats.win_rate || 0) >= 50 ? 'text-success' : 'text-error'">
                  {{ stats.win_rate?.toFixed(1) || '0.0' }}%
                </div>
              </v-card-text>
            </v-card>
          </v-col>

          <v-col cols="6">
            <v-card variant="outlined">
              <v-card-text>
                <div class="text-caption text-grey">Avg P/L</div>
                <div class="text-h6" :class="(stats.average_profit_loss || 0) >= 0 ? 'text-success' : 'text-error'">
                  {{ (stats.average_profit_loss || 0) >= 0 ? '+' : '' }}${{ stats.average_profit_loss?.toFixed(2) || '0.00' }}
                </div>
              </v-card-text>
            </v-card>
          </v-col>

          <v-col cols="6">
            <v-card variant="outlined">
              <v-card-text>
                <div class="text-caption text-grey">Buys / Sells</div>
                <div class="text-h6">
                  <span class="text-success">{{ stats.total_buys || 0 }}</span> /
                  <span class="text-error">{{ stats.total_sells || 0 }}</span>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>

        <v-list density="compact" class="mt-4">
          <v-list-item>
            <v-list-item-title>Largest Win</v-list-item-title>
            <template v-slot:append>
              <span class="text-success font-weight-bold">+${{ stats.largest_win?.toFixed(2) || '0.00' }}</span>
            </template>
          </v-list-item>
          <v-list-item>
            <v-list-item-title>Largest Loss</v-list-item-title>
            <template v-slot:append>
              <span class="text-error font-weight-bold">-${{ stats.largest_loss?.toFixed(2) || '0.00' }}</span>
            </template>
          </v-list-item>
        </v-list>
      </div>
    </v-card-text>
  </v-card>
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
