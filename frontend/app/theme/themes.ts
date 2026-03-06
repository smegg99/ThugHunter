// frontend/app/theme/themes.ts
import type { ThemeDefinition } from 'vuetify'

const baseVars = {
  'border-opacity': 0.12,
  'high-emphasis-opacity': 0.87,
  'medium-emphasis-opacity': 0.6,
  'disabled-opacity': 0.38,
  'idle-opacity': 0.04,
  'hover-opacity': 0.04,
  'focus-opacity': 0.12,
  'selected-opacity': 0.08,
  'activated-opacity': 0.12,
  'pressed-opacity': 0.12,
  'dragged-opacity': 0.08,
  'theme-kbd': '#212529',
  'theme-on-kbd': '#FFFFFF',
} as const

const baseVarsDark = {
  ...baseVars,
  'idle-opacity': 0.1,
  'theme-code': '#2B2B2B',
  'theme-on-code': '#CCCCCC',
} as const

const baseVarsLight = {
  ...baseVars,
  'theme-code': '#F5F5F5',
  'theme-on-code': '#000000',
} as const

const baseVarsHighContrast = {
  'border-opacity': 0.4,
  'high-emphasis-opacity': 1,
  'medium-emphasis-opacity': 0.87,
  'disabled-opacity': 0.5,
  'idle-opacity': 0.1,
  'hover-opacity': 0.08,
  'focus-opacity': 0.16,
  'selected-opacity': 0.16,
  'activated-opacity': 0.2,
  'pressed-opacity': 0.2,
  'dragged-opacity': 0.12,
} as const

export const light: ThemeDefinition = {
  dark: false,
  colors: {
    'background': '#F5F7FF',
    'surface': '#FFFFFF',
    'primary': '#7B5CDE',
    'primary-darken-1': '#5938C0',
    'secondary': '#C93D8E',
    'secondary-darken-1': '#8F2C64',
    'success': '#2EBF65',
    'warning': '#C79314',
    'error': '#E53955',
    'info': '#1C74D4',

    'surface-bright': '#FFFFFF',
    'surface-light': '#F3F4FD',
    'surface-variant': '#E4E7FB',
    'on-surface-variant': '#171827',

    'on-background': '#10111B',
    'on-surface': '#151726',
    'on-primary': '#F7FBFF',
    'on-secondary': '#FFF6FB',
    'on-success': '#FFFFFF',
    'on-warning': '#000000',
    'on-error': '#FFFFFF',
    'on-info': '#FFFFFF',

    'background-alt': '#EDF0FF',
    'surface-soft': '#F3F4FD',
    'surface-alt': '#E4E7FB',

    'primary-soft': '#9A7FF0',
    'primary-alt': '#5938C0',

    'secondary-soft': '#E76AAE',
    'secondary-alt': '#8F2C64',

    'accent': '#7E5CED',
    'accent-soft': '#B29CFF',
    'accent-alt': '#FFB86C',
    'on-accent': '#100B1E',

    'foreground': '#0D0E17',
    'text-primary': '#171827',
    'text-secondary': '#3A3F5C',
    'text-muted': '#7B82A9',
    'comment': '#9AA4D3',

    'success-soft': '#E6F9EE',
    'warning-soft': '#FFF7D9',
    'error-soft': '#FFE7ED',
    'info-soft': '#E0F0FF',

    'border': '#D1D5F0',
    'border-muted': '#E2E6F6',
    'divider': '#D8DCF5',

    'input-background': '#FFFFFF',
    'input-border': '#CBD2F0',
    'input-placeholder': '#9AA0C2',

    'navbar': '#F5F7FF',
    'navbar-active': '#2F8FA7',
    'navbar-hover': '#E6E9FB',

    'scrollbar': '#C3CAE8',
    'scrollbar-track': '#EFF1FD',
  },
  variables: {
    ...baseVarsLight,

    'font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'body-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'font-family-base':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'heading-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'code-font-family':
      '\'JetBrains Mono\', \'Fira Code\', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \'Liberation Mono\', \'Courier New\', monospace',

    'font-size-root': '16px',
    'line-height-root': 1.5,

    'border-radius-root': '0.75rem',
    'border-radius-sm': '0.375rem',
    'border-radius-lg': '1rem',
    'border-radius-pill': '999px',
    'chip-border-radius': '999px',
    'button-border-radius': '999px',

    'border-width-root': '1px',
    'overlay-scrim-background': '#000000',
    'overlay-scrim-opacity': 0.5,

    'button-font-weight': 600,
    'button-letter-spacing': '0.04em',
    'button-text-transform': 'none',

    'app-bar-elevation': 0,
    'card-elevation': 1,
    'dialog-elevation': 16,
  },
}

export const dark: ThemeDefinition = {
  dark: true,
  colors: {
    'background': '#10111B',
    'surface': '#181926',
    'primary': '#BD93F9',
    'primary-darken-1': '#7A5BCF',
    'secondary': '#F57FB5',
    'secondary-darken-1': '#B15491',
    'success': '#5CF29B',
    'warning': '#F8E16B',
    'error': '#FF5C7C',
    'info': '#7DE3FF',

    'surface-bright': '#202235',
    'surface-light': '#2A2E45',
    'surface-variant': '#2A3042',
    'on-surface-variant': '#ECEFF4',

    'on-background': '#ECEFF4',
    'on-surface': '#E4E7F5',
    'on-primary': '#071016',
    'on-secondary': '#190815',
    'on-success': '#071016',
    'on-warning': '#071016',
    'on-error': '#071016',
    'on-info': '#071016',

    'background-alt': '#151726',
    'surface-soft': '#202235',
    'surface-alt': '#2A2E45',

    'primary-soft': '#9F7BEB',
    'primary-alt': '#7A5BCF',

    'secondary-soft': '#D76AAE',
    'secondary-alt': '#B15491',

    'accent': '#C5A3FF',
    'accent-soft': '#9D7CFF',
    'accent-alt': '#FFB86C',
    'on-accent': '#120B1F',

    'foreground': '#F8F8F2',
    'text-primary': '#ECEFF4',
    'text-secondary': '#A5B3D6',
    'text-muted': '#6C7393',
    'comment': '#6272A4',

    'success-soft': '#163329',
    'warning-soft': '#3A321A',
    'error-soft': '#3F1C2A',
    'info-soft': '#12303B',

    'border': '#2B3045',
    'border-muted': '#1E2233',
    'divider': '#2A3042',

    'input-background': '#151727',
    'input-border': '#303552',
    'input-placeholder': '#70789F',

    'navbar': '#0B0C14',
    'navbar-active': '#F57FB5',
    'navbar-hover': '#2A3045',

    'scrollbar': '#2F3450',
    'scrollbar-track': '#11121A',
  },
  variables: {
    ...baseVarsDark,

    'font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'body-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'font-family-base':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'heading-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'code-font-family':
      '\'JetBrains Mono\', \'Fira Code\', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \'Liberation Mono\', \'Courier New\', monospace',

    'font-size-root': '16px',
    'line-height-root': 1.5,

    'border-radius-root': '0.75rem',
    'border-radius-sm': '0.375rem',
    'border-radius-lg': '1rem',
    'border-radius-pill': '999px',
    'chip-border-radius': '999px',
    'button-border-radius': '999px',

    'border-width-root': '1px',
    'overlay-scrim-background': '#000000',
    'overlay-scrim-opacity': 0.65,

    'button-font-weight': 600,
    'button-letter-spacing': '0.04em',
    'button-text-transform': 'none',

    'app-bar-elevation': 0,
    'card-elevation': 2,
    'dialog-elevation': 24,
  },
}

export const lightHighContrast: ThemeDefinition = {
  dark: false,
  colors: {
    'background': '#FFFFFF',
    'surface': '#FFFFFF',
    'primary': '#4A00B0',
    'primary-darken-1': '#3700B3',
    'secondary': '#AD1457',
    'secondary-darken-1': '#880E4F',
    'success': '#1B5E20',
    'warning': '#E65100',
    'error': '#B71C1C',
    'info': '#01579B',

    'surface-bright': '#FFFFFF',
    'surface-light': '#FAFAFA',
    'surface-variant': '#F5F5F5',
    'on-surface-variant': '#000000',

    'on-background': '#000000',
    'on-surface': '#000000',
    'on-primary': '#FFFFFF',
    'on-secondary': '#FFFFFF',
    'on-success': '#FFFFFF',
    'on-warning': '#FFFFFF',
    'on-error': '#FFFFFF',
    'on-info': '#FFFFFF',

    'background-alt': '#FAFAFA',
    'surface-soft': '#FAFAFA',
    'surface-alt': '#F5F5F5',

    'primary-soft': '#6200EA',
    'primary-alt': '#3700B3',

    'secondary-soft': '#C2185B',
    'secondary-alt': '#880E4F',

    'accent': '#0050B3',
    'accent-soft': '#1565C0',
    'accent-alt': '#E65100',
    'on-accent': '#FFFFFF',

    'foreground': '#000000',
    'text-primary': '#000000',
    'text-secondary': '#212121',
    'text-muted': '#424242',
    'comment': '#616161',

    'success-soft': '#E8F5E9',
    'warning-soft': '#FFF3E0',
    'error-soft': '#FFEBEE',
    'info-soft': '#E1F5FE',

    'border': '#212121',
    'border-muted': '#424242',
    'divider': '#424242',

    'input-background': '#FFFFFF',
    'input-border': '#212121',
    'input-placeholder': '#616161',

    'navbar': '#FFFFFF',
    'navbar-active': '#4A00B0',
    'navbar-hover': '#F5F5F5',

    'scrollbar': '#424242',
    'scrollbar-track': '#FAFAFA',

    'focus-ring': '#0050B3',
  },
  variables: {
    ...baseVarsLight,
    ...baseVarsHighContrast,

    'font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'body-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'font-family-base':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'heading-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'code-font-family':
      '\'JetBrains Mono\', \'Fira Code\', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \'Liberation Mono\', \'Courier New\', monospace',

    'font-size-root': '16px',
    'line-height-root': 1.5,

    'border-radius-root': '0.75rem',
    'border-radius-sm': '0.375rem',
    'border-radius-lg': '1rem',
    'border-radius-pill': '999px',
    'chip-border-radius': '999px',
    'button-border-radius': '999px',

    'border-width-root': '2px',
    'overlay-scrim-background': '#000000',
    'overlay-scrim-opacity': 0.6,

    'button-font-weight': 600,
    'button-letter-spacing': '0.04em',
    'button-text-transform': 'none',

    'app-bar-elevation': 0,
    'card-elevation': 0,
    'dialog-elevation': 16,
  },
}

export const darkHighContrast: ThemeDefinition = {
  dark: true,
  colors: {
    'background': '#000000',
    'surface': '#0D0D0D',
    'primary': '#E8D4FF',
    'primary-darken-1': '#BB94FF',
    'secondary': '#FFB4D9',
    'secondary-darken-1': '#FF5CAF',
    'success': '#69F0AE',
    'warning': '#FFE57F',
    'error': '#FF5252',
    'info': '#40C4FF',

    'surface-bright': '#141414',
    'surface-light': '#1A1A1A',
    'surface-variant': '#1A1A1A',
    'on-surface-variant': '#FFFFFF',

    'on-background': '#FFFFFF',
    'on-surface': '#FFFFFF',
    'on-primary': '#000000',
    'on-secondary': '#000000',
    'on-success': '#000000',
    'on-warning': '#000000',
    'on-error': '#000000',
    'on-info': '#000000',

    'background-alt': '#0A0A0A',
    'surface-soft': '#141414',
    'surface-alt': '#1A1A1A',

    'primary-soft': '#D4B8FF',
    'primary-alt': '#BB94FF',

    'secondary-soft': '#FF8AC5',
    'secondary-alt': '#FF5CAF',

    'accent': '#80EEFF',
    'accent-soft': '#4DE5FF',
    'accent-alt': '#FFE066',
    'on-accent': '#000000',

    'foreground': '#FFFFFF',
    'text-primary': '#FFFFFF',
    'text-secondary': '#E0E0E0',
    'text-muted': '#BDBDBD',
    'comment': '#9E9E9E',

    'success-soft': '#0A1F14',
    'warning-soft': '#1F1A0A',
    'error-soft': '#1F0A0A',
    'info-soft': '#0A151F',

    'border': '#666666',
    'border-muted': '#444444',
    'divider': '#555555',

    'input-background': '#0A0A0A',
    'input-border': '#888888',
    'input-placeholder': '#BDBDBD',

    'navbar': '#000000',
    'navbar-active': '#FFE066',
    'navbar-hover': '#1A1A1A',

    'scrollbar': '#666666',
    'scrollbar-track': '#0A0A0A',

    'focus-ring': '#FFE066',
  },
  variables: {
    ...baseVarsDark,
    ...baseVarsHighContrast,

    'font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'body-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'font-family-base':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'heading-font-family':
      '\'Overpass\', system-ui, -apple-system, BlinkMacSystemFont, \'Segoe UI\', sans-serif',
    'code-font-family':
      '\'JetBrains Mono\', \'Fira Code\', ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, \'Liberation Mono\', \'Courier New\', monospace',

    'font-size-root': '16px',
    'line-height-root': 1.5,

    'border-radius-root': '0.75rem',
    'border-radius-sm': '0.375rem',
    'border-radius-lg': '1rem',
    'border-radius-pill': '999px',
    'chip-border-radius': '999px',
    'button-border-radius': '999px',

    'border-width-root': '2px',
    'overlay-scrim-background': '#000000',
    'overlay-scrim-opacity': 0.8,

    'button-font-weight': 600,
    'button-letter-spacing': '0.04em',
    'button-text-transform': 'none',

    'app-bar-elevation': 0,
    'card-elevation': 0,
    'dialog-elevation': 24,
  },
}

export const themes = {
  light,
  dark,
  lightHighContrast,
  darkHighContrast,
} as const

export type ThemeName = keyof typeof themes
