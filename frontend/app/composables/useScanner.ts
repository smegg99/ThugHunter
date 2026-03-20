// app/composables/useScanner.ts
import { Events, Call } from '@wailsio/runtime'
import type { ScanProgressData, HostItem, HostPage, VNCPage, FilterOptions, BrowseStats, ScreenshotResult, AuthFilter, ScreenshotFilter, ScanMode } from '~/types/scanner'

const SERVICE = 'smegg.me/thughunter/services/scanner.Service'

const EVENT_SCAN_STARTED = 'scanner:scan_started'
const EVENT_SCAN_PROGRESS = 'scanner:scan_progress'
const EVENT_SCAN_COMPLETED = 'scanner:scan_completed'
const EVENT_SCAN_ERROR = 'scanner:scan_error'
const POLL_INTERVAL_MS = 200

// Module-level state shared across all composable instances.
const running = ref(false)
const scanning = ref(false)
const stopping = ref(false)
const scanMode = ref<ScanMode | ''>('')
const progress = ref<ScanProgressData | null>(null)
const log = useLogger()

export function useScanner() {
  const cleanups: (() => void)[] = []
  let pollTimer: ReturnType<typeof setInterval> | null = null
  let autoClearTimer: ReturnType<typeof setTimeout> | null = null

  async function refresh() {
    try {
      const [isRunning, prog] = await Promise.all([
        Call.ByName(`${SERVICE}.IsRunning`) as Promise<boolean>,
        Call.ByName(`${SERVICE}.GetProgress`) as Promise<ScanProgressData | null>,
      ])
      running.value = isRunning
      progress.value = prog
      scanMode.value = prog?.mode || ''
    }
    catch (err) {
      log.error('useScanner: failed to refresh', { error: String(err) })
    }
  }

  async function rescan() {
    scanning.value = true
    progress.value = null
    if (autoClearTimer) {
      clearTimeout(autoClearTimer)
      autoClearTimer = null
    }
    try {
      await Call.ByName(`${SERVICE}.Start`)
      running.value = true
      scanMode.value = 'hosts'
      startPolling()
    }
    catch (err) {
      log.error('useScanner: failed to start scan', { error: String(err) })
    }
    finally {
      scanning.value = false
    }
  }

  async function scanHosts() {
    scanning.value = true
    progress.value = null
    if (autoClearTimer) {
      clearTimeout(autoClearTimer)
      autoClearTimer = null
    }
    try {
      await Call.ByName(`${SERVICE}.StartHostScan`)
      running.value = true
      scanMode.value = 'hosts'
      startPolling()
    }
    catch (err) {
      log.error('useScanner: failed to start host scan', { error: String(err) })
    }
    finally {
      scanning.value = false
    }
  }

  async function scanScreenshots() {
    scanning.value = true
    progress.value = null
    if (autoClearTimer) {
      clearTimeout(autoClearTimer)
      autoClearTimer = null
    }
    try {
      await Call.ByName(`${SERVICE}.StartScreenshots`)
      running.value = true
      scanMode.value = 'screenshots'
      startPolling()
    }
    catch (err) {
      log.error('useScanner: failed to start screenshot capture', { error: String(err) })
    }
    finally {
      scanning.value = false
    }
  }

  async function stopScan() {
    stopping.value = true
    stopPolling()
    try {
      await Call.ByName(`${SERVICE}.Stop`)
      running.value = false
      scanMode.value = ''
      // stopping remains true until EVENT_SCAN_COMPLETED / EVENT_SCAN_ERROR
      // arrives, so stale progress events cannot flip running back to true.
    }
    catch (err) {
      log.error('useScanner: failed to stop scan', { error: String(err) })
      // RPC error – completion event may never fire, so reset now.
      stopping.value = false
      running.value = false
      scanMode.value = ''
    }
  }

  async function listHosts(page: number, pageSize: number, sortBy = 'ping_ms', sortOrder = 'asc', search = '', countries: string[] = [], labels: string[] = [], hardware = ''): Promise<HostPage> {
    try {
      const result = await Call.ByName(`${SERVICE}.ListHosts`, page, pageSize, sortBy, sortOrder, search, countries, labels, hardware) as HostPage | null
      return result ?? { items: [], total: 0 }
    }
    catch (err) {
      log.error('useScanner: failed to list hosts', { error: String(err) })
      return { items: [], total: 0 }
    }
  }

  async function listVNC(page: number, pageSize: number, sortBy = 'ip', sortOrder = 'asc', search = '', countries: string[] = [], labels: string[] = [], hardware = '', authFilter: AuthFilter = 'all', screenshotFilter: ScreenshotFilter = 'all'): Promise<VNCPage> {
    try {
      const result = await Call.ByName(`${SERVICE}.ListVNCServices`, page, pageSize, sortBy, sortOrder, search, countries, labels, hardware, authFilter, screenshotFilter) as VNCPage | null
      return result ?? { items: [], total: 0 }
    }
    catch (err) {
      log.error('useScanner: failed to list VNC services', { error: String(err) })
      return { items: [], total: 0 }
    }
  }

  async function getFilterOptions(): Promise<FilterOptions> {
    try {
      const result = await Call.ByName(`${SERVICE}.GetFilterOptions`) as FilterOptions | null
      return result ?? { countries: [], labels: [] }
    }
    catch (err) {
      log.error('useScanner: failed to get filter options', { error: String(err) })
      return { countries: [], labels: [] }
    }
  }

  async function getBrowseStats(): Promise<BrowseStats> {
    try {
      const result = await Call.ByName(`${SERVICE}.GetBrowseStats`) as BrowseStats | null
      return result ?? { total_hosts: 0, ping_ok_hosts: 0, total_vnc: 0, no_auth_vnc: 0, screenshot_vnc: 0 }
    }
    catch (err) {
      log.error('useScanner: failed to get browse stats', { error: String(err) })
      return { total_hosts: 0, ping_ok_hosts: 0, total_vnc: 0, no_auth_vnc: 0, screenshot_vnc: 0 }
    }
  }

  function startPolling() {
    stopPolling()
    pollTimer = setInterval(async () => {
      if (!running.value) return
      try {
        const result = await Call.ByName(`${SERVICE}.GetProgress`) as ScanProgressData | null
        if (running.value) progress.value = result
      }
      catch { /* ignore */ }
    }, POLL_INTERVAL_MS)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  onMounted(() => {
    refresh()

    cleanups.push(
      Events.On(EVENT_SCAN_STARTED, (ev: { data: any }) => {
        running.value = true
        scanMode.value = typeof ev?.data === 'string' ? ev.data as ScanMode : 'hosts'
        startPolling()
      }) ?? (() => {}),
      Events.On(EVENT_SCAN_PROGRESS, (ev: { data: any }) => {
        if (ev?.data) {
          progress.value = ev.data
          if (!stopping.value) {
            running.value = ev.data.running ?? running.value
          }
          if (ev.data.mode) scanMode.value = ev.data.mode
        }
      }) ?? (() => {}),
      Events.On(EVENT_SCAN_COMPLETED, (ev: { data: any }) => {
        if (ev?.data) progress.value = ev.data
        running.value = false
        stopping.value = false
        scanMode.value = ''
        stopPolling()
        if (autoClearTimer) clearTimeout(autoClearTimer)
        autoClearTimer = setTimeout(() => { progress.value = null; autoClearTimer = null }, 5000)
      }) ?? (() => {}),
      Events.On(EVENT_SCAN_ERROR, () => {
        running.value = false
        stopping.value = false
        scanMode.value = ''
        stopPolling()
        if (autoClearTimer) clearTimeout(autoClearTimer)
        autoClearTimer = setTimeout(() => { progress.value = null; autoClearTimer = null }, 5000)
      }) ?? (() => {}),
    )
  })

  onUnmounted(() => {
    stopPolling()
    cleanups.forEach(fn => fn())
  })

  async function getHostByIP(ip: string): Promise<HostItem | null> {
    try {
      return await Call.ByName(`${SERVICE}.GetHostByIP`, ip) as HostItem | null
    }
    catch (err) {
      log.error('useScanner: failed to get host by IP', { error: String(err) })
      return null
    }
  }

  async function getVNCScreenshot(id: number): Promise<string> {
    try {
      return await Call.ByName(`${SERVICE}.GetVNCScreenshot`, id) as string ?? ''
    }
    catch (err) {
      log.error('useScanner: failed to get VNC screenshot', { error: String(err) })
      return ''
    }
  }

  async function refreshScreenshots(ids: number[]): Promise<ScreenshotResult[]> {
    if (!ids.length) return []
    try {
      const result = await Call.ByName(`${SERVICE}.RefreshScreenshots`, ids) as ScreenshotResult[] | null
      return result ?? []
    }
    catch (err) {
      log.error('useScanner: failed to refresh screenshots', { error: String(err) })
      return []
    }
  }

  function clearProgress() {
    if (autoClearTimer) {
      clearTimeout(autoClearTimer)
      autoClearTimer = null
    }
    stopPolling()
    progress.value = null
  }

  async function toggleHostFavorite(id: number): Promise<boolean> {
    try {
      return await Call.ByName(`${SERVICE}.ToggleHostFavorite`, id) as boolean
    }
    catch (err) {
      log.error('useScanner: failed to toggle host favorite', { error: String(err) })
      return false
    }
  }

  async function toggleVNCFavorite(id: number): Promise<boolean> {
    try {
      return await Call.ByName(`${SERVICE}.ToggleVNCFavorite`, id) as boolean
    }
    catch (err) {
      log.error('useScanner: failed to toggle VNC favorite', { error: String(err) })
      return false
    }
  }

  return {
    running: readonly(running),
    scanning: readonly(scanning),
    stopping: readonly(stopping),
    scanMode: readonly(scanMode),
    progress: readonly(progress),
    refresh,
    rescan,
    scanHosts,
    scanScreenshots,
    stopScan,
    clearProgress,
    listHosts,
    listVNC,
    getFilterOptions,
    getBrowseStats,
    getHostByIP,
    getVNCScreenshot,
    refreshScreenshots,
    toggleHostFavorite,
    toggleVNCFavorite,
  }
}
