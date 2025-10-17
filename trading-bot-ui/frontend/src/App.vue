<template>
  <div id="app" class="min-h-screen bg-gray-900 text-gray-100">
    <!-- Setup Wizard (first-time API key setup) -->
    <SetupWizard v-if="!setupComplete" @setup-complete="handleSetupComplete" />

    <!-- PIN Lock Screen -->
    <PinLock v-else-if="isLocked" :hasPin="hasPin" @unlocked="handleUnlock" />

    <!-- Main App (only shown when setup complete and unlocked) -->
    <div v-else>
    <!-- Header -->
    <header class="bg-gray-800 border-b border-gray-700 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <h1 class="text-2xl font-bold text-blue-400">Trading Bot</h1>
          <span v-if="botStatus.running" class="flex items-center text-green-400">
            <span class="w-2 h-2 bg-green-400 rounded-full mr-2 animate-pulse"></span>
            Running
          </span>
          <span v-else class="flex items-center text-gray-400">
            <span class="w-2 h-2 bg-gray-400 rounded-full mr-2"></span>
            Stopped
          </span>
        </div>
        <div class="flex items-center space-x-4">
          <span v-if="botStatus.symbol" class="text-sm text-gray-400">
            {{ botStatus.symbol }}
          </span>
          <span v-if="botStatus.trading_mode"
                :class="botStatus.trading_mode === 'live' ? 'text-red-400' : 'text-blue-400'"
                class="text-sm font-semibold px-3 py-1 rounded-full border"
                :style="{ borderColor: botStatus.trading_mode === 'live' ? '#f87171' : '#60a5fa' }">
            {{ botStatus.trading_mode === 'live' ? '🔴 LIVE' : '📝 PAPER' }}
          </span>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <div class="container mx-auto px-6 py-8">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Left Panel: Bot Controls -->
        <div class="lg:col-span-1 space-y-6">
          <BotControls
            :botStatus="botStatus"
            @start-bot="startBot"
            @stop-bot="stopBot"
          />

          <PerformanceStats :stats="stats" />

          <CurrentPosition :position="botStatus.position" />
        </div>

        <!-- Right Panel: Trade History & Charts -->
        <div class="lg:col-span-2 space-y-6">
          <ActivityLog />
          <TradeHistory :trades="trades" />
        </div>
      </div>
    </div>
    </div><!-- End unlocked content -->
  </div>
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

export default {
  name: 'App',
  components: {
    SetupWizard,
    PinLock,
    BotControls,
    PerformanceStats,
    CurrentPosition,
    ActivityLog,
    TradeHistory
  },
  setup() {
    const setupComplete = ref(true) // Will be set to false if setup is incomplete
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
      // After setup, check if PIN is needed
      await checkLockStatus()
    }

    const handleUnlock = () => {
      isLocked.value = false
      refreshData()
    }

    onMounted(async () => {
      // Check setup status first
      await checkSetupStatus()

      // Only check lock status and load data if setup is complete
      if (setupComplete.value) {
        await checkLockStatus()

        // Only load data if unlocked
        if (!isLocked.value) {
          refreshData()
        }
      }

      // Set up auto-refresh every 5 seconds
      refreshInterval = setInterval(refreshData, 5000)

      // Listen for bot events
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

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

#app {
  min-height: 100vh;
}
</style>
