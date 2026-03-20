// app/types/locale.ts
export type LocaleCode = 'en' | 'pl'

export interface LocaleInfo {
  code: LocaleCode
  name: string
  language: string
  files: string[]
}

export const LOCALE_CODES = ['en', 'pl'] as const satisfies readonly LocaleCode[]

export const DEFAULT_LOCALE: LocaleCode = 'en'
