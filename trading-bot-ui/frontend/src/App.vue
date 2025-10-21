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
          <v-chip
            v-if="botStatus.symbol"
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
          >
            <v-icon
              :icon="botStatus.running ? 'mdi-circle' : 'mdi-circle-outline'"
              start
              size="small"
            ></v-icon>
            {{ botStatus.running ? 'Running' : 'Stopped' }}
          </v-chip>
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
              />

              <PerformanceStats :stats="stats" class="mt-4" />

              <CurrentPosition :position="botStatus.position" class="mt-4" />
            </v-col>

            <!-- Right Column: Activity & Trades -->
            <v-col cols="12" lg="8">
              <ActivityLog class="mb-4" />

              <TradeHistory :trades="trades" />
            </v-col>
          </v-row>
        </v-container>
      </v-main>
    </template>
  </v-app>
</template>

<script>
import { ref, onMounted, onUnmounted } from 'vue'
import { GetBotStatus, StartBot, StopBot, GetTradeHistory, GetTradeSummary, IsLocked, HasPIN, IsSetupComplete } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import SetupWizard from './components/SetupWizard.vue'
import PinLock from './components/PinLock.vue'
import BotControls from './components/BotControls.vue'
import PerformanceStats from './components/PerformanceStats.vue'
import CurrentPosition from './components/CurrentPosition.vue'
import ActivityLog from './components/ActivityLog.vue'
import TradeHistory from './components/TradeHistory.vue'
import WalletBalance from './components/WalletBalance.vue'

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
    WalletBalance
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
      } catch (error) {
        console.error('Failed to load stats:', error)
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
        await loadBotStatus()
      } catch (error) {
        console.error('Failed to stop bot:', error)
        alert('Failed to stop bot: ' + error)
      }
    }

    const refreshData = () => {
      loadBotStatus()
      loadTradeHistory()
      loadStats()
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

    onMounted(async () => {
      await checkSetupStatus()

      if (setupComplete.value) {
        await checkLockStatus()

        if (!isLocked.value) {
          refreshData()
        }
      }

      refreshInterval = setInterval(refreshData, 5000)

      EventsOn('bot:started', () => {
        console.log('Bot started event received')
        refreshData()
      })

      EventsOn('bot:stopped', () => {
        console.log('Bot stopped event received')
        refreshData()
      })

      EventsOn('bot:error', (error) => {
        console.error('Bot error:', error)
        alert('Bot error: ' + error)
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
      startBot,
      stopBot,
      handleSetupComplete,
      handleUnlock
    }
  }
}
</script>
