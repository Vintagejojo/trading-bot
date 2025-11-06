<template>
  <v-dialog v-model="dialog" max-width="500">
    <template v-slot:activator="{ props }">
      <v-list-item v-bind="props">
        <template v-slot:prepend>
          <v-icon icon="mdi-lock-reset"></v-icon>
        </template>
        <v-list-item-title>Change PIN</v-list-item-title>
        <v-list-item-subtitle>Update your security PIN</v-list-item-subtitle>
      </v-list-item>
    </template>

    <v-card>
      <v-card-title class="text-h5 bg-primary">
        <v-icon icon="mdi-lock-reset" class="mr-2"></v-icon>
        Change Security PIN
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <v-alert type="info" variant="tonal" density="compact" class="mb-4">
          For security, you must enter your current PIN before setting a new one.
        </v-alert>

        <form @submit.prevent="handleSubmit">
          <!-- Current PIN -->
          <v-text-field
            v-model="oldPin"
            label="Current PIN"
            placeholder="Enter your current PIN"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-lock"
            :type="showOldPin ? 'text' : 'password'"
            :append-inner-icon="showOldPin ? 'mdi-eye-off' : 'mdi-eye'"
            @click:append-inner="showOldPin = !showOldPin"
            class="mb-4"
            inputmode="numeric"
            @input="error = ''"
          ></v-text-field>

          <!-- New PIN -->
          <v-text-field
            v-model="newPin"
            label="New PIN (4-6 digits)"
            placeholder="Enter your new PIN"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-lock-plus"
            :type="showNewPin ? 'text' : 'password'"
            :append-inner-icon="showNewPin ? 'mdi-eye-off' : 'mdi-eye'"
            @click:append-inner="showNewPin = !showNewPin"
            class="mb-4"
            inputmode="numeric"
            @input="error = ''"
          ></v-text-field>

          <!-- Confirm New PIN -->
          <v-text-field
            v-model="confirmPin"
            label="Confirm New PIN"
            placeholder="Re-enter your new PIN"
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
            PIN changed successfully!
          </v-alert>
        </form>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn
          variant="text"
          @click="closeDialog"
          :disabled="loading"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          variant="flat"
          @click="handleSubmit"
          :disabled="!oldPin || !newPin || !confirmPin || loading"
          :loading="loading"
        >
          Change PIN
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import { ref } from 'vue'
import { ChangePIN } from '../../wailsjs/go/main/App'

export default {
  name: 'SettingsDialog',
  setup() {
    const dialog = ref(false)
    const oldPin = ref('')
    const newPin = ref('')
    const confirmPin = ref('')
    const showOldPin = ref(false)
    const showNewPin = ref(false)
    const showConfirmPin = ref(false)
    const error = ref('')
    const success = ref(false)
    const loading = ref(false)

    const handleSubmit = async () => {
      error.value = ''
      success.value = false
      loading.value = true

      try {
        // Validate inputs
        if (!oldPin.value) {
          error.value = 'Current PIN is required'
          loading.value = false
          return
        }

        if (!newPin.value) {
          error.value = 'New PIN is required'
          loading.value = false
          return
        }

        if (newPin.value.length < 4 || newPin.value.length > 6) {
          error.value = 'New PIN must be 4-6 digits'
          loading.value = false
          return
        }

        if (!/^\d+$/.test(newPin.value)) {
          error.value = 'PIN must contain only numbers'
          loading.value = false
          return
        }

        if (newPin.value !== confirmPin.value) {
          error.value = 'New PINs do not match'
          loading.value = false
          return
        }

        if (oldPin.value === newPin.value) {
          error.value = 'New PIN must be different from current PIN'
          loading.value = false
          return
        }

        // Call backend to change PIN
        await ChangePIN(oldPin.value, newPin.value)
        success.value = true

        // Reset form and close after delay
        setTimeout(() => {
          closeDialog()
        }, 1500)

      } catch (err) {
        error.value = err.toString().replace('Error: ', '')
      } finally {
        loading.value = false
      }
    }

    const closeDialog = () => {
      dialog.value = false
      // Reset form
      setTimeout(() => {
        oldPin.value = ''
        newPin.value = ''
        confirmPin.value = ''
        showOldPin.value = false
        showNewPin.value = false
        showConfirmPin.value = false
        error.value = ''
        success.value = false
      }, 300)
    }

    return {
      dialog,
      oldPin,
      newPin,
      confirmPin,
      showOldPin,
      showNewPin,
      showConfirmPin,
      error,
      success,
      loading,
      handleSubmit,
      closeDialog
    }
  }
}
</script>
