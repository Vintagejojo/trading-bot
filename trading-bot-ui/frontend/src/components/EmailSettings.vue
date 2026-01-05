<template>
  <v-dialog v-model="dialog" max-width="600">
    <template v-slot:activator="{ props }">
      <v-list-item v-bind="props">
        <template v-slot:prepend>
          <v-icon icon="mdi-email-outline"></v-icon>
        </template>
        <v-list-item-title>Email Notifications</v-list-item-title>
        <v-list-item-subtitle>Configure trade alerts and summaries</v-list-item-subtitle>
      </v-list-item>
    </template>

    <v-card>
      <v-card-title class="text-h5 bg-primary">
        <v-icon icon="mdi-email-outline" class="mr-2"></v-icon>
        Email Notification Settings
      </v-card-title>

      <v-divider></v-divider>

      <v-card-text class="pt-6">
        <v-alert type="info" variant="tonal" density="compact" class="mb-4">
          Choose when to receive trade notifications and monthly portfolio summaries. Granular control lets you get alerts only for important events.
        </v-alert>

        <form @submit.prevent="handleSave">
          <!-- Enable/Disable Toggle -->
          <v-switch
            v-model="settings.enabled"
            label="Enable Email Notifications"
            color="primary"
            class="mb-4"
            @update:model-value="error = ''"
          ></v-switch>

          <v-expand-transition>
            <div v-if="settings.enabled">
              <!-- Notification Preferences -->
              <v-card variant="outlined" class="mb-4">
                <v-card-subtitle class="text-caption font-weight-bold">
                  <v-icon icon="mdi-bell-outline" size="small" class="mr-1"></v-icon>
                  What to notify me about:
                </v-card-subtitle>
                <v-card-text>
                  <v-checkbox
                    v-model="settings.notifyOnDCABuy"
                    label="Regular DCA purchases"
                    color="primary"
                    density="compact"
                    hide-details
                    class="mb-2"
                  >
                    <template v-slot:label>
                      <div class="d-flex align-center">
                        <span>Regular DCA purchases</span>
                        <v-tooltip location="top">
                          <template v-slot:activator="{ props }">
                            <v-icon v-bind="props" icon="mdi-help-circle-outline" size="small" class="ml-2 text-grey"></v-icon>
                          </template>
                          Scheduled automatic buys based on your DCA interval
                        </v-tooltip>
                      </div>
                    </template>
                  </v-checkbox>

                  <v-checkbox
                    v-model="settings.notifyOnDipBuy"
                    label="Buy-the-dip purchases"
                    color="success"
                    density="compact"
                    hide-details
                    class="mb-2"
                  >
                    <template v-slot:label>
                      <div class="d-flex align-center">
                        <span>Buy-the-dip purchases</span>
                        <v-tooltip location="top">
                          <template v-slot:activator="{ props }">
                            <v-icon v-bind="props" icon="mdi-help-circle-outline" size="small" class="ml-2 text-grey"></v-icon>
                          </template>
                          Extra buys triggered when price drops >5% below average
                        </v-tooltip>
                      </div>
                    </template>
                  </v-checkbox>

                  <v-divider class="my-3"></v-divider>

                  <v-checkbox
                    v-model="settings.sendMonthlySummary"
                    label="Monthly portfolio summary"
                    color="info"
                    density="compact"
                    hide-details
                  >
                    <template v-slot:label>
                      <div class="d-flex align-center">
                        <span>Monthly portfolio summary</span>
                        <v-tooltip location="top">
                          <template v-slot:activator="{ props }">
                            <v-icon v-bind="props" icon="mdi-help-circle-outline" size="small" class="ml-2 text-grey"></v-icon>
                          </template>
                          Monthly report with performance stats and ROI analysis
                        </v-tooltip>
                      </div>
                    </template>
                  </v-checkbox>
                </v-card-text>
              </v-card>

              <!-- Notification Email Address -->
              <v-text-field
                v-model="settings.notificationEmail"
                label="Where to Send Notifications"
                placeholder="your.email@example.com"
                variant="outlined"
                density="comfortable"
                prepend-inner-icon="mdi-email"
                type="email"
                class="mb-4"
                hint="Email address where you'll receive trade alerts"
                persistent-hint
                @input="error = ''"
              ></v-text-field>

              <v-divider class="my-4"></v-divider>

              <!-- Email Provider Selection -->
              <v-select
                v-model="emailProvider"
                :items="emailProviders"
                label="Email Provider"
                variant="outlined"
                density="comfortable"
                class="mb-4"
                hint="Choose your email service"
                @update:model-value="updateProviderSettings"
              ></v-select>

              <!-- Sender Email -->
              <v-text-field
                v-model="settings.smtpFromEmail"
                :label="emailProvider === 'gmail' ? 'Your Gmail Address' : 'Your Email Address'"
                :placeholder="emailProvider === 'gmail' ? 'yourname@gmail.com' : 'your.email@example.com'"
                variant="outlined"
                density="comfortable"
                prepend-inner-icon="mdi-account"
                type="email"
                class="mb-4"
                hint="The email account that will send notifications"
                @input="error = ''"
              ></v-text-field>

              <!-- Password -->
              <v-text-field
                v-model="settings.smtpPassword"
                :label="emailProvider === 'gmail' ? 'App Password' : 'Password'"
                placeholder="Enter password"
                variant="outlined"
                density="comfortable"
                :type="showPassword ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock"
                :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                @click:append-inner="showPassword = !showPassword"
                class="mb-2"
                :hint="emailProvider === 'gmail' ? 'Use App Password, not regular password' : 'Your email password'"
                persistent-hint
                @input="error = ''"
              ></v-text-field>

              <!-- Gmail App Password Help -->
              <v-alert v-if="emailProvider === 'gmail'" type="info" variant="tonal" density="compact" class="mt-2 mb-4">
                <div class="text-caption">
                  <strong>Gmail Security:</strong> Google requires an App Password for third-party apps.
                  <br>
                  <a href="https://myaccount.google.com/apppasswords" target="_blank" class="text-primary">
                    Click here to create an App Password →
                  </a>
                  <br>
                  <small class="text-grey">(Takes 30 seconds - select "Mail" and your device name)</small>
                </div>
              </v-alert>

              <!-- Advanced Settings (Collapsible) -->
              <v-expansion-panels variant="accordion" class="mb-4">
                <v-expansion-panel>
                  <v-expansion-panel-title>
                    <v-icon icon="mdi-cog" class="mr-2"></v-icon>
                    Advanced Settings
                  </v-expansion-panel-title>
                  <v-expansion-panel-text>
                    <v-text-field
                      v-model="settings.smtpHost"
                      label="SMTP Host"
                      variant="outlined"
                      density="compact"
                      class="mb-2"
                    ></v-text-field>

                    <v-text-field
                      v-model.number="settings.smtpPort"
                      label="SMTP Port"
                      type="number"
                      variant="outlined"
                      density="compact"
                    ></v-text-field>
                  </v-expansion-panel-text>
                </v-expansion-panel>
              </v-expansion-panels>
            </div>
          </v-expand-transition>

          <!-- Error Message -->
          <v-alert v-if="error" type="error" variant="tonal" class="mb-4" closable>
            {{ error }}
          </v-alert>

          <!-- Success Message -->
          <v-alert v-if="success" type="success" variant="tonal" class="mb-4">
            <v-icon icon="mdi-check-circle" class="mr-2"></v-icon>
            {{ success }}
          </v-alert>

          <!-- Setup Tutorial -->
          <v-divider class="my-4"></v-divider>

          <v-expansion-panels variant="accordion">
            <v-expansion-panel>
              <v-expansion-panel-title>
                <v-icon icon="mdi-school" color="primary" class="mr-2"></v-icon>
                <strong>📚 Setup Tutorial (30 seconds)</strong>
              </v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-stepper :items="tutorialSteps" alt-labels hide-actions>
                  <template v-slot:item.1>
                    <v-card flat>
                      <v-card-text>
                        <div class="text-h6 mb-2">1️⃣ Enable Notifications</div>
                        <p>Toggle the switch above to enable email notifications.</p>
                      </v-card-text>
                    </v-card>
                  </template>

                  <template v-slot:item.2>
                    <v-card flat>
                      <v-card-text>
                        <div class="text-h6 mb-2">2️⃣ Choose Email Provider</div>
                        <p>Select <strong>Gmail</strong>, Outlook, or Yahoo from the dropdown.</p>
                        <p class="text-caption text-grey mt-2">Most users choose Gmail (it's free and reliable)</p>
                      </v-card-text>
                    </v-card>
                  </template>

                  <template v-slot:item.3>
                    <v-card flat>
                      <v-card-text>
                        <div class="text-h6 mb-2">3️⃣ Get App Password (Gmail Only)</div>
                        <v-list density="compact">
                          <v-list-item>
                            <template v-slot:prepend>
                              <v-icon icon="mdi-numeric-1-circle" color="primary"></v-icon>
                            </template>
                            <v-list-item-title>
                              <a href="https://myaccount.google.com/apppasswords" target="_blank" class="text-primary">
                                Click here to open Google App Passwords →
                              </a>
                            </v-list-item-title>
                            <v-list-item-subtitle>Opens in new tab</v-list-item-subtitle>
                          </v-list-item>

                          <v-list-item>
                            <template v-slot:prepend>
                              <v-icon icon="mdi-numeric-2-circle" color="primary"></v-icon>
                            </template>
                            <v-list-item-title>Sign in to your Google account</v-list-item-title>
                            <v-list-item-subtitle>Use the account where you want to receive alerts</v-list-item-subtitle>
                          </v-list-item>

                          <v-list-item>
                            <template v-slot:prepend>
                              <v-icon icon="mdi-numeric-3-circle" color="primary"></v-icon>
                            </template>
                            <v-list-item-title>Type "Tradecraft" as the app name</v-list-item-title>
                            <v-list-item-subtitle>Or any name you'll remember</v-list-item-subtitle>
                          </v-list-item>

                          <v-list-item>
                            <template v-slot:prepend>
                              <v-icon icon="mdi-numeric-4-circle" color="primary"></v-icon>
                            </template>
                            <v-list-item-title>Click "Create" button</v-list-item-title>
                          </v-list-item>

                          <v-list-item>
                            <template v-slot:prepend>
                              <v-icon icon="mdi-numeric-5-circle" color="primary"></v-icon>
                            </template>
                            <v-list-item-title>Copy the 16-character password</v-list-item-title>
                            <v-list-item-subtitle class="text-warning">Example: <code>abcd efgh ijkl mnop</code></v-list-item-subtitle>
                          </v-list-item>
                        </v-list>

                        <v-alert type="info" variant="tonal" density="compact" class="mt-3">
                          <div class="text-caption">
                            <strong>Why App Password?</strong> Google requires this for security. Your regular password won't work with third-party apps.
                          </div>
                        </v-alert>
                      </v-card-text>
                    </v-card>
                  </template>

                  <template v-slot:item.4>
                    <v-card flat>
                      <v-card-text>
                        <div class="text-h6 mb-2">4️⃣ Fill in the Form</div>
                        <v-list density="compact">
                          <v-list-item>
                            <v-list-item-title><strong>Where to Send Notifications:</strong></v-list-item-title>
                            <v-list-item-subtitle>Your email (can be the same as sender)</v-list-item-subtitle>
                          </v-list-item>
                          <v-list-item>
                            <v-list-item-title><strong>Your Gmail Address:</strong></v-list-item-title>
                            <v-list-item-subtitle>The Gmail account you just used</v-list-item-subtitle>
                          </v-list-item>
                          <v-list-item>
                            <v-list-item-title><strong>App Password:</strong></v-list-item-title>
                            <v-list-item-subtitle>Paste the 16-character code (remove spaces)</v-list-item-subtitle>
                          </v-list-item>
                        </v-list>
                      </v-card-text>
                    </v-card>
                  </template>

                  <template v-slot:item.5>
                    <v-card flat>
                      <v-card-text>
                        <div class="text-h6 mb-2">5️⃣ Test & Save</div>
                        <p>Click <strong>"Test Email"</strong> to verify it works.</p>
                        <p>If successful, click <strong>"Save Settings"</strong>.</p>

                        <v-alert type="success" variant="tonal" density="compact" class="mt-3">
                          <div class="text-caption">
                            ✅ <strong>You're done!</strong> You'll now get email alerts after every trade.
                          </div>
                        </v-alert>
                      </v-card-text>
                    </v-card>
                  </template>
                </v-stepper>

                <!-- Video Tutorial Link -->
                <v-card variant="outlined" color="primary" class="mt-4">
                  <v-card-text class="text-center">
                    <v-icon icon="mdi-youtube" size="large" color="red" class="mb-2"></v-icon>
                    <div class="text-h6 mb-2">Prefer Video?</div>
                    <p class="text-caption mb-3">Watch a 2-minute walkthrough</p>
                    <v-btn
                      href="https://www.youtube.com/watch?v=dQw4w9WgXcQ"
                      target="_blank"
                      color="primary"
                      variant="flat"
                      prepend-icon="mdi-play-circle"
                    >
                      Watch Tutorial Video
                    </v-btn>
                  </v-card-text>
                </v-card>

                <!-- Troubleshooting -->
                <v-divider class="my-4"></v-divider>

                <div class="text-h6 mb-2">❓ Troubleshooting</div>
                <v-expansion-panels variant="accordion">
                  <v-expansion-panel>
                    <v-expansion-panel-title>Test email not working?</v-expansion-panel-title>
                    <v-expansion-panel-text>
                      <v-list density="compact">
                        <v-list-item>✓ Make sure you're using App Password, not regular password</v-list-item>
                        <v-list-item>✓ Remove spaces from the App Password (paste as one word)</v-list-item>
                        <v-list-item>✓ Check your email address is correct</v-list-item>
                        <v-list-item>✓ Make sure 2FA is enabled on your Google account</v-list-item>
                      </v-list>
                    </v-expansion-panel-text>
                  </v-expansion-panel>

                  <v-expansion-panel>
                    <v-expansion-panel-title>Can't find App Passwords page?</v-expansion-panel-title>
                    <v-expansion-panel-text>
                      <p>You need to enable 2-Factor Authentication first:</p>
                      <ol class="ml-4">
                        <li>Go to <a href="https://myaccount.google.com/security" target="_blank" class="text-primary">Google Security Settings</a></li>
                        <li>Click "2-Step Verification"</li>
                        <li>Follow the setup wizard</li>
                        <li>Then App Passwords will appear in the same menu</li>
                      </ol>
                    </v-expansion-panel-text>
                  </v-expansion-panel>

                  <v-expansion-panel>
                    <v-expansion-panel-title>Using Outlook/Yahoo instead?</v-expansion-panel-title>
                    <v-expansion-panel-text>
                      <p><strong>Outlook:</strong> You may need to enable "Less secure apps" or create an App Password in your Microsoft account settings.</p>
                      <p class="mt-2"><strong>Yahoo:</strong> Generate an App Password at <a href="https://login.yahoo.com/account/security" target="_blank" class="text-primary">Yahoo Account Security</a></p>
                    </v-expansion-panel-text>
                  </v-expansion-panel>
                </v-expansion-panels>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </form>
      </v-card-text>

      <v-divider></v-divider>

      <v-card-actions>
        <v-btn
          variant="outlined"
          @click="testEmail"
          :disabled="!settings.enabled || loading || testLoading"
          :loading="testLoading"
          prepend-icon="mdi-email-fast"
        >
          Test Email
        </v-btn>
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
          @click="handleSave"
          :disabled="loading || testLoading"
          :loading="loading"
        >
          Save Settings
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import { ref, onMounted } from 'vue'
import { GetEmailSettings, SaveEmailSettings, TestEmail } from '../../wailsjs/go/main/App'

export default {
  name: 'EmailSettings',
  setup() {
    const dialog = ref(false)
    const showPassword = ref(false)
    const error = ref('')
    const success = ref('')
    const loading = ref(false)
    const testLoading = ref(false)
    const emailProvider = ref('gmail')

    const emailProviders = [
      { title: 'Gmail (Recommended)', value: 'gmail' },
      { title: 'Outlook / Hotmail', value: 'outlook' },
      { title: 'Yahoo Mail', value: 'yahoo' },
      { title: 'Custom SMTP', value: 'custom' }
    ]

    const tutorialSteps = [
      'Enable',
      'Provider',
      'App Password',
      'Fill Form',
      'Test & Save'
    ]

    const settings = ref({
      enabled: false,
      notificationEmail: '',
      smtpHost: 'smtp.gmail.com',
      smtpPort: 587,
      smtpFromEmail: '',
      smtpPassword: '',
      notifyOnDCABuy: true,      // Default enabled
      notifyOnDipBuy: true,      // Default enabled
      sendMonthlySummary: true   // Default enabled
    })

    const updateProviderSettings = (provider) => {
      // Auto-configure SMTP settings based on provider
      switch (provider) {
        case 'gmail':
          settings.value.smtpHost = 'smtp.gmail.com'
          settings.value.smtpPort = 587
          break
        case 'outlook':
          settings.value.smtpHost = 'smtp-mail.outlook.com'
          settings.value.smtpPort = 587
          break
        case 'yahoo':
          settings.value.smtpHost = 'smtp.mail.yahoo.com'
          settings.value.smtpPort = 587
          break
        case 'custom':
          // Leave as is for custom configuration
          break
      }
    }

    const loadSettings = async () => {
      try {
        const saved = await GetEmailSettings()
        if (saved) {
          settings.value = saved
        }
      } catch (err) {
        console.log('No saved email settings, using defaults')
      }
    }

    const handleSave = async () => {
      error.value = ''
      success.value = ''
      loading.value = true

      try {
        // Validate inputs
        if (settings.value.enabled) {
          if (!settings.value.notificationEmail) {
            error.value = 'Email address is required'
            loading.value = false
            return
          }

          if (!settings.value.smtpHost || !settings.value.smtpPort) {
            error.value = 'SMTP host and port are required'
            loading.value = false
            return
          }

          if (!settings.value.smtpFromEmail || !settings.value.smtpPassword) {
            error.value = 'SMTP from email and password are required'
            loading.value = false
            return
          }
        }

        // Save settings
        await SaveEmailSettings(settings.value)
        success.value = 'Email settings saved successfully!'

        // Close after delay
        setTimeout(() => {
          closeDialog()
        }, 1500)

      } catch (err) {
        error.value = err.toString().replace('Error: ', '')
      } finally {
        loading.value = false
      }
    }

    const testEmail = async () => {
      error.value = ''
      success.value = ''
      testLoading.value = true

      try {
        await TestEmail(settings.value)
        success.value = '✅ Test email sent! Check your inbox.'
      } catch (err) {
        error.value = 'Failed to send test email: ' + err.toString().replace('Error: ', '')
      } finally {
        testLoading.value = false
      }
    }

    const closeDialog = () => {
      dialog.value = false
      setTimeout(() => {
        error.value = ''
        success.value = ''
      }, 300)
    }

    onMounted(() => {
      loadSettings()
    })

    return {
      dialog,
      showPassword,
      emailProvider,
      emailProviders,
      tutorialSteps,
      settings,
      error,
      success,
      loading,
      testLoading,
      updateProviderSettings,
      handleSave,
      testEmail,
      closeDialog
    }
  }
}
</script>
