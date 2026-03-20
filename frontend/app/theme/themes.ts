// app/theme/themes.ts
import type { ThemeDefinition } from 'vuetify'

const sharedVars = {
  'high-emphasis-opacity': 0.87,
  'medium-emphasis-opacity': 0.6,
  'disabled-opacity': 0.38,
  'border-opacity': 0.12,
  'hover-opacity': 0.04,
  'focus-opacity': 0.12,
  'selected-opacity': 0.08,
  'activated-opacity': 0.12,
  'pressed-opacity': 0.12,
  'app-bar-elevation': 0,
  'card-elevation': 0,
  'dialog-elevation': 0,
} as const

export const light: ThemeDefinition = {
  dark: false,
  colors: {
    'background': '#F5F7FF',
    'surface': '#FFFFFF',
    'surface-bright': '#FFFFFF',
    'surface-light': '#F3F4FD',
    'surface-variant': '#E4E7FB',
    'on-background': '#10111B',
    'on-surface': '#151726',
    'on-surface-variant': '#171827',
    'primary': '#7B5CDE',
    'secondary': '#C93D8E',
    'success': '#2EBF65',
    'warning': '#C79314',
    'error': '#E53955',
    'info': '#1C74D4',
  },
  variables: { ...sharedVars },
}

export const dark: ThemeDefinition = {
  dark: true,
  colors: {
    'background': '#10111B',
    'surface': '#181926',
    'surface-bright': '#202235',
    'surface-light': '#2A2E45',
    'surface-variant': '#2A3042',
    'on-background': '#ECEFF4',
    'on-surface': '#E4E7F5',
    'on-surface-variant': '#ECEFF4',
    'primary': '#BD93F9',
    'secondary': '#F57FB5',
    'success': '#5CF29B',
    'warning': '#F8E16B',
    'error': '#FF5C7C',
    'info': '#7DE3FF',
  },
  variables: { ...sharedVars },
}

export const lightHighContrast: ThemeDefinition = {
  dark: false,
  colors: {
    'background': '#FFFFFF',
    'surface': '#FFFFFF',
    'surface-bright': '#FFFFFF',
    'surface-light': '#FAFAFA',
    'surface-variant': '#F5F5F5',
    'on-background': '#000000',
    'on-surface': '#000000',
    'on-surface-variant': '#000000',
    'primary': '#4A00B0',
    'secondary': '#AD1457',
    'success': '#1B5E20',
    'warning': '#E65100',
    'error': '#B71C1C',
    'info': '#01579B',
  },
  variables: {
    ...sharedVars,
    'border-opacity': 0.4,
    'high-emphasis-opacity': 1,
    'medium-emphasis-opacity': 0.87,
    'disabled-opacity': 0.5,
    'border-width-root': '2px',
  },
}

export const darkHighContrast: ThemeDefinition = {
  dark: true,
  colors: {
    'background': '#000000',
    'surface': '#0D0D0D',
    'surface-bright': '#141414',
    'surface-light': '#1A1A1A',
    'surface-variant': '#1A1A1A',
    'on-background': '#FFFFFF',
    'on-surface': '#FFFFFF',
    'on-surface-variant': '#FFFFFF',
    'primary': '#E8D4FF',
    'secondary': '#FFB4D9',
    'success': '#69F0AE',
    'warning': '#FFE57F',
    'error': '#FF5252',
    'info': '#40C4FF',
  },
  variables: {
    ...sharedVars,
    'border-opacity': 0.4,
    'high-emphasis-opacity': 1,
    'medium-emphasis-opacity': 0.87,
    'disabled-opacity': 0.5,
    'border-width-root': '2px',
  },
}

export const themes = {
  light,
  dark,
  lightHighContrast,
  darkHighContrast,
} as const

export type ThemeName = keyof typeof themes
