// Shared color helpers for latency/progress indicators.

/** Returns a Vuetify color name based on ping latency. */
export function pingColor(ms: number): string {
  if (ms <= 50) return 'success'
  if (ms <= 150) return 'warning'
  return 'error'
}

/** Returns a Vuetify color name based on a percentage value (RAM, CPU, progress bars). */
export function percentColor(pct: number, thresholds = { warn: 60, danger: 85 }, low = 'primary'): string {
  if (pct > thresholds.danger) return 'error'
  if (pct > thresholds.warn) return 'warning'
  return low
}

/** Returns a Vuetify color name for a success rate (0–1). */
export function successRateColor(rate: number): string {
  if (rate < 0) return 'primary'
  if (rate >= 0.8) return 'success'
  if (rate >= 0.5) return 'warning'
  return 'error'
}
