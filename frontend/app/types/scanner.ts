// app/types/scanner.ts

/** Screenshot pipeline stage: 0 = not started, 1 = running, 2 = done. */
export type ScreenshotStage = 0 | 1 | 2

/** VNC auth type from Go backend: -1 = unknown, 1 = none, 2 = password. */
export type VNCAuthType = -1 | 1 | 2

/** Tray manager phase: 0 = idle, 1 = starting, 2 = running, 3 = stopping. */
export type TrayPhase = 0 | 1 | 2 | 3

export type AuthFilter = 'all' | 'open' | 'closed'
export type ScreenshotFilter = 'all' | 'has' | 'none'

/** Active scan mode: hosts = ping+probe, screenshots = capture only. */
export type ScanMode = 'hosts' | 'screenshots'

export interface ScanProgressData {
  running: boolean
  mode: ScanMode | ''
  total_hosts: number
  scanned: number
  ping_ok: number
  probe_ok: number
  saved: number
  elapsed_secs: number
  screenshot_stage: ScreenshotStage
  screenshot_total: number
  screenshot_done: number
}

export interface HostItem {
  ID: number
  UpdatedAt: string
  ip: string
  city: string
  region: string
  country_code: string
  os: string
  hardware: string
  labels: string[]
  services: Record<string, string[]>
  software: string[]
  ping_ms: number
  is_favorite: boolean
}

export interface VNCItem {
  ID: number
  UpdatedAt: string
  ip: string
  port: number
  host_id: number
  latency_ms: number
  rfb_version: string
  auth_type: VNCAuthType
  no_auth: boolean
  is_favorite: boolean
  country_code: string
  city: string
  os: string
  hardware: string
  has_screenshot: boolean
  screenshot_at: string | null
  stale_screenshot: boolean
}

export interface ScreenshotResult {
  id: number
  screenshot: string
}

export interface HostPage {
  items: HostItem[]
  total: number
}

export interface VNCPage {
  items: VNCItem[]
  total: number
}

export interface FilterOptions {
  countries: string[]
  labels: string[]
}

export interface SortFilterParams {
  sortBy: string
  sortOrder: 'asc' | 'desc'
  countries: string[]
  labels: string[]
  hardware: string
  pageSize: number
  authFilter: AuthFilter
  screenshotFilter: ScreenshotFilter
}

export interface BrowseStats {
  total_hosts: number
  ping_ok_hosts: number
  total_vnc: number
  no_auth_vnc: number
  screenshot_vnc: number
}
