<!-- app/pages/browse.vue -->
<template>
  <div>
  <!-- Search bar centered in the app bar -->
  <Teleport to="#app-bar-search-slot">
    <LayoutAppBarSearch v-model="search" :placeholder="searchPlaceholder" />
  </Teleport>

  <!-- Action buttons + live stats to the right of the search bar -->
  <Teleport to="#app-bar-actions">
    <Transition name="filter-btn">
      <v-btn icon density="comfortable" variant="text" :title="t('browse.sortFilter.title')" :disabled="running"
        @click="showSortFilter = true">
        <v-badge v-if="hasActiveFilters" dot color="primary">
          <v-icon icon="mdi-sort-variant" />
        </v-badge>
        <v-icon v-else icon="mdi-sort-variant" />
      </v-btn>
    </Transition>

    <!-- Per-tab scan/screenshot button -->
    <v-btn v-if="activeTab === 'all'" icon density="comfortable"
      :color="running && scanMode === 'hosts' ? 'error' : 'primary'" variant="text"
      :loading="scanning || stopping" :disabled="(running && scanMode !== 'hosts') || !stats?.total_hosts"
      :title="running && scanMode === 'hosts' ? t('browse.stopScan') : t('browse.scanHosts')"
      @click="onHostScanClick">
      <v-icon :icon="running && scanMode === 'hosts' ? 'mdi-stop' : 'mdi-magnify-scan'" />
    </v-btn>
    <v-btn v-else-if="activeTab === 'vnc'" icon density="comfortable"
      :color="running && scanMode === 'screenshots' ? 'error' : 'secondary'" variant="text"
      :loading="scanning || stopping" :disabled="(running && scanMode !== 'screenshots') || !stats?.total_vnc"
      :title="running && scanMode === 'screenshots' ? t('browse.stopScan') : t('browse.captureScreenshots')"
      @click="onScreenshotClick">
      <v-icon :icon="running && scanMode === 'screenshots' ? 'mdi-stop' : 'mdi-camera'" />
    </v-btn>

    <!-- Inline progress (compact) -->
    <BrowseScanProgress v-if="running && scanMode === 'hosts'" :progress="progress" :running="running" compact />
    <BrowseScanProgress v-else-if="running && scanMode === 'screenshots'" :progress="progress" :running="running" compact />

    <!-- Stats when idle -->
    <div v-else-if="stats" class="d-flex align-center text-caption text-medium-emphasis" style="gap: 16px; white-space: nowrap; font-family: monospace">
      <template v-if="activeTab === 'all'">
        <span class="d-inline-flex align-center" style="min-width: 5ch"><v-icon icon="mdi-server" size="16" class="mr-1" />{{ stats.total_hosts }}</span>
        <span class="d-inline-flex align-center" style="min-width: 5ch"><v-icon icon="mdi-table-tennis" size="16" class="mr-1" />{{ stats.ping_ok_hosts
          }}</span>
      </template>
      <template v-else-if="activeTab === 'vnc'">
        <span class="d-inline-flex align-center" style="min-width: 5ch"><v-icon icon="mdi-monitor" size="16" class="mr-1" />{{ stats.total_vnc }}</span>
        <span class="d-inline-flex align-center" style="min-width: 5ch"><v-icon icon="mdi-lock-open-variant" size="16" class="mr-1" />{{ stats.no_auth_vnc
          }}</span>
        <span class="d-inline-flex align-center" style="min-width: 5ch"><v-icon icon="mdi-camera" size="16" class="mr-1" />{{ stats.screenshot_vnc }}</span>
      </template>
    </div>
  </Teleport>

  <v-container fluid class="pa-0" style="width: 100%">
    <!-- Tabs -->
    <v-tabs v-model="activeTab" color="primary" class="browse-tabs">
      <v-tab v-for="tab in tabs" :key="tab.value" :value="tab.value" :disabled="tab.disabled">
        {{ tab.label }}
      </v-tab>
    </v-tabs>
    <v-divider />

    <div class="pa-4">
      <!-- Tab content -->
      <v-window v-model="activeTab" :touch="false" :transition="false" :reverse-transition="false">
        <!-- All Hosts -->
        <v-window-item value="all" eager>
          <BrowseHostList ref="hostListRef" :search="search" :sort-filter="hostSortFilter" :saved-page="savedPages.all"
            :frozen="running && scanMode === 'hosts'" @update:page="p => savePage('all', p)" />
        </v-window-item>

        <!-- VNC Services -->
        <v-window-item value="vnc" eager>
          <BrowseVNCList ref="vncListRef" :search="search" :sort-filter="vncSortFilter" :saved-page="savedPages.vnc"
            :frozen="running && scanMode === 'screenshots'" @update:page="p => savePage('vnc', p)" />
        </v-window-item>

        <!-- WIP tabs: empty for now -->
        <v-window-item v-for="wip in wipTabs" :key="wip" :value="wip" eager />
      </v-window>
    </div>
  </v-container>

  <!-- Sort / Filter dialog -->
  <BrowseSortFilterDialog v-model="showSortFilter" :current="activeSortFilter" :sort-fields="activeSortFields"
    :filter-options="filterOptions" :show-no-auth-only="activeTab === 'vnc'" @apply="onApplySort" />
  </div>
</template>

<script setup lang="ts">
import type { SortFilterParams, FilterOptions, BrowseStats } from '~/types/scanner'

const { t } = useI18n()
const { running, scanning, stopping, scanMode, progress, scanHosts, scanScreenshots, stopScan, clearProgress, getFilterOptions, getBrowseStats } = useScanner()

const activeTab = ref('all')
const search = ref('')
const showSortFilter = ref(false)
const stats = ref<BrowseStats | null>(null)

const hostListRef = ref<{ reload: () => void } | null>(null)
const vncListRef = ref<{ reload: () => void } | null>(null)

const filterOptions = ref<FilterOptions>({ countries: [], labels: [] })

const PAGES_STORAGE_KEY = 'thughunter:browse:pages'

// Saved pages per tab, persisted in localStorage
const savedPages = reactive<Record<string, number>>(loadSavedPages())

function loadSavedPages(): Record<string, number> {
  try {
    const raw = localStorage.getItem(PAGES_STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  }
  catch { /* ignore */ }
  return { all: 1, vnc: 1 }
}

function savePage(tab: string, page: number) {
  savedPages[tab] = page
  localStorage.setItem(PAGES_STORAGE_KEY, JSON.stringify(savedPages))
}

// Dynamic search placeholder based on active tab
const searchPlaceholder = computed(() => {
  if (activeTab.value === 'vnc') return t('browse.searchVNC')
  return t('browse.search')
})

// Per-tab sort/filter state
const hostSortFilter = reactive<SortFilterParams>({
  sortBy: 'ping_ms',
  sortOrder: 'asc',
  countries: [],
  labels: [],
  hardware: '',
  pageSize: 24,
  authFilter: 'all',
  screenshotFilter: 'all',
})

const vncSortFilter = reactive<SortFilterParams>({
  sortBy: 'latency_ms',
  sortOrder: 'asc',
  countries: [],
  labels: [],
  hardware: '',
  pageSize: 24,
  authFilter: 'all',
  screenshotFilter: 'has',
})

// Sort fields per tab
const hostSortFields = computed(() => [
  { value: 'ping_ms', label: t('browse.sortFilter.fields.ping_ms') },
  { value: 'ip', label: t('browse.sortFilter.fields.ip') },
  { value: 'city', label: t('browse.sortFilter.fields.city') },
  { value: 'country_code', label: t('browse.sortFilter.fields.country_code') },
  { value: 'created_at', label: t('browse.sortFilter.fields.created_at') },
])

const vncSortFields = computed(() => [
  { value: 'latency_ms', label: t('browse.sortFilter.fields.ping_ms') },
  { value: 'ip', label: t('browse.sortFilter.fields.ip') },
  { value: 'port', label: t('browse.sortFilter.fields.port') },
])

// Expose the correct sort fields / state for the active tab
const activeSortFilter = computed<SortFilterParams>(() =>
  activeTab.value === 'vnc' ? vncSortFilter : hostSortFilter,
)

const activeSortFields = computed(() =>
  activeTab.value === 'vnc' ? vncSortFields.value : hostSortFields.value,
)

const hasActiveFilters = computed(() => {
  const sf = activeSortFilter.value
  return sf.countries.length > 0 || sf.labels.length > 0 || sf.hardware !== '' || sf.authFilter !== 'all' || sf.screenshotFilter !== 'all'
})

function onApplySort(params: SortFilterParams) {
  const target = activeTab.value === 'vnc' ? vncSortFilter : hostSortFilter
  Object.assign(target, params)
}

const tabs = computed(() => [
  { value: 'all', label: t('browse.tabs.all') },
  { value: 'vnc', label: t('browse.tabs.vnc') },
  { value: 'spice', label: t('browse.tabs.spice'), disabled: true },
  { value: 'rdp', label: t('browse.tabs.rdp'), disabled: true },
  { value: 'modprobe', label: t('browse.tabs.modprobe'), disabled: true },
  { value: 'pkcam', label: t('browse.tabs.pkcam'), disabled: true },
])

const wipTabs = ['spice', 'rdp', 'modprobe', 'pkcam']

async function onHostScanClick() {
  if (running.value && scanMode.value === 'hosts') {
    await stopScan()
  }
  else {
    await scanHosts()
  }
}

async function onScreenshotClick() {
  if (running.value && scanMode.value === 'screenshots') {
    await stopScan()
  }
  else {
    await scanScreenshots()
  }
}

// Reload when switching tabs so data is always fresh.
// Also clear the scan summary when the user switches tabs.
watch(activeTab, (tab) => {
  if (!running.value && progress.value) {
    clearProgress()
  }
  if (!running.value) {
    nextTick(() => {
      if (tab === 'all') hostListRef.value?.reload()
      else if (tab === 'vnc') vncListRef.value?.reload()
    })
  }
  scrollToTop()
})

// Reload the active list when a scan completes.
watch(running, (isRunning, wasRunning) => {
  if (!isRunning && wasRunning) {
    // Reset sort/filter to defaults and go to page 1
    Object.assign(hostSortFilter, { sortBy: 'ping_ms', sortOrder: 'asc', countries: [], labels: [], hardware: '', pageSize: 24, authFilter: 'all', screenshotFilter: 'all' })
    Object.assign(vncSortFilter, { sortBy: 'latency_ms', sortOrder: 'asc', countries: [], labels: [], hardware: '', pageSize: 24, authFilter: 'all', screenshotFilter: 'has' })
    search.value = ''
    savedPages.all = 1
    savedPages.vnc = 1
    localStorage.setItem(PAGES_STORAGE_KEY, JSON.stringify(savedPages))

    hostListRef.value?.reload()
    if (activeTab.value === 'vnc') vncListRef.value?.reload()
    // Refresh filter options and stats after scan - new data may appear
    loadFilterOptions()
    loadStats()
  }
})

async function loadFilterOptions() {
  filterOptions.value = await getFilterOptions()
}

async function loadStats() {
  stats.value = await getBrowseStats()
}

onMounted(() => {
  loadFilterOptions()
  loadStats()
})
</script>

<style scoped>
/* Sort/filter button enter/leave transition */
.filter-btn-enter-active,
.filter-btn-leave-active {
  transition: opacity 0.25s ease, transform 0.25s ease;
}

.filter-btn-enter-from,
.filter-btn-leave-to {
  opacity: 0;
  transform: scale(0.7);
}
</style>
