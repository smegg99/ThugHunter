// app/composables/useImapVerify.ts
import * as ConfigService from '~~bindings/smegg.me/thughunter/services/config/service.js'
import { extractWailsError } from '~/utils/wailsError'

export function useImapVerify(config: ReturnType<typeof useConfigSync>['config']) {
  const { t, te } = useI18n()

  const verifying = ref(false)
  const verified = ref(false)
  const verifyError = ref('')

  let verifyGeneration = 0
  let verifyTimer: ReturnType<typeof setTimeout> | null = null
  let lastVerifiedSnapshot = ''

  function localizeImapError(raw: string): string {
    const code = raw.trim()
    const key = `settings.imap.errors.${code}`
    if (te(key)) return t(key)
    return raw
  }

  function imapSnapshot(): string {
    return JSON.stringify({
      h: config.imap.host,
      p: config.imap.port,
      u: config.imap.catch_all_username,
      pw: config.imap.catch_all_password,
      m: config.imap.mbox,
      t: config.imap.use_tls,
    })
  }

  const randomDelay = (min = 500, max = 1000) =>
    new Promise(r => setTimeout(r, Math.floor(Math.random() * (max - min + 1)) + min))

  function scheduleVerify(delay = 300) {
    if (verifyTimer) clearTimeout(verifyTimer)
    verifyTimer = setTimeout(() => verifyConnection(), delay)
  }

  async function verifyConnection() {
    const gen = ++verifyGeneration
    const snap = imapSnapshot()
    verifying.value = true
    verified.value = false
    verifyError.value = ''
    try {
      await Promise.all([ConfigService.VerifyImapConnection(), randomDelay()])
      if (gen !== verifyGeneration) return
      verified.value = true
      lastVerifiedSnapshot = snap
    }
    catch (err: any) {
      if (gen !== verifyGeneration) return
      verifyError.value = localizeImapError(extractWailsError(err))
      lastVerifiedSnapshot = ''
    }
    finally {
      if (gen === verifyGeneration) verifying.value = false
    }
  }

  function onSave() {
    const snap = imapSnapshot()
    if (snap === lastVerifiedSnapshot) return
    scheduleVerify()
  }

  return {
    verifying,
    verified,
    verifyError,
    verifyConnection,
    onSave,
  }
}
