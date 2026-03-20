// app/composables/useMonitor.ts
import { Events } from '@wailsio/runtime'
import * as MonitorService from '~~bindings/smegg.me/thughunter/services/monitor/service.js'
import type { SystemSnapshot } from '~~bindings/smegg.me/thughunter/services/monitor/models.js'

const EVENT_SYSTEM = 'monitor:system'
const POLL_INTERVAL = 1500
const HISTORY_LENGTH = 30

const snapshot = ref<SystemSnapshot | null>(null)
const cpuHistory = ref<number[]>(new Array(HISTORY_LENGTH).fill(0))

export function useMonitor() {
  const log = useLogger()
  const cleanups: (() => void)[] = []

  function pushHistory(totalPercent: number) {
    cpuHistory.value = [...cpuHistory.value.slice(-(HISTORY_LENGTH - 1)), Math.round(totalPercent)]
  }

  onMounted(async () => {
    try {
      const s = await MonitorService.GetSnapshot()
      if (s) {
        snapshot.value = s
        pushHistory(s.cpu.totalPercent)
      }
      log.debug('monitor: initial snapshot loaded')
    }
    catch (err) {
      log.error('monitor: failed to get initial snapshot', { error: String(err) })
    }

    MonitorService.StartPolling(POLL_INTERVAL)

    cleanups.push(
      Events.On(EVENT_SYSTEM, (ev: { data: SystemSnapshot }) => {
        if (ev.data) {
          snapshot.value = ev.data
          pushHistory(ev.data.cpu?.totalPercent ?? 0)
        }
      }) ?? (() => {}),
    )
  })

  onUnmounted(() => {
    MonitorService.StopPolling()
    cleanups.forEach(fn => fn())
  })

  return { snapshot, cpuHistory }
}
