import {createApp} from 'vue'
import App from './App.vue'
import './style.css'

// Vuetify
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'neonDark',
    themes: {
      neonDark: {
        dark: true,
        colors: {
          // Base colors (muted neutrals)
          background: '#1A1A1A',    // Off-black for main background
          surface: '#2C2C2C',       // Charcoal for cards/panels

          // Neon accents (limited use for maximum impact)
          primary: '#9D00FF',       // Neon Purple (buttons, highlights)
          secondary: '#00FFFF',     // Neon Cyan (info, neutral elements)
          accent: '#9D00FF',        // Neon Purple (emphasis)

          // Status colors (neon variants)
          success: '#39FF14',       // Neon Green (profits, positive)
          error: '#FF073A',         // Neon Red/Pink (losses, negative)
          warning: '#FFFF33',       // Neon Yellow (alerts, caution)
          info: '#00FFFF',          // Neon Cyan (information)

          // Additional semantic colors
          'profit': '#39FF14',      // Explicit profit color
          'loss': '#FF073A',        // Explicit loss color
        },
      },
    },
  },
})

createApp(App).use(vuetify).mount('#app')
