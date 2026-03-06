import { Events } from '@wailsio/runtime'
import * as ConfigService from '~~bindings/smegg.me/thughunter/services/config/service.js'
import { Config } from '~~bindings/smegg.me/thughunter/common/config/models.js'

const EVENT_CONFIG_CHANGED = 'config:changed'

const config = reactive(new Config())
let initialized = false
let updatingFromBackend = false

// This composable synchronizes the full application config between frontend and backend. It loads the config from the backend on first mount, watches for deep changes to push updates back, and listens for backend-emitted config changes to keep the frontend in sync. I intend this to be sorta like a base class for configs, so that other scope specific parts of the config have their own composables that use this one and react to changes in the relevant config sections (e.g., useThemeSync for theme, accent, and language preferences).
export function useConfigSync() {
  watch(config, async () => {
    if (updatingFromBackend) return
    try {
      await ConfigService.SetConfig(new Config(config))
    }
    catch (err) {
      console.error('failed to save config', err)
    }
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
      }
      catch (err) {
        console.error('failed to load config', err)
      }
    }

    offConfigChanged = Events.On(EVENT_CONFIG_CHANGED, (ev: { data: any }) => {
      updatingFromBackend = true
      Object.assign(config, Config.createFrom(ev.data))
      nextTick(() => { updatingFromBackend = false })
    })
  })

  onUnmounted(() => {
    offConfigChanged?.()
  })

  return { config }
}
