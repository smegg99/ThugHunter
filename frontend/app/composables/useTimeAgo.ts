// app/composables/useTimeAgo.ts
const UNITS: [string, number][] = [
  ['year', 365 * 24 * 60 * 60],
  ['month', 30 * 24 * 60 * 60],
  ['week', 7 * 24 * 60 * 60],
  ['day', 24 * 60 * 60],
  ['hour', 60 * 60],
  ['minute', 60],
  ['second', 1],
]

export function useTimeAgo() {
  const { t } = useI18n()
  const now = ref(Date.now())
  let timer: ReturnType<typeof setInterval> | null = null

  onMounted(() => {
    timer = setInterval(() => { now.value = Date.now() }, 30_000)
  })

  onUnmounted(() => {
    if (timer) clearInterval(timer)
  })

  function timeAgo(date: string | Date | null | undefined): string {
    if (!date) return '-'
    const elapsed = (now.value - new Date(date).getTime()) / 1000
    if (Math.abs(elapsed) < 5) return t('timeAgo.justNow')

    const past = elapsed > 0
    const abs = Math.abs(elapsed)

    for (const [unit, seconds] of UNITS) {
      if (abs >= seconds) {
        const count = Math.floor(abs / seconds)
        return past
          ? t('timeAgo.past', { count, unit: t(`timeAgo.units.${unit}`, count) })
          : t('timeAgo.future', { count, unit: t(`timeAgo.units.${unit}`, count) })
      }
    }
    return t('timeAgo.justNow')
  }

  return { timeAgo }
}
