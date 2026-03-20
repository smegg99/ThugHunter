// app/types/config.ts
export const THEME_MODE = {
	AUTO: 'auto',
	LIGHT: 'light',
	DARK: 'dark',
	LIGHT_HIGH_CONTRAST: 'lightHighContrast',
	DARK_HIGH_CONTRAST: 'darkHighContrast',
} as const

export type ThemeMode = typeof THEME_MODE[keyof typeof THEME_MODE]

export const ACCENT_MODE = {
	AUTO: 'auto',
	CUSTOM: 'custom',
} as const

export type AccentMode = typeof ACCENT_MODE[keyof typeof ACCENT_MODE]

export const DEFAULT_THEME_MODE: ThemeMode = THEME_MODE.AUTO
export const DEFAULT_THEME_NAME: Exclude<ThemeMode, typeof THEME_MODE.AUTO> = THEME_MODE.LIGHT
export const DEFAULT_ACCENT_MODE: AccentMode = ACCENT_MODE.CUSTOM

export const THEME_MODES = Object.values(THEME_MODE) as ThemeMode[]
export const ACCENT_MODES = Object.values(ACCENT_MODE) as AccentMode[]

export type BaseTheme = 'light' | 'dark'
export type ThemeVariant = 'normal' | 'highContrast'

export const BASE_THEMES: { value: BaseTheme, icon: string, labelKey: string }[] = [
	{ value: 'light', icon: 'mdi-white-balance-sunny', labelKey: 'settings.themes.light' },
	{ value: 'dark', icon: 'mdi-moon-waning-crescent', labelKey: 'settings.themes.dark' },
]

export const THEME_VARIANTS: { value: ThemeVariant, icon: string, labelKey: string }[] = [
	{ value: 'normal', icon: 'mdi-circle-half-full', labelKey: 'settings.variants.normal' },
	{ value: 'highContrast', icon: 'mdi-contrast-box', labelKey: 'settings.variants.highContrast' },
]

export function splitThemeMode(mode: ThemeMode): { base: BaseTheme, variant: ThemeVariant } {
	switch (mode) {
		case THEME_MODE.DARK: return { base: 'dark', variant: 'normal' }
		case THEME_MODE.LIGHT_HIGH_CONTRAST: return { base: 'light', variant: 'highContrast' }
		case THEME_MODE.DARK_HIGH_CONTRAST: return { base: 'dark', variant: 'highContrast' }
		case THEME_MODE.LIGHT:
		case THEME_MODE.AUTO:
		default: return { base: 'light', variant: 'normal' }
	}
}

export function joinThemeMode(base: BaseTheme, variant: ThemeVariant): ThemeMode {
	if (base === 'dark' && variant === 'highContrast') return THEME_MODE.DARK_HIGH_CONTRAST
	if (base === 'dark') return THEME_MODE.DARK
	if (variant === 'highContrast') return THEME_MODE.LIGHT_HIGH_CONTRAST
	return THEME_MODE.LIGHT
}
