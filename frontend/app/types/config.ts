// frontend/app/types/config.ts
export const THEME_MODE = {
	AUTO: 'auto',
	LIGHT: 'light',
	DARK: 'dark',
	LIGHT_HIGH_CONTRAST: 'lightHighContrast',
	DARK_HIGH_CONTRAST: 'darkHighContrast',
} as const

export type ThemeMode = typeof THEME_MODE[keyof typeof THEME_MODE]
export type LocaleCode = 'en' | 'pl'
export const ACCENT_MODE = {
	AUTO: 'auto',
	CUSTOM: 'custom',
} as const

export type AccentMode = typeof ACCENT_MODE[keyof typeof ACCENT_MODE]

export const DEFAULT_THEME_MODE: ThemeMode = THEME_MODE.AUTO
export const DEFAULT_THEME_NAME: Exclude<ThemeMode, typeof THEME_MODE.AUTO> = THEME_MODE.LIGHT
export const DEFAULT_ACCENT_MODE: AccentMode = ACCENT_MODE.AUTO

export const THEME_MODES = Object.values(THEME_MODE) as ThemeMode[]
export const LOCALE_CODES: LocaleCode[] = ['en', 'pl']
export const ACCENT_MODES = Object.values(ACCENT_MODE) as AccentMode[]
