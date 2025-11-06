<template>
  <v-dialog :model-value="true" persistent fullscreen>
    <v-card color="surface">
      <v-container class="fill-height">
        <v-row align="center" justify="center">
          <v-col cols="12" md="8" lg="6" xl="5">
            <v-card elevation="8">
              <!-- Header -->
              <v-card-title class="text-center py-8">
                <div class="d-flex flex-column align-center">
                  <div class="text-h2 mb-4">🚀</div>
                  <div class="text-h4 text-primary font-weight-bold mb-2">
                    Welcome to Tradecraft
                  </div>
                  <div class="text-subtitle-1 text-medium-emphasis">
                    {{ step === 1 ? "Let's set up your Binance API connection" : "Create a PIN for security" }}
                  </div>
                </div>
              </v-card-title>

              <v-divider></v-divider>

              <!-- Step Indicator -->
              <v-card-text class="pt-4 pb-0">
                <v-stepper
                  v-model="step"
                  :items="['API Keys', 'Security PIN']"
                  alt-labels
                  flat
                  hide-actions
                ></v-stepper>
              </v-card-text>

              <v-divider></v-divider>

              <!-- Step 1: API Keys -->
              <v-card-text v-if="step === 1" class="pa-6">
                <!-- Instructions Panel -->
                <v-card v-if="showInstructions" variant="tonal" color="primary" class="mb-4">
                  <v-card-title class="text-h6">
                    <v-icon icon="mdi-information" class="mr-2"></v-icon>
                    How to Get Your API Keys
                  </v-card-title>
                  <v-card-text>
                    <v-list density="comfortable" bg-color="transparent">
                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">1</v-chip>
                        </template>
                        <v-list-item-title>
                          Log in to
                          <a href="https://www.binance.us" target="_blank" class="text-primary text-decoration-underline">
                            Binance US
                          </a>
                        </v-list-item-title>
                      </v-list-item>

                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">2</v-chip>
                        </template>
                        <v-list-item-title>Go to Profile → API Management</v-list-item-title>
                      </v-list-item>

                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">3</v-chip>
                        </template>
                        <v-list-item-title>Click "Create API" and complete verification</v-list-item-title>
                      </v-list-item>

                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">4</v-chip>
                        </template>
                        <v-list-item-title>Copy your API Key and Secret Key</v-list-item-title>
                      </v-list-item>
                    </v-list>

                    <v-divider class="my-3"></v-divider>

                    <v-alert type="warning" variant="tonal" density="compact">
                      <div class="text-subtitle-2 font-weight-bold mb-2">Security Settings:</div>
                      <ul class="text-caption">
                        <li>✓ Enable "Enable Trading"</li>
                        <li>✗ DISABLE "Enable Withdrawals"</li>
                        <li>Consider IP whitelisting for extra security</li>
                      </ul>
                    </v-alert>
                  </v-card-text>

                  <v-card-actions>
                    <v-spacer></v-spacer>
                    <v-btn variant="text" @click="showInstructions = false">
                      Got it!
                    </v-btn>
                  </v-card-actions>
                </v-card>

                <!-- API Key Form -->
                <form @submit.prevent="handleAPISubmit">
                  <!-- API Key Input -->
                  <v-text-field
                    v-model="apiKey"
                    label="Binance API Key"
                    placeholder="Enter your Binance API Key"
                    variant="outlined"
                    density="comfortable"
                    prepend-inner-icon="mdi-key"
                    class="mb-4 mono-font"
                    @input="error = ''"
                  ></v-text-field>

                  <!-- API Secret Input -->
                  <v-text-field
                    v-model="apiSecret"
                    label="Binance API Secret"
                    placeholder="Enter your Binance API Secret"
                    variant="outlined"
                    density="comfortable"
                    prepend-inner-icon="mdi-lock"
                    :type="showSecret ? 'text' : 'password'"
                    :append-inner-icon="showSecret ? 'mdi-eye-off' : 'mdi-eye'"
                    @click:append-inner="showSecret = !showSecret"
                    class="mb-4 mono-font"
                    @input="error = ''"
                  ></v-text-field>

                  <!-- Error Message -->
                  <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable>
                    {{ error }}
                  </v-alert>

                  <!-- Buttons -->
                  <v-btn
                    type="submit"
                    block
                    size="x-large"
                    color="primary"
                    :disabled="!apiKey || !apiSecret || loading"
                    :loading="loading"
                  >
                    Continue to Security Setup
                  </v-btn>

                  <v-btn
                    v-if="!showInstructions"
                    block
                    variant="text"
                    class="mt-2"
                    @click="showInstructions = true"
                  >
                    <v-icon icon="mdi-information-outline" start></v-icon>
                    Show Instructions Again
                  </v-btn>
                </form>
              </v-card-text>

              <!-- Step 2: PIN Creation -->
              <v-card-text v-else class="pa-6">
                <v-alert type="info" variant="tonal" class="mb-4">
                  <div class="text-subtitle-2 font-weight-bold mb-2">Create a Security PIN</div>
                  <div class="text-caption">
                    Your PIN will protect access to Tradecraft and your API credentials.
                    Choose a 4-6 digit PIN that you'll remember.
                  </div>
                </v-alert>

                <form @submit.prevent="handlePINSubmit">
                  <!-- PIN Input -->
                  <v-text-field
                    v-model="pin"
                    label="Create PIN (4-6 digits)"
                    placeholder="Enter your PIN"
                    variant="outlined"
                    density="comfortable"
                    prepend-inner-icon="mdi-lock"
                    :type="showPin ? 'text' : 'password'"
                    :append-inner-icon="showPin ? 'mdi-eye-off' : 'mdi-eye'"
                    @click:append-inner="showPin = !showPin"
                    class="mb-4"
                    inputmode="numeric"
                    @input="error = ''"
                  ></v-text-field>

                  <!-- Confirm PIN Input -->
                  <v-text-field
                    v-model="confirmPin"
                    label="Confirm PIN"
                    placeholder="Re-enter your PIN"
                    variant="outlined"
                    density="comfortable"
                    prepend-inner-icon="mdi-lock-check"
                    :type="showConfirmPin ? 'text' : 'password'"
                    :append-inner-icon="showConfirmPin ? 'mdi-eye-off' : 'mdi-eye'"
                    @click:append-inner="showConfirmPin = !showConfirmPin"
                    class="mb-4"
                    inputmode="numeric"
                    @input="error = ''"
                  ></v-text-field>

                  <!-- Error Message -->
                  <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable>
                    {{ error }}
                  </v-alert>

                  <!-- Success Message -->
                  <v-alert v-if="success" type="success" variant="tonal" class="mb-4">
                    <v-icon icon="mdi-check-circle" class="mr-2"></v-icon>
                    Setup complete! Launching Tradecraft...
                  </v-alert>

                  <!-- Buttons -->
                  <v-btn
                    type="submit"
                    block
                    size="x-large"
                    color="success"
                    :disabled="!pin || !confirmPin || loading"
                    :loading="loading"
                    class="mb-2"
                  >
                    Complete Setup
                  </v-btn>

                  <v-btn
                    block
                    variant="outlined"
                    @click="step = 1"
                    :disabled="loading"
                  >
                    <v-icon icon="mdi-arrow-left" start></v-icon>
                    Back to API Keys
                  </v-btn>
                </form>
              </v-card-text>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </v-dialog>
</template>

<script>
import { ref } from 'vue'
import { SaveAPIKeys, SetPIN } from '../../wailsjs/go/main/App'

export default {
  name: 'SetupWizard',
  emits: ['setup-complete'],
  setup(props, { emit }) {
    const step = ref(1)
    const showInstructions = ref(true)

    // Step 1: API Keys
    const apiKey = ref('')
    const apiSecret = ref('')
    const showSecret = ref(false)

    // Step 2: PIN
    const pin = ref('')
    const confirmPin = ref('')
    const showPin = ref(false)
    const showConfirmPin = ref(false)

    // Common
    const error = ref('')
    const success = ref(false)
    const loading = ref(false)

    const handleAPISubmit = async () => {
      error.value = ''
      loading.value = true

      try {
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

        await SaveAPIKeys(apiKey.value.trim(), apiSecret.value.trim())

        // Move to PIN step
        step.value = 2

      } catch (err) {
        error.value = err.toString().replace('Error: ', '')
      } finally {
        loading.value = false
      }
    }

    const handlePINSubmit = async () => {
      error.value = ''
      loading.value = true

      try {
        // Validate PIN
        if (!pin.value) {
          error.value = 'PIN is required'
          loading.value = false
          return
        }

        if (pin.value.length < 4 || pin.value.length > 6) {
          error.value = 'PIN must be 4-6 digits'
          loading.value = false
          return
        }

        if (!/^\d+$/.test(pin.value)) {
          error.value = 'PIN must contain only numbers'
          loading.value = false
          return
        }

        if (pin.value !== confirmPin.value) {
          error.value = 'PINs do not match'
          loading.value = false
          return
        }

        // Save PIN
        await SetPIN(pin.value)
        success.value = true

        // Complete setup
        setTimeout(() => {
          emit('setup-complete')
        }, 1500)

      } catch (err) {
        error.value = err.toString().replace('Error: ', '')
      } finally {
        loading.value = false
      }
    }

    return {
      step,
      showInstructions,
      apiKey,
      apiSecret,
      showSecret,
      pin,
      confirmPin,
      showPin,
      showConfirmPin,
      error,
      success,
      loading,
      handleAPISubmit,
      handlePINSubmit
    }
  }
}
</script>

<style scoped>
.mono-font :deep(input) {
  font-family: 'Courier New', Courier, monospace;
}
</style>
