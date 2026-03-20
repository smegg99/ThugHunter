// app/utils/wailsError.ts

/**
 * Wails v3 alpha wraps binding errors as: new Error(JSON.stringify({message, cause, kind})).
 * This extracts the inner message from such wrapped errors.
 */
export function extractWailsError(err: unknown): string {
  let msg = ''
  if (err instanceof Error) msg = err.message
  else if (typeof err === 'string') msg = err
  else if (err && typeof err === 'object' && 'message' in err) msg = String((err as any).message)
  else msg = String(err)

  try {
    const parsed = JSON.parse(msg)
    if (typeof parsed?.message === 'string') return parsed.message
  }
  catch { }
  return msg
}
