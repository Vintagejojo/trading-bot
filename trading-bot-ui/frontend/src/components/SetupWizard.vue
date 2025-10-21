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
                    Welcome to Jojo's Nerd Casino
                  </div>
                  <div class="text-subtitle-1 text-medium-emphasis">
                    Let's set up your Binance API connection
                  </div>
                </div>
              </v-card-title>

              <v-divider></v-divider>

              <!-- Instructions Panel -->
              <v-card-text v-if="showInstructions" class="pa-6">
                <v-card variant="tonal" color="primary" class="mb-4">
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
                          <a
                            href="https://www.binance.com"
                            target="_blank"
                            class="text-primary text-decoration-underline"
                          >
                            Binance.com
                          </a>
                        </v-list-item-title>
                      </v-list-item>

                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">2</v-chip>
                        </template>
                        <v-list-item-title>
                          Go to <span class="font-weight-bold">Profile → API Management</span>
                        </v-list-item-title>
                      </v-list-item>

                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">3</v-chip>
                        </template>
                        <v-list-item-title>
                          Click <span class="font-weight-bold">"Create API"</span> and name it "Trading Bot"
                        </v-list-item-title>
                      </v-list-item>

                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">4</v-chip>
                        </template>
                        <v-list-item-title>Complete 2FA verification</v-list-item-title>
                      </v-list-item>

                      <v-list-item>
                        <template v-slot:prepend>
                          <v-chip color="primary" size="small" class="font-weight-bold">5</v-chip>
                        </template>
                        <v-list-item-title>
                          Copy your <span class="font-weight-bold">API Key</span> and
                          <span class="font-weight-bold">Secret Key</span>
                        </v-list-item-title>
                      </v-list-item>
                    </v-list>
                  </v-card-text>
                </v-card>

                <!-- Security Warning -->
                <v-alert type="error" variant="tonal" prominent class="mb-4">
                  <v-alert-title class="text-h6 mb-3">
                    <v-icon icon="mdi-shield-alert" class="mr-2"></v-icon>
                    IMPORTANT Security Settings
                  </v-alert-title>
                  <v-list density="compact" bg-color="transparent">
                    <v-list-item>
                      <template v-slot:prepend>
                        <v-icon icon="mdi-check-circle" color="success" size="small"></v-icon>
                      </template>
                      <v-list-item-title>Enable "Enable Trading"</v-list-item-title>
                    </v-list-item>
                    <v-list-item>
                      <template v-slot:prepend>
                        <v-icon icon="mdi-close-circle" color="error" size="small"></v-icon>
                      </template>
                      <v-list-item-title class="font-weight-bold">
                        DISABLE "Enable Withdrawals" (for security!)
                      </v-list-item-title>
                    </v-list-item>
                    <v-list-item>
                      <template v-slot:prepend>
                        <v-icon icon="mdi-check-circle" color="success" size="small"></v-icon>
                      </template>
                      <v-list-item-title>Consider restricting to your IP address</v-list-item-title>
                    </v-list-item>
                  </v-list>
                </v-alert>

                <!-- Testing Info -->
                <v-alert type="info" variant="tonal" class="mb-4">
                  <v-alert-title class="text-h6 mb-2">
                    <v-icon icon="mdi-lightbulb-on" class="mr-2"></v-icon>
                    For Testing
                  </v-alert-title>
                  <div>
                    Use
                    <a
                      href="https://testnet.binance.vision"
                      target="_blank"
                      class="text-info font-weight-bold text-decoration-underline"
                    >
                      Binance Testnet
                    </a>
                    to test with fake money before using real funds.
                  </div>
                </v-alert>

                <v-btn
                  block
                  size="x-large"
                  color="primary"
                  @click="showInstructions = false"
                >
                  Got it, continue to setup
                </v-btn>
              </v-card-text>

              <!-- API Key Form -->
              <v-card-text v-else class="pa-6">
                <v-form @submit.prevent="handleSubmit">
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

                  <!-- Success Message -->
                  <v-alert v-if="success" type="success" variant="tonal" class="mb-4">
                    <v-icon icon="mdi-check-circle" class="mr-2"></v-icon>
                    API keys saved successfully!
                  </v-alert>

                  <!-- Buttons -->
                  <v-btn
                    type="submit"
                    block
                    size="x-large"
                    color="primary"
                    :disabled="!apiKey || !apiSecret || loading"
                    :loading="loading"
                    class="mb-3"
                  >
                    Save API Keys
                  </v-btn>

                  <v-btn
                    block
                    size="large"
                    variant="outlined"
                    @click="showInstructions = true"
                  >
                    <v-icon icon="mdi-arrow-left" class="mr-2"></v-icon>
                    Back to instructions
                  </v-btn>

                  <!-- Security Note -->
                  <v-card variant="tonal" color="grey-darken-3" class="mt-6">
                    <v-card-text class="text-center">
                      <div class="text-body-2 mb-2">
                        <v-icon icon="mdi-lock-outline" size="small" class="mr-1"></v-icon>
                        Your API keys are stored locally on your computer
                      </div>
                      <div class="text-caption text-medium-emphasis">
                        Never shared or transmitted anywhere
                      </div>
                      <div class="text-caption text-medium-emphasis mono-font mt-1">
                        ~/.config/trading-bot/.env
                      </div>
                    </v-card-text>
                  </v-card>
                </v-form>
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
        success.value = true

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

<style scoped>
.mono-font :deep(input) {
  font-family: 'Courier New', Courier, monospace;
}
</style>