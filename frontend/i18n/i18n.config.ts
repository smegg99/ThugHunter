// i18n/i18n.config.ts
export default defineI18nConfig(() => ({
  legacy: false,
  fallbackLocale: 'en',

  missingWarn: import.meta.dev,
  fallbackWarn: import.meta.dev,

  datetimeFormats: {
    en: {
      short: { year: 'numeric', month: 'short', day: '2-digit' },
      medium: { year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' },
    },
    pl: {
      short: { year: 'numeric', month: 'short', day: '2-digit' },
      medium: { year: 'numeric', month: 'short', day: '2-digit', hour: '2-digit', minute: '2-digit', hourCycle: 'h23' },
    },
  },
  numberFormats: {
    en: { currency: { style: 'currency', currency: 'USD' } },
    pl: { currency: { style: 'currency', currency: 'PLN' } },
  },
}))
