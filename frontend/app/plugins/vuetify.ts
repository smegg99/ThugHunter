// app/plugins/vuetify.ts
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import { md3 } from 'vuetify/blueprints'
import { themes } from '~/theme/themes'
import { DEFAULT_THEME_NAME } from '~/types/config'
import defaults from '~/theme/defaults'

export default defineNuxtPlugin((app) => {
  const vuetify = createVuetify({
    blueprint: md3,
    theme: {
      defaultTheme: DEFAULT_THEME_NAME,
      themes,
      variations: {
        colors: ['primary', 'secondary', 'error', 'info', 'success', 'warning'],
        lighten: 2,
        darken: 2,
      },
    },
    defaults: defaults,
  })
  app.vueApp.use(vuetify)
})
