<template>
  <v-app>
    <!-- Setup Wizard -->
    <SetupWizard v-if="!setupComplete" @setup-complete="handleSetupComplete" />

    <!-- PIN Lock Screen -->
    <PinLock v-else-if="isLocked" :hasPin="hasPin" @unlocked="handleUnlock" />

    <!-- Main App -->
    <template v-else>
      <!-- App Bar -->
      <v-app-bar color="surface" elevation="2">
        <v-app-bar-title>
          <v-icon icon="mdi-robot" class="mr-2" color="primary"></v-icon>
          Tradecraft
        </v-app-bar-title>

        <template v-slot:append>
          <!-- Live Price Ticker -->
          <div v-if="botStatus.running && currentPrice > 0" class="d-flex align-center mr-3">
            <v-divider vertical class="mr-3"></v-divider>
            <div class="text-caption text-grey mr-1">{{ botStatus.symbol?.replace('USDT', '/USDT') }}</div>
            <div
              class="text-subtitle-1 font-weight-bold crypto-mono mr-2"
              :class="[
                isPriceUp ? 'neon-glow-success' : 'neon-glow-error',
                isPricePulsing ? 'neon-pulse' : ''
              ]"
            >
              ${{ currentPrice.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}
            </div>
            <v-chip
              v-if="priceChangePercent24h !== 0"
              :color="priceChangePercent24h >= 0 ? 'success' : 'error'"
              size="small"
              variant="flat"
              class="neon-card-info"
            >
              <v-icon
                :icon="priceChangePercent24h >= 0 ? 'mdi-arrow-up' : 'mdi-arrow-down'"
                size="x-small"
                class="mr-1"
              ></v-icon>
              <span class="crypto-mono">{{ Math.abs(priceChangePercent24h).toFixed(2) }}%</span>
            </v-chip>
          </div>

          <v-chip
            v-if="botStatus.symbol && !botStatus.running"
            class="mr-2"
            variant="outlined"
            color="info"
          >
            {{ botStatus.symbol }}
          </v-chip>

          <v-chip
            v-if="botStatus.trading_mode"
            :color="botStatus.trading_mode === 'live' ? 'error' : 'info'"
            variant="flat"
            class="mr-2"
          >
            {{ botStatus.trading_mode === 'live' ? 'LIVE' : 'PAPER' }}
          </v-chip>

          <v-chip
            :color="botStatus.running ? 'success' : 'grey'"
            variant="flat"
            class="mr-2"
          >
            <v-icon
              :icon="botStatus.running ? 'mdi-circle' : 'mdi-circle-outline'"
              start
              size="small"
            ></v-icon>
            {{ botStatus.running ? 'Running' : 'Stopped' }}
          </v-chip>

          <v-menu>
            <template v-slot:activator="{ props }">
              <v-btn icon="mdi-cog" v-bind="props"></v-btn>
            </template>
            <v-list>
              <v-list-subheader>Account Settings</v-list-subheader>

              <SettingsDialog />

              <EmailSettings />

              <v-divider></v-divider>

              <v-list-subheader>Trade History Management</v-list-subheader>

              <v-list-item @click="showAllTradesModal = true">
                <template v-slot:prepend>
                  <v-icon icon="mdi-history"></v-icon>
                </template>
                <v-list-item-title>View Lifetime Trades</v-list-item-title>
                <v-list-item-subtitle>Full history with filters & export</v-list-item-subtitle>
              </v-list-item>

              <v-list-item @click="clearDemoTrades">
                <template v-slot:prepend>
                  <v-icon icon="mdi-delete-sweep" color="warning"></v-icon>
                </template>
                <v-list-item-title>Clear Demo Trades</v-list-item-title>
                <v-list-item-subtitle>Remove paper trading test data</v-list-item-subtitle>
              </v-list-item>

              <v-divider></v-divider>

              <v-list-item @click="resetAPIKeys">
                <template v-slot:prepend>
                  <v-icon icon="mdi-refresh"></v-icon>
                </template>
                <v-list-item-title>Reset API Keys</v-list-item-title>
                <v-list-item-subtitle>Clear credentials and restart setup</v-list-item-subtitle>
              </v-list-item>
            </v-list>
          </v-menu>
        </template>
      </v-app-bar>

      <!-- Main Content -->
      <v-main>
        <v-container fluid class="pa-4">
          <v-row>
            <!-- Left Column: Bot Controls -->
            <v-col cols="12" lg="4">
              <WalletBalance class="mb-4" />

              <BotControls
                :botStatus="botStatus"
                @start-bot="startBot"
                @stop-bot="stopBot"
                @refresh-data="refreshData"
              />

              <PerformanceStats
                :stats="stats"
                :portfolioStats="portfolioStats"
                :strategy="botStatus.strategy"
                :currentPrice="currentPrice"
                :priceChange24h="priceChange24h"
                :priceChangePercent24h="priceChangePercent24h"
                :isPriceUp="isPriceUp"
                :isPricePulsing="isPricePulsing"
                class="mt-4"
              />

              <CurrentPosition :position="botStatus.position" class="mt-4" />
            </v-col>

            <!-- Right Column: Charts & Activity -->
            <v-col cols="12" lg="8">
              <!-- Trading Chart - only show for multitimeframe strategy -->
              <TradingChart v-if="botStatus.running && botStatus.strategy === 'multitimeframe'" class="mb-4" />

              <ActivityLog class="mb-4" />

              <TradeHistory :trades="trades" @view-all="showAllTradesModal = true" />
            </v-col>
          </v-row>
        </v-container>
      </v-main>
    </template>
  </v-app>

  <!-- All Trades Modal -->
  <AllTradesModal v-model="showAllTradesModal" :trades="trades" />
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { GetBotStatus, StartBot, StopBot, GetTradeHistory, GetTradeSummary, GetPortfolioStats, IsLocked, HasPIN, IsSetupComplete, ResetSetup, ClearDemoTrades } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import SetupWizard from './components/SetupWizard.vue'
import PinLock from './components/PinLock.vue'
import BotControls from './components/BotControls.vue'
import PerformanceStats from './components/PerformanceStats.vue'
import CurrentPosition from './components/CurrentPosition.vue'
import ActivityLog from './components/ActivityLog.vue'
import TradeHistory from './components/TradeHistory.vue'
import AllTradesModal from './components/AllTradesModal.vue'
import WalletBalance from './components/WalletBalance.vue'
import SettingsDialog from './components/SettingsDialog.vue'
import EmailSettings from './components/EmailSettings.vue'
import TradingChart from './components/TradingChart.vue'

export default {
  name: 'App',
  components: {
    SetupWizard,
    PinLock,
    BotControls,
    PerformanceStats,
    CurrentPosition,
    ActivityLog,
    TradeHistory,
    AllTradesModal,
    WalletBalance,
    SettingsDialog,
    EmailSettings,
    TradingChart
  },
  setup() {
    const setupComplete = ref(true)
    const isLocked = ref(true)
    const hasPin = ref(false)
    const botStatus = ref({
      running: false,
      strategy: '',
      symbol: '',
      trading_mode: 'paper',
      position: null,
      last_trade: null
    })
    const trades = ref([])
    const stats = ref({})
    const portfolioStats = ref({})
    const showAllTradesModal = ref(false)

    // Live price tracking
    const currentPrice = ref(0)
    const priceChange24h = ref(0)
    const priceChangePercent24h = ref(0)
    const isPriceUp = ref(true)
    const prices24h = ref([]) // Store prices for 24h calculation
    const isPricePulsing = ref(false) // Track pulse animation state

    let refreshInterval = null

    const loadBotStatus = async () => {
      try {
        botStatus.value = await GetBotStatus()
      } catch (error) {
        console.error('Failed to load bot status:', error)
      }
    }

    const loadTradeHistory = async () => {
      try {
        trades.value = await GetTradeHistory(50)
      } catch (error) {
        console.error('Failed to load trades:', error)
      }
    }

    const loadStats = async () => {
      try {
        stats.value = await GetTradeSummary()
        console.log('📊 Stats loaded:', stats.value)
      } catch (error) {
        console.error('Failed to load stats:', error)
      }
    }

    const loadPortfolioStats = async () => {
      try {
        portfolioStats.value = await GetPortfolioStats()
        console.log('💼 Portfolio stats loaded:', portfolioStats.value)
      } catch (error) {
        console.error('Failed to load portfolio stats:', error)
      }
    }

    const startBot = async (config) => {
      try {
        await StartBot(
          config.strategy,
          config.symbol,
          config.quantity,
          config.paperTrading,
          config.params
        )
        await loadBotStatus()
      } catch (error) {
        console.error('Failed to start bot:', error)
        alert('Failed to start bot: ' + error)
      }
    }

    const stopBot = async () => {
      try {
        await StopBot()
        // Force immediate UI update
        botStatus.value.running = false
        await loadBotStatus()
      } catch (error) {
        console.error('Failed to stop bot:', error)
        // Even if there's an error, try to sync state
        await loadBotStatus()
        alert('Failed to stop bot: ' + error)
      }
    }

    const refreshData = () => {
      loadBotStatus()
      loadTradeHistory()
      loadStats()
      loadPortfolioStats()
    }

    const checkSetupStatus = async () => {
      try {
        setupComplete.value = await IsSetupComplete()
      } catch (error) {
        console.error('Failed to check setup status:', error)
      }
    }

    const checkLockStatus = async () => {
      try {
        isLocked.value = await IsLocked()
        hasPin.value = await HasPIN()
      } catch (error) {
        console.error('Failed to check lock status:', error)
      }
    }

    const handleSetupComplete = async () => {
      setupComplete.value = true
      await checkLockStatus()
    }

    const handleUnlock = () => {
      isLocked.value = false
      refreshData()
    }

    const resetAPIKeys = async () => {
      if (!confirm('Are you sure you want to reset your API keys? This will clear your credentials and return you to the setup wizard.')) {
        return
      }

      try {
        await ResetSetup()
        // Reload the page to restart from setup wizard
        window.location.reload()
      } catch (error) {
        console.error('Failed to reset setup:', error)
        alert('Failed to reset: ' + error)
      }
    }

    const clearDemoTrades = async () => {
      if (!confirm('Clear all demo/paper trades? This will remove test trading data but keep live trades.')) {
        return
      }

      try {
        await ClearDemoTrades()

        // Force UI refresh by clearing arrays first
        trades.value = []
        stats.value = {}
        portfolioStats.value = {}

        // Then reload fresh data
        await loadTradeHistory()
        await loadStats()
        await loadPortfolioStats()

        alert('✅ Demo trades cleared successfully!')
      } catch (error) {
        console.error('Failed to clear demo trades:', error)
        alert('Failed to clear demo trades: ' + error)
      }
    }

    onMounted(async () => {
      await checkSetupStatus()

      if (setupComplete.value) {
        await checkLockStatus()

        if (!isLocked.value) {
          refreshData()
        }
      }

      refreshInterval = setInterval(refreshData, 5000)

      EventsOn('bot:started', (strategy) => {
        console.log('Bot started event received:', strategy)
        window.dispatchEvent(new CustomEvent('bot-started', { detail: strategy }))
        refreshData()
      })

      EventsOn('bot:stopped', () => {
        console.log('Bot stopped event received')
        window.dispatchEvent(new CustomEvent('bot-stopped'))
        // Force immediate UI state update
        botStatus.value.running = false
        // Then refresh all data
        refreshData()
      })

      EventsOn('bot:error', (error) => {
        console.error('Bot error:', error)
        alert('Bot error: ' + error)
      })

      // Forward all bot events to ActivityLog via event bus
      EventsOn('bot:connected', (data) => {
        window.dispatchEvent(new CustomEvent('activity-log', { detail: { type: 'bot:connected', data } }))
      })

      EventsOn('bot:candle', (data) => {
        window.dispatchEvent(new CustomEvent('activity-log', { detail: { type: 'bot:candle', data } }))

        // Update live price tracking
        if (data && data.data && data.data.price) {
          const newPrice = data.data.price
          const oldPrice = currentPrice.value

          currentPrice.value = newPrice
          isPriceUp.value = newPrice >= oldPrice

          // Trigger pulse animation on price change
          if (oldPrice > 0 && newPrice !== oldPrice) {
            isPricePulsing.value = true
            setTimeout(() => {
              isPricePulsing.value = false
            }, 2000) // Match CSS animation duration
          }

          // Store price with timestamp for 24h calculation
          const now = Date.now()
          prices24h.value.push({ price: newPrice, timestamp: now })

          // Keep only last 24 hours of prices
          const dayAgo = now - (24 * 60 * 60 * 1000)
          prices24h.value = prices24h.value.filter(p => p.timestamp > dayAgo)

          // Calculate 24h change if we have data from 24h ago
          if (prices24h.value.length > 0) {
            const price24hAgo = prices24h.value[0].price
            priceChange24h.value = newPrice - price24hAgo
            priceChangePercent24h.value = ((newPrice - price24hAgo) / price24hAgo) * 100
          }
        }
      })

      EventsOn('bot:indicator', (data) => {
        window.dispatchEvent(new CustomEvent('activity-log', { detail: { type: 'bot:indicator', data } }))
      })

      EventsOn('bot:trade', (data) => {
        window.dispatchEvent(new CustomEvent('activity-log', { detail: { type: 'bot:trade', data } }))
      })

      EventsOn('bot:status', (data) => {
        window.dispatchEvent(new CustomEvent('activity-log', { detail: { type: 'bot:status', data } }))
      })
    })

    onUnmounted(() => {
      if (refreshInterval) {
        clearInterval(refreshInterval)
      }
    })

    return {
      setupComplete,
      isLocked,
      hasPin,
      botStatus,
      trades,
      stats,
      portfolioStats,
      showAllTradesModal,
      currentPrice,
      priceChange24h,
      priceChangePercent24h,
      isPriceUp,
      isPricePulsing,
      startBot,
      stopBot,
      refreshData,
      handleSetupComplete,
      handleUnlock,
      resetAPIKeys,
      clearDemoTrades
    }
  }
}
</script>
