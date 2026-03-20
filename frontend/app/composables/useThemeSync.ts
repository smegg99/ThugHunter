// app/composables/useThemeSync.ts
import { Events } from '@wailsio/runtime'
import * as ThemeService from '~~bindings/smegg.me/thughunter/services/theme/service.js'
import { Info as ThemeInfo } from '~~bindings/smegg.me/thughunter/common/theme/models.js'
import { useTheme } from 'vuetify'
import {
  argbFromHex,
  themeFromSourceColor,
  hexFromArgb,
} from '@material/material-color-utilities'
import {
  ACCENT_MODE,
  DEFAULT_ACCENT_MODE,
  DEFAULT_THEME_MODE,
  DEFAULT_THEME_NAME,
  THEME_MODE,
} from '~/types/config'
import type { LocaleCode } from '~/types/locale'
import type { ThemeMode,  AccentMode } from '~/types/config'

const EVENT_THEME_CHANGED = 'theme:changed'

const DEFAULT_ACCENT = '#3c00ff'

const themeInfo = reactive(new ThemeInfo())
let themeInitialized = false
const log = useLogger()

export function useThemeSync() {
  const { config } = useConfigSync()
  const vuetifyTheme = useTheme()
  const { locale, locales, setLocale } = useI18n()

  const localeItems = computed(() =>
    (locales.value as Array<{ code: string, name: string }>).map(l => ({
      code: l.code as LocaleCode,
      name: l.name,
    })),
  )

  const prefs = computed(() => config.preferences)

  function applyTheme(mode: ThemeMode) {
    if (mode === THEME_MODE.AUTO) {
      vuetifyTheme.global.name.value = themeInfo.dark_mode ? THEME_MODE.DARK : DEFAULT_THEME_NAME
    }
    else {
      vuetifyTheme.global.name.value = mode
    }
  }

  function applyLocale(code: LocaleCode) {
    if (locale.value !== code) {
      setLocale(code)
    }
  }

  watch(() => config.preferences.theme, (newTheme) => {
    applyTheme((newTheme || DEFAULT_THEME_MODE) as ThemeMode)
  })

  watch(() => config.preferences.language, (newLang) => {
    applyLocale((newLang || 'en') as LocaleCode)
  })

  watch(() => themeInfo.dark_mode, () => {
    if ((config.preferences.theme || DEFAULT_THEME_MODE) === THEME_MODE.AUTO) {
      applyTheme(THEME_MODE.AUTO)
    }
  })

  const effectiveAccent = computed(() => {
    const mode = (config.preferences.accent_mode || DEFAULT_ACCENT_MODE) as AccentMode
    if (mode === ACCENT_MODE.CUSTOM || !themeInfo.accent_color) {
      return config.preferences.accent_color || DEFAULT_ACCENT
    }
    return themeInfo.accent_color
  })

  const isAccentWatchSupported = computed(() => themeInfo.accent_watch_supported)

  function applyAccent(hex: string) {
    const sourceArgb = argbFromHex(hex)
    const materialTheme = themeFromSourceColor(sourceArgb)

    const currentThemeName = vuetifyTheme.global.name.value
    const themeDef = vuetifyTheme.global.current.value
    const isDark = themeDef.dark

    const scheme = isDark ? materialTheme.schemes.dark : materialTheme.schemes.light

    const colors: Record<string, string> = {
      'primary': hexFromArgb(scheme.primary),
      'on-primary': hexFromArgb(scheme.onPrimary),
      'primary-darken-1': hexFromArgb(isDark ? scheme.primaryContainer : scheme.onPrimaryContainer),
      'secondary': hexFromArgb(scheme.secondary),
      'on-secondary': hexFromArgb(scheme.onSecondary),
      'secondary-darken-1': hexFromArgb(isDark ? scheme.secondaryContainer : scheme.onSecondaryContainer),
      'error': hexFromArgb(scheme.error),
      'on-error': hexFromArgb(scheme.onError),
      'accent': hexFromArgb(scheme.tertiary),
      'on-accent': hexFromArgb(scheme.onTertiary),
    }

    const softColors: Record<string, string> = {
      'primary-soft': hexFromArgb(isDark ? scheme.primaryContainer : scheme.primaryContainer),
      'primary-alt': hexFromArgb(isDark ? scheme.onPrimaryContainer : scheme.onPrimaryContainer),
      'secondary-soft': hexFromArgb(isDark ? scheme.secondaryContainer : scheme.secondaryContainer),
      'secondary-alt': hexFromArgb(isDark ? scheme.onSecondaryContainer : scheme.onSecondaryContainer),
      'accent-soft': hexFromArgb(isDark ? scheme.tertiaryContainer : scheme.tertiaryContainer),
      'accent-alt': hexFromArgb(isDark ? scheme.onTertiaryContainer : scheme.onTertiaryContainer),
    }

    const theme = vuetifyTheme.themes.value[currentThemeName]
    if (theme?.colors) {
      Object.assign(theme.colors, colors, softColors)
    }
  }

  watch(effectiveAccent, (hex) => {
    if (hex) applyAccent(hex)
  })

  // Re-apply when accent mode changes, even if the resolved hex is identical
  watch(() => config.preferences.accent_mode, () => {
    if (effectiveAccent.value) applyAccent(effectiveAccent.value)
  })

  watch(() => vuetifyTheme.global.name.value, () => {
    if (effectiveAccent.value) applyAccent(effectiveAccent.value)
  })

  let offThemeChanged: (() => void) | undefined

  onMounted(async () => {
    if (!themeInitialized) {
      themeInitialized = true
      try {
        const info = await ThemeService.GetTheme()
        Object.assign(themeInfo, info)
      }
      catch (err) {
        log.error('failed to load theme info', { error: String(err) })
      }
    }

    offThemeChanged = Events.On(EVENT_THEME_CHANGED, (ev: { data: any }) => {
      Object.assign(themeInfo, ThemeInfo.createFrom(ev.data))
    })

    applyTheme((config.preferences.theme || DEFAULT_THEME_MODE) as ThemeMode)
    applyLocale((config.preferences.language || 'en') as LocaleCode)
    if (effectiveAccent.value) applyAccent(effectiveAccent.value)
  })

  onUnmounted(() => {
    offThemeChanged?.()
  })

  return { prefs, localeItems, effectiveAccent, isAccentWatchSupported }
}
