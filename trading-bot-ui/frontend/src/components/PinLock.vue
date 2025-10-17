<template>
  <div class="fixed inset-0 bg-gray-900 flex items-center justify-center z-50">
    <div class="bg-gray-800 rounded-lg p-8 border border-gray-700 max-w-md w-full">
      <!-- Lock Icon -->
      <div class="text-center mb-6">
        <div class="text-6xl mb-4">🔒</div>
        <h2 class="text-2xl font-bold text-blue-400">
          {{ hasPin ? 'Enter PIN' : 'Set PIN' }}
        </h2>
        <p class="text-gray-400 text-sm mt-2">
          {{ hasPin ? 'Unlock to access trading bot' : 'Create a PIN to secure your trading bot' }}
        </p>
      </div>

      <!-- PIN Input -->
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <input
            v-model="pin"
            type="password"
            :placeholder="hasPin ? 'Enter PIN' : 'Create PIN (min 4 digits)'"
            maxlength="10"
            class="w-full bg-gray-700 border border-gray-600 rounded px-4 py-3 text-gray-100 text-center text-lg tracking-widest focus:outline-none focus:border-blue-500"
            autofocus
            @input="error = ''"
          />
        </div>

        <!-- Confirm PIN (only for setup) -->
        <div v-if="!hasPin">
          <input
            v-model="confirmPin"
            type="password"
            placeholder="Confirm PIN"
            maxlength="10"
            class="w-full bg-gray-700 border border-gray-600 rounded px-4 py-3 text-gray-100 text-center text-lg tracking-widest focus:outline-none focus:border-blue-500"
            @input="error = ''"
          />
        </div>

        <!-- Error Message -->
        <div v-if="error" class="text-red-400 text-sm text-center">
          {{ error }}
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          class="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 px-4 rounded transition-colors"
        >
          {{ hasPin ? 'Unlock' : 'Set PIN' }}
        </button>

        <!-- Skip PIN (only on first setup) -->
        <button
          v-if="!hasPin"
          type="button"
          @click="handleSkip"
          class="w-full bg-gray-700 hover:bg-gray-600 text-gray-300 font-semibold py-2 px-4 rounded transition-colors text-sm"
        >
          Skip (Not Recommended)
        </button>
      </form>

      <!-- Info Text -->
      <div class="mt-6 text-xs text-gray-500 text-center">
        <p>Your PIN is stored securely using SHA-256 hashing</p>
        <p class="mt-1">Never shared or transmitted</p>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { UnlockApp, SetPIN } from '../../wailsjs/go/main/App'

export default {
  name: 'PinLock',
  props: {
    hasPin: {
      type: Boolean,
      required: true
    }
  },
  emits: ['unlocked'],
  setup(props, { emit }) {
    const pin = ref('')
    const confirmPin = ref('')
    const error = ref('')

    const handleSubmit = async () => {
      error.value = ''

      // Validation
      if (pin.value.length < 4) {
        error.value = 'PIN must be at least 4 characters'
        return
      }

      if (props.hasPin) {
        // Unlock with existing PIN
        try {
          await UnlockApp(pin.value)
          emit('unlocked')
        } catch (err) {
          error.value = 'Incorrect PIN'
          pin.value = ''
        }
      } else {
        // Set new PIN
        if (pin.value !== confirmPin.value) {
          error.value = 'PINs do not match'
          return
        }

        try {
          await SetPIN(pin.value)
          emit('unlocked')
        } catch (err) {
          error.value = err.toString()
        }
      }
    }

    const handleSkip = () => {
      if (confirm('Are you sure you want to skip PIN protection? Your trading bot will be unprotected.')) {
        emit('unlocked')
      }
    }

    return {
      pin,
      confirmPin,
      error,
      handleSubmit,
      handleSkip
    }
  }
}
</script>
