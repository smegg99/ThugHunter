import { fileURLToPath } from 'node:url'
import path from 'node:path'
import vuetify, { transformAssetUrls } from 'vite-plugin-vuetify'
import wailsPlugin from '@wailsio/runtime/plugins/vite'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const bindingsDir = path.resolve(__dirname, 'bindings')

const localeFiles = ['common.json']

export default defineNuxtConfig({
  app: {
    head: {
      style: [
        {
          // Prevent flash before Vuetify/JS loads. The actual theme is
          // applied from the backend config once the app mounts.
          innerHTML: `html, body { background-color: #F5F7FF; }`,
        },
      ],
    },
  },
  modules: [
    '@unocss/nuxt',
    '@nuxtjs/i18n',
    '@pinia/nuxt',
    'pinia-plugin-persistedstate/nuxt',
    '@nuxt/eslint',
  ],
  build: {
    transpile: ['vuetify'],
  },
  ssr: false,
  devServer: { port: 9245 },
  devtools: { enabled: true },
  css: ['~/assets/css/fonts.css'],
  features: {
    inlineStyles: false,
  },
  compatibilityDate: '2025-07-15',
  alias: {
    '~~bindings': bindingsDir,
  },
  nitro: {
    compressPublicAssets: true,
  },
  vite: {
    plugins: [
      wailsPlugin(bindingsDir) as any,
      vuetify({
        autoImport: true,
        styles: { configFile: 'assets/css/vuetify.scss' },
      }),
    ],
    vue: {
      template: {
        transformAssetUrls,
      },
    },
  },
  eslint: {
    config: {
      stylistic: true,
    },
  },
  i18n: {
    strategy: 'prefix',
    defaultLocale: 'en',

    langDir: 'locales',

    locales: [
      {
        code: 'en',
        name: 'English',
        language: 'en-US',
        files: localeFiles.map(f => `en/${f}`),
      },
      {
        code: 'pl',
        name: 'Polski',
        language: 'pl-PL',
        files: localeFiles.map(f => `pl/${f}`),
      },
    ],

    compilation: {
      strictMessage: false,
      escapeHtml: false,
    },

    detectBrowserLanguage: false,

    vueI18n: './i18n/i18n.config.ts',
  },
})
