<template>
  <div class="fixed inset-0 bg-gray-900 flex items-center justify-center z-50">
    <div class="bg-gray-800 rounded-lg p-8 border border-gray-700 max-w-2xl w-full">
      <!-- Header -->
      <div class="text-center mb-8">
        <div class="text-6xl mb-4">🚀</div>
        <h1 class="text-3xl font-bold text-blue-400 mb-2">Welcome to Trading Bot</h1>
        <p class="text-gray-400">Let's set up your Binance API connection</p>
      </div>

      <!-- Instructions Panel -->
      <div v-if="showInstructions" class="mb-6 bg-gray-700 rounded-lg p-6">
        <h3 class="text-lg font-semibold text-blue-400 mb-3">How to Get Your API Keys:</h3>
        <ol class="space-y-2 text-sm text-gray-300">
          <li class="flex items-start">
            <span class="text-blue-400 mr-2">1.</span>
            <span>Log in to <a href="https://www.binance.com" target="_blank" class="text-blue-400 underline">Binance.com</a></span>
          </li>
          <li class="flex items-start">
            <span class="text-blue-400 mr-2">2.</span>
            <span>Go to <strong>Profile → API Management</strong></span>
          </li>
          <li class="flex items-start">
            <span class="text-blue-400 mr-2">3.</span>
            <span>Click <strong>"Create API"</strong> and name it "Trading Bot"</span>
          </li>
          <li class="flex items-start">
            <span class="text-blue-400 mr-2">4.</span>
            <span>Complete 2FA verification</span>
          </li>
          <li class="flex items-start">
            <span class="text-blue-400 mr-2">5.</span>
            <span>Copy your <strong>API Key</strong> and <strong>Secret Key</strong></span>
          </li>
        </ol>

        <div class="mt-4 p-4 bg-red-900 bg-opacity-30 border border-red-700 rounded">
          <p class="text-red-400 text-sm font-semibold mb-2">⚠️ IMPORTANT Security Settings:</p>
          <ul class="text-sm text-gray-300 space-y-1">
            <li>✓ Enable "Enable Trading"</li>
            <li>✗ <strong>DISABLE "Enable Withdrawals"</strong> (for security!)</li>
            <li>✓ Consider restricting to your IP address</li>
          </ul>
        </div>

        <div class="mt-4 p-4 bg-blue-900 bg-opacity-30 border border-blue-700 rounded">
          <p class="text-blue-400 text-sm font-semibold mb-1">💡 For Testing:</p>
          <p class="text-sm text-gray-300">
            Use <a href="https://testnet.binance.vision" target="_blank" class="text-blue-400 underline">Binance Testnet</a>
            to test with fake money before using real funds.
          </p>
        </div>

        <button @click="showInstructions = false"
                class="mt-4 w-full bg-gray-600 hover:bg-gray-500 text-white py-2 rounded transition-colors">
          Got it, continue to setup
        </button>
      </div>

      <!-- API Key Form -->
      <form v-else @submit.prevent="handleSubmit" class="space-y-6">
        <!-- API Key Input -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">
            Binance API Key
          </label>
          <input
            v-model="apiKey"
            type="text"
            placeholder="Enter your Binance API Key"
            class="w-full bg-gray-700 border border-gray-600 rounded px-4 py-3 text-gray-100 font-mono text-sm focus:outline-none focus:border-blue-500"
            @input="error = ''"
          />
        </div>

        <!-- API Secret Input -->
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-2">
            Binance API Secret
          </label>
          <input
            v-model="apiSecret"
            :type="showSecret ? 'text' : 'password'"
            placeholder="Enter your Binance API Secret"
            class="w-full bg-gray-700 border border-gray-600 rounded px-4 py-3 text-gray-100 font-mono text-sm focus:outline-none focus:border-blue-500"
            @input="error = ''"
          />
          <label class="flex items-center mt-2 text-sm text-gray-400 cursor-pointer">
            <input v-model="showSecret" type="checkbox" class="mr-2">
            Show secret
          </label>
        </div>

        <!-- Error Message -->
        <div v-if="error" class="bg-red-900 bg-opacity-30 border border-red-700 rounded p-3">
          <p class="text-red-400 text-sm">{{ error }}</p>
        </div>

        <!-- Success Message -->
        <div v-if="success" class="bg-green-900 bg-opacity-30 border border-green-700 rounded p-3">
          <p class="text-green-400 text-sm">✓ API keys saved successfully!</p>
        </div>

        <!-- Buttons -->
        <div class="space-y-3">
          <button
            type="submit"
            :disabled="!apiKey || !apiSecret || loading"
            class="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed text-white font-semibold py-3 px-4 rounded transition-colors"
          >
            {{ loading ? 'Saving...' : 'Save API Keys' }}
          </button>

          <button
            type="button"
            @click="showInstructions = true"
            class="w-full bg-gray-700 hover:bg-gray-600 text-gray-300 py-2 px-4 rounded transition-colors text-sm"
          >
            ← Back to instructions
          </button>
        </div>

        <!-- Security Note -->
        <div class="text-xs text-gray-500 text-center space-y-1">
          <p>🔒 Your API keys are stored locally on your computer</p>
          <p>Never shared or transmitted anywhere</p>
          <p>Encrypted file location: <code class="text-gray-400">~/.config/trading-bot/.env</code></p>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { SaveAPIKeys } from '../../wailsjs/go/main/App'

export default {
  name: 'SetupWizard',
  emits: ['setup-complete'],
  setup(props, { emit }) {
    const showInstructions = ref(true)
    const apiKey = ref('')
    const apiSecret = ref('')
    const showSecret = ref(false)
    const error = ref('')
    const success = ref(false)
    const loading = ref(false)

    const handleSubmit = async () => {
      error.value = ''
      success.value = false
      loading.value = true

      try {
        // Validate
        if (!apiKey.value.trim()) {
          error.value = 'API Key is required'
          loading.value = false
          return
        }

        if (!apiSecret.value.trim()) {
          error.value = 'API Secret is required'
          loading.value = false
          return
        }

        // Save to backend
        await SaveAPIKeys(apiKey.value.trim(), apiSecret.value.trim())

        success.value = true

        // Emit completion after short delay
        setTimeout(() => {
          emit('setup-complete')
        }, 1000)

      } catch (err) {
        error.value = err.toString().replace('Error: ', '')
      } finally {
        loading.value = false
      }
    }

    return {
      showInstructions,
      apiKey,
      apiSecret,
      showSecret,
      error,
      success,
      loading,
      handleSubmit
    }
  }
}
</script>
