// app/composables/useScraper.ts
import { Call, Events } from '@wailsio/runtime'
import * as ScraperService from '~~bindings/smegg.me/thughunter/services/scraper/service.js'
import * as ConfigService from '~~bindings/smegg.me/thughunter/services/config/service.js'
import type { RegisterOpts, RunProgress, RunSummary } from '~~bindings/smegg.me/thughunter/core/scraper/models.js'
import type { Account } from '~~bindings/smegg.me/thughunter/core/models/models.js'
import type { AccountPage } from '~/types'
import type { TrayPhase } from '~/types/scanner'

const SERVICE = 'smegg.me/thughunter/services/scraper.Service'

const randomDelay = (min = 400, max = 900) =>
  new Promise<void>(r => setTimeout(r, min + Math.random() * (max - min)))

const EVENT_RUN_STATE_CHANGED = 'scraper:service:run_state_changed'
const EVENT_PROGRESS = 'scraper:progress'
const EVENT_ACCOUNTS_CHANGED = 'scraper:service:accounts_changed'
const EVENT_RUN_SUMMARY = 'scraper:run_summary'
const EVENT_TRAY_STATE = 'tray:state_changed'

const agents = ref<RunProgress['agents']>([])
const accounts = ref<Account[]>([])
const running = ref(false)
const starting = ref(false)
const stopping = ref(false)
const progress = ref<RunProgress | null>(null)
const summary = ref<RunSummary | null>(null)
// Mirrors the Go tray.Manager state: 0=idle, 1=starting, 2=running, 3=stopping
const trayPhase = ref<TrayPhase>(0)
const imapReady = ref(false)
const log = useLogger()

export function useScraper() {
  const cleanups: (() => void)[] = []
  let pollTimer: ReturnType<typeof setInterval> | null = null

  async function refresh() {
    try {
      const [isRunning, prog, accountList] = await Promise.all([
        ScraperService.IsRunning(),
        ScraperService.GetProgress(),
        ScraperService.ListAccounts(),
      ])
      running.value = isRunning
      progress.value = prog
      agents.value = prog?.agents ?? []
      accounts.value = accountList ?? []
      if (!isRunning) summary.value = null
      if (running.value) startPolling()
    }
    catch (err) {
      log.error('failed to refresh scraper state', { error: String(err) })
    }
  }

  async function start() {
    starting.value = true
    try {
      await Promise.all([ScraperService.Start(), randomDelay()])
      running.value = true
      startPolling()
    }
    catch (err) {
      log.error('failed to start scraper', { error: String(err) })
    }
    finally {
      starting.value = false
    }
  }

  async function stop() {
    stopping.value = true
    try {
      await Promise.all([ScraperService.Stop(), randomDelay()])
      running.value = false
      agents.value = []
    }
    catch (err) {
      log.error('failed to stop scraper', { error: String(err) })
    }
    finally {
      stopping.value = false
    }
  }

  async function refreshAccounts() {
    starting.value = true
    try {
      await Promise.all([ScraperService.RefreshAccounts(), randomDelay()])
      running.value = true
      startPolling()
    }
    catch (err) {
      log.error('failed to start account refresh', { error: String(err) })
    }
    finally {
      starting.value = false
    }
  }

  async function checkImapReady() {
    try {
      await ConfigService.VerifyImapConnection()
      imapReady.value = true
    }
    catch {
      imapReady.value = false
    }
  }

  async function registerAccounts(opts: RegisterOpts) {
    starting.value = true
    try {
      await Promise.all([ScraperService.RegisterAccounts(opts), randomDelay()])
      running.value = true
      startPolling()
    }
    catch (err) {
      log.error('failed to start registration run', { error: String(err) })
    }
    finally {
      starting.value = false
    }
  }

  async function loadAccountPage(page: number, pageSize: number, sortBy: string, sortOrder: string, search: string = ''): Promise<AccountPage> {
    try {
      const result = await Call.ByName(`${SERVICE}.ListAccountsPaginated`, page, pageSize, sortBy, sortOrder, search)
      return { items: result?.items ?? [], total: result?.total ?? 0 }
    }
    catch (err) {
      log.error('failed to load account page', { error: String(err) })
      return { items: [], total: 0 }
    }
  }

  async function deleteAccount(id: number): Promise<void> {
    await Call.ByName(`${SERVICE}.DeleteAccount`, id)
  }

  async function updateAccountPassword(id: number, password: string): Promise<void> {
    await Call.ByName(`${SERVICE}.UpdateAccountPassword`, id, password)
  }

  async function addAccount(email: string, password: string): Promise<void> {
    await Call.ByName(`${SERVICE}.AddAccount`, email, password)
  }

  async function canStartRun(): Promise<boolean> {
    try {
      return await Call.ByName(`${SERVICE}.CanStartRun`)
    }
    catch (err) {
      log.error('failed to check if can start run', { error: String(err) })
      return false
    }
  }

  function startPolling() {
    stopPolling()
    pollTimer = setInterval(async () => {
      if (!running.value) return
      const prog = await ScraperService.GetProgress()
      if (prog) {
        progress.value = prog
        agents.value = prog.agents ?? []
      }
    }, 2000)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  onMounted(() => {
    refresh()
    checkImapReady()

    cleanups.push(
      Events.On(EVENT_RUN_STATE_CHANGED, (ev: { data: any }) => {
        running.value = !!ev.data
        if (running.value) {
          startPolling()
        }
        else {
          stopping.value = false
          stopPolling()
          agents.value = []
        }
      }) ?? (() => {}),
      Events.On(EVENT_PROGRESS, (ev: { data: any }) => {
        if (ev.data) {
          progress.value = ev.data
          agents.value = ev.data.agents ?? []
        }
      }) ?? (() => {}),
      Events.On(EVENT_ACCOUNTS_CHANGED, () => {
        ScraperService.ListAccounts().then((list) => {
          accounts.value = list ?? []
        })
      }) ?? (() => {}),
      Events.On(EVENT_RUN_SUMMARY, (ev: { data: any }) => {
        if (ev.data) {
          summary.value = ev.data
        }
      }) ?? (() => {}),
      Events.On(EVENT_TRAY_STATE, (ev: { data: any }) => {
        trayPhase.value = typeof ev.data === 'number' ? ev.data as TrayPhase : 0
      }) ?? (() => {}),
    )
  })

  onUnmounted(() => {
    stopPolling()
    cleanups.forEach(fn => fn())
  })

  return {
    agents,
    accounts,
    running,
    starting,
    stopping,
    progress,
    summary,
    trayPhase,
    imapReady,
    start,
    stop,
    refresh,
    refreshAccounts,
    registerAccounts,
    loadAccountPage,
    deleteAccount,
    updateAccountPassword,
    addAccount,
    canStartRun,
  }
}
