// app/composables/useConfigSync.ts
import { Events } from '@wailsio/runtime'
import * as ConfigService from '~~bindings/smegg.me/thughunter/services/config/service.js'
import { Config } from '~~bindings/smegg.me/thughunter/common/config/models.js'

const EVENT_CONFIG_CHANGED = 'config:changed'

const config = reactive(new Config())
let initialized = false
let updatingFromBackend = false
let saveTimer: ReturnType<typeof setTimeout> | null = null

export function useConfigSync() {
  const log = useLogger()

  // Persists config changes to the backend whenever the reactive state changes.
  // Debounced to avoid sending intermediate states (e.g. empty number fields
  // while the user is typing).
  watch(config, () => {
    if (updatingFromBackend) return
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(async () => {
      saveTimer = null
      try {
        await ConfigService.SetConfig(new Config(config))
      }
      catch (err) {
        log.error('config: failed to save', { error: String(err) })
      }
    }, 400)
  }, { deep: true })

  let offConfigChanged: (() => void) | undefined

  onMounted(async () => {
    if (!initialized) {
      initialized = true
      try {
        const c = await ConfigService.GetConfig()
        updatingFromBackend = true
        Object.assign(config, c)
        nextTick(() => { updatingFromBackend = false })
        log.debug('config: loaded from backend')
      }
      catch (err) {
        log.error('config: failed to load', { error: String(err) })
      }
    }

    offConfigChanged = Events.On(EVENT_CONFIG_CHANGED, (ev: { data: unknown }) => {
      updatingFromBackend = true
      Object.assign(config, Config.createFrom(ev.data))
      nextTick(() => { updatingFromBackend = false })
      log.debug('config: updated from backend event')
    })
  })

  onUnmounted(() => {
    offConfigChanged?.()
  })

  return { config }
}
