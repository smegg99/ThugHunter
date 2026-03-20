// app/types/index.ts
import type { Account } from '~~bindings/smegg.me/thughunter/core/models/models.js'

export type AccountRow = Account & { status: string }
export type SortItem = { key: string; order: 'asc' | 'desc' }
export interface AccountPage { items: Account[]; total: number }
export interface SecondaryNavItem { to: string; icon: string; title: string }

export type {
  ScreenshotStage,
  VNCAuthType,
  TrayPhase,
  AuthFilter,
  ScreenshotFilter,
  ScanMode,
  ScanProgressData,
  HostItem,
  VNCItem,
  ScreenshotResult,
  HostPage,
  VNCPage,
  FilterOptions,
  SortFilterParams,
  BrowseStats,
} from './scanner'
