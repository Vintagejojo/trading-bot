<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <v-icon icon="mdi-wallet" class="mr-2" color="warning"></v-icon>
      Wallet Balance
      <v-spacer></v-spacer>
      <v-btn
        icon="mdi-refresh"
        size="small"
        variant="text"
        @click="loadBalance"
        :loading="loading"
      ></v-btn>
    </v-card-title>

    <v-card-text>
      <div v-if="error" class="text-center py-4">
        <v-alert type="error" variant="tonal" class="mb-2">
          {{ error }}
        </v-alert>
        <v-btn size="small" @click="loadBalance">Retry</v-btn>
      </div>

      <div v-else-if="loading && !balances.length" class="text-center py-8">
        <v-progress-circular indeterminate color="primary"></v-progress-circular>
        <p class="text-caption text-grey mt-2">Loading wallet...</p>
      </div>

      <div v-else-if="!balances.length" class="text-center py-4 text-grey">
        <v-icon icon="mdi-wallet-outline" size="48" class="mb-2"></v-icon>
        <p>No balance data</p>
      </div>

      <div v-else>
        <!-- Total Value -->
        <v-card variant="outlined" class="mb-4" color="success">
          <v-card-text>
            <div class="text-caption text-grey">Total Balance (USDT)</div>
            <div class="text-h5 font-weight-bold">
              ${{ totalBalanceUSDT.toFixed(2) }}
            </div>
          </v-card-text>
        </v-card>

        <!-- Asset List -->
        <v-list density="compact">
          <v-list-subheader>Assets</v-list-subheader>
          <v-list-item
            v-for="balance in filteredBalances"
            :key="balance.asset"
          >
            <template v-slot:prepend>
              <v-avatar size="32" color="primary">
                <span class="text-caption">{{ balance.asset.substring(0, 2) }}</span>
              </v-avatar>
            </template>

            <v-list-item-title>{{ balance.asset }}</v-list-item-title>
            <v-list-item-subtitle>
              Free: {{ parseFloat(balance.free).toFixed(8) }}
            </v-list-item-subtitle>

            <template v-slot:append>
              <div class="text-right">
                <div class="font-weight-bold">{{ parseFloat(balance.free).toFixed(4) }}</div>
                <div class="text-caption text-grey" v-if="balance.locked > 0">
                  Locked: {{ parseFloat(balance.locked).toFixed(4) }}
                </div>
              </div>
            </template>
          </v-list-item>
        </v-list>

        <!-- Last Updated -->
        <div class="text-caption text-grey text-center mt-4">
          Last updated: {{ lastUpdated }}
        </div>
      </div>
    </v-card-text>
  </v-card>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { GetWalletBalance } from '../../wailsjs/go/main/App'

export default {
  name: 'WalletBalance',
  setup() {
    const balances = ref([])
    const loading = ref(false)
    const error = ref('')
    const lastUpdated = ref('')

    const loadBalance = async () => {
      loading.value = true
      error.value = ''
      try {
        const result = await GetWalletBalance()
        balances.value = result || []
        lastUpdated.value = new Date().toLocaleTimeString()
      } catch (err) {
        console.error('Failed to load wallet balance:', err)
        error.value = err.toString() || 'Failed to load wallet balance'
      } finally {
        loading.value = false
      }
    }

    // Filter out zero balances and sort by value
    const filteredBalances = computed(() => {
      return balances.value
        .filter(b => parseFloat(b.free) > 0 || parseFloat(b.locked) > 0)
        .sort((a, b) => parseFloat(b.free) - parseFloat(a.free))
    })

    // Calculate total in USDT (simplified - just USDT assets)
    const totalBalanceUSDT = computed(() => {
      const usdtBalance = balances.value.find(b => b.asset === 'USDT')
      return usdtBalance ? parseFloat(usdtBalance.free) : 0
    })

    onMounted(() => {
      loadBalance()
    })

    return {
      balances,
      loading,
      error,
      lastUpdated,
      filteredBalances,
      totalBalanceUSDT,
      loadBalance
    }
  }
}
</script>
