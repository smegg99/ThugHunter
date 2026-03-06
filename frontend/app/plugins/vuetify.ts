// frontend/app/plugins/vuetify.ts
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import { themes } from '~/theme/themes'
import { DEFAULT_THEME_NAME } from '~/types/config'

export default defineNuxtPlugin((app) => {
  const vuetify = createVuetify({
    theme: {
      defaultTheme: DEFAULT_THEME_NAME,
      themes,
      variations: {
        colors: ['primary', 'secondary', 'error', 'info', 'success', 'warning'],
        lighten: 2,
        darken: 2,
      },
    },
  })
  app.vueApp.use(vuetify)
})
