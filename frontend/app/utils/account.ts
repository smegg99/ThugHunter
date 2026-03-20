// app/utils/account.ts
import type { Account } from '~~bindings/smegg.me/thughunter/core/models/models.js'

export function creditsExpired(account: Account): boolean {
  if (!account.credits_expire_at) return true
  return new Date() > new Date(account.credits_expire_at)
}

export function accountStatus(account: Account): string {
  if (creditsExpired(account)) return 'expired'
  if (account.credits_amount > 0) return 'active'
  return 'noCredits'
}
