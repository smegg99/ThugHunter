// app/composables/useLogger.ts
import * as LoggerService from '~~bindings/smegg.me/thughunter/services/logger/service.js'
import type { LogLevel, Fields } from '~/types/logger'

const levelToBackend: Record<LogLevel, (msg: string, fields: Fields) => Promise<void>> = {
  trace: LoggerService.LogTrace,
  debug: LoggerService.LogDebug,
  info: LoggerService.LogInfo,
  warn: LoggerService.LogWarn,
  error: LoggerService.LogError,
}

const levelToConsole: Record<LogLevel, (...args: unknown[]) => void> = {
  trace: console.debug,
  debug: console.debug,
  info: console.info,
  warn: console.warn,
  error: console.error,
}

let consoleLogEnabled = true

async function loadConfig() {
  try {
    const cfg = await LoggerService.GetFrontendLoggerConfig()
    consoleLogEnabled = cfg.console_log
  }
  catch {
    consoleLogEnabled = true
  }
}

loadConfig()

function log(level: LogLevel, msg: string, fields: Fields = {}) {
  if (consoleLogEnabled) {
    const hasFields = Object.keys(fields).length > 0
    levelToConsole[level](
      `[frontend] ${msg}`,
      ...(hasFields ? [fields] : []),
    )
  }

  levelToBackend[level](msg, fields).catch(() => {})
}

export function useLogger() {
  return {
    trace: (msg: string, fields?: Fields) => log('trace', msg, fields ?? {}),
    debug: (msg: string, fields?: Fields) => log('debug', msg, fields ?? {}),
    info: (msg: string, fields?: Fields) => log('info', msg, fields ?? {}),
    warn: (msg: string, fields?: Fields) => log('warn', msg, fields ?? {}),
    error: (msg: string, fields?: Fields) => log('error', msg, fields ?? {}),
    refreshConfig: loadConfig,
  }
}
