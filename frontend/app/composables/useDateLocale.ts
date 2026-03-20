// app/composables/useDateLocale.ts
export function useDateLocale() {
  const { locale, locales } = useI18n()

  return computed(() => {
    const current = (locales.value as Array<{ code: string; language?: string }>)
      .find(l => l.code === locale.value)
    return current?.language || locale.value
  })
}
