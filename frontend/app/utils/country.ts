// Shared country display helpers used by browse cards, detail dialogs, and filter UI.

/** Converts a 2-letter ISO country code to its flag emoji. */
export function countryFlag(cc: string | undefined | null): string {
  const upper = cc?.toUpperCase()
  if (!upper || upper.length !== 2) return ''
  return String.fromCodePoint(...[...upper].map(c => 0x1F1E6 + c.charCodeAt(0) - 65))
}

/** Returns the localized country name for a given ISO code. */
export function countryName(cc: string | undefined | null, locale: string): string {
  if (!cc) return ''
  try {
    return new Intl.DisplayNames([locale], { type: 'region' }).of(cc.toUpperCase()) ?? cc
  }
  catch {
    return cc
  }
}
