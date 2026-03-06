export default defineI18nConfig(() => ({
  legacy: false,
  fallbackLocale: 'en',

  missingWarn: import.meta.dev,
  fallbackWarn: import.meta.dev,

  datetimeFormats: {
    en: { short: { year: 'numeric', month: 'short', day: '2-digit' } },
    pl: { short: { year: 'numeric', month: 'short', day: '2-digit' } },
  },
  numberFormats: {
    en: { currency: { style: 'currency', currency: 'USD' } },
    pl: { currency: { style: 'currency', currency: 'PLN' } },
  },
}))
