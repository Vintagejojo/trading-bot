<template>
  <v-app>
    <v-main class="d-flex align-center justify-center" style="min-height: 100vh;">
      <v-container>
        <v-row justify="center">
          <v-col cols="12" sm="8" md="6" lg="4">
            <v-card elevation="8">
              <v-card-text class="text-center pa-8">
                <!-- Lock Icon -->
                <v-icon icon="mdi-lock" size="80" color="primary" class="mb-4"></v-icon>

                <h2 class="text-h4 font-weight-bold mb-2">
                  {{ hasPin ? 'Enter PIN' : 'Set PIN' }}
                </h2>
                <p class="text-body-2 text-grey mb-6">
                  {{ hasPin ? 'Unlock to access trading bot' : 'Create a PIN to secure your trading bot' }}
                </p>

                <!-- PIN Input -->
                <v-form @submit.prevent="handleSubmit">
                  <v-text-field
                    v-model="pin"
                    type="password"
                    :label="hasPin ? 'Enter PIN' : 'Create PIN (min 4 digits)'"
                    variant="outlined"
                    density="comfortable"
                    maxlength="10"
                    autofocus
                    class="mb-4 text-center"
                    @input="error = ''"
                  ></v-text-field>

                  <!-- Confirm PIN (only for setup) -->
                  <v-text-field
                    v-if="!hasPin"
                    v-model="confirmPin"
                    type="password"
                    label="Confirm PIN"
                    variant="outlined"
                    density="comfortable"
                    maxlength="10"
                    class="mb-4"
                    @input="error = ''"
                  ></v-text-field>

                  <!-- Error Message -->
                  <v-alert
                    v-if="error"
                    type="error"
                    variant="tonal"
                    class="mb-4"
                  >
                    {{ error }}
                  </v-alert>

                  <!-- Submit Button -->
                  <v-btn
                    type="submit"
                    block
                    size="large"
                    color="primary"
                    class="mb-3"
                  >
                    {{ hasPin ? 'Unlock' : 'Set PIN' }}
                  </v-btn>

                  <!-- Skip PIN (only on first setup) -->
                  <v-btn
                    v-if="!hasPin"
                    block
                    variant="outlined"
                    color="grey"
                    @click="handleSkip"
                  >
                    Skip (Not Recommended)
                  </v-btn>
                </v-form>

                <!-- Info Text -->
                <div class="mt-6">
                  <p class="text-caption text-grey">
                    Your PIN is stored securely using SHA-256 hashing
                  </p>
                  <p class="text-caption text-grey">
                    Never shared or transmitted
                  </p>
                </div>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
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
