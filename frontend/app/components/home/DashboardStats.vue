<!-- app/components/home/DashboardStats.vue -->
<template>
  <div class="pa-3 ma-4">
    <div class="d-flex align-center justify-center ga-4 flex-wrap-nowrap">
      <div class="d-flex align-center ga-2 flex-nowrap">
        <v-icon icon="mdi-server-network" size="22" />
        <span class="text-subtitle-1 font-weight-bold" style="font-family: monospace">{{
          stats.totalHosts.toLocaleString()
          }}</span>
        <span class="text-subtitle-2 text-medium-emphasis text-no-wrap" style="font-family: monospace">{{
          t('home.stats.hosts') }}</span>
      </div>
      <v-divider vertical class="my-1" />
      <div class="d-flex align-center ga-2 flex-nowrap">
        <v-icon icon="mdi-credit-card-outline" size="22" />
        <span class="text-subtitle-1 font-weight-bold" style="font-family: monospace">{{
          stats.totalCredits.toLocaleString()
          }}</span>
        <span class="text-subtitle-2 text-medium-emphasis text-no-wrap" style="font-family: monospace">{{
          t('home.stats.credits') }}</span>
      </div>
      <v-divider vertical class="my-1" />
      <div class="d-flex align-center ga-2 flex-nowrap">
        <v-icon icon="mdi-magnify" size="22" />
        <span class="text-subtitle-1 font-weight-bold" style="font-family: monospace">{{
          possibleQueries.toLocaleString()
          }}</span>
        <span class="text-subtitle-2 text-medium-emphasis text-no-wrap" style="font-family: monospace">{{
          t('home.stats.queries') }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Call, Events } from '@wailsio/runtime'

const SERVICE = 'smegg.me/thughunter/services/scraper.Service'
const { t } = useI18n()
const log = useLogger()

const stats = reactive({
  totalHosts: 0,
  totalAccounts: 0,
  totalCredits: 0,
  usableAccounts: 0,
})

const CREDITS_PER_QUERY = 5

const possibleQueries = computed(() => Math.floor(stats.totalCredits / CREDITS_PER_QUERY))

async function loadStats() {
  try {
    const result = await Call.ByName(`${SERVICE}.GetDashboardStats`)
    if (result) {
      stats.totalHosts = result.total_hosts ?? 0
      stats.totalAccounts = result.total_accounts ?? 0
      stats.totalCredits = result.total_credits ?? 0
      stats.usableAccounts = result.usable_accounts ?? 0
    }
  }
  catch (err) {
    log.error('failed to load dashboard stats', { error: String(err) })
  }
}

const cleanups: (() => void)[] = []

onMounted(() => {
  loadStats()
  cleanups.push(
    Events.On('scraper:service:accounts_changed', loadStats) ?? (() => { }),
    Events.On('scraper:run_completed', loadStats) ?? (() => { }),
  )
})

onUnmounted(() => {
  cleanups.forEach(fn => fn())
})
</script>
