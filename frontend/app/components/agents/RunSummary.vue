<!-- app/components/agents/RunSummary.vue -->
<template>
  <div v-if="summary">
    <template v-if="summary.mode === 'scrape'">
      <v-alert v-if="summary.accounts_exhausted" type="warning" variant="tonal" density="compact" class="mb-3">
        {{ t('agents.run.accountsExhausted') }}
      </v-alert>
      <v-row dense>
        <v-col class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold">{{ summary.completed_queries }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.completed') }}</span>
        </v-col>
        <v-col class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold text-warning">{{ summary.empty_queries }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.empty') }}</span>
        </v-col>
        <v-col class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold">{{ summary.total_queries }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.total') }}</span>
        </v-col>
        <v-col class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold text-success">{{ summary.total_hosts }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.hosts') }}</span>
        </v-col>
        <v-col class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold">{{ summary.used_credits }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.usedCredits') }}</span>
        </v-col>
      </v-row>
    </template>

    <template v-if="summary.mode === 'refresh'">
      <v-row dense>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold text-success">{{ summary.refreshed_accounts }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.refreshed') }}</span>
        </v-col>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold text-error">{{ summary.failed_accounts }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.failed') }}</span>
        </v-col>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold text-warning">{{ refreshRemaining }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.remaining') }}</span>
        </v-col>
      </v-row>
      <v-row dense class="mt-3">
        <v-col cols="3" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-2 font-weight-medium">{{ refreshSuccessRateText }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.successRate') }}</span>
        </v-col>
        <v-col cols="3" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-2 font-weight-medium">{{ formatDuration(summary.duration_secs) }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.elapsed') }}</span>
        </v-col>
        <v-col cols="3" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-2 font-weight-medium">{{ summary.total_credits }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.totalCredits') }}</span>
        </v-col>
        <v-col cols="3" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-2 font-weight-medium">{{ summary.possible_queries }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.possibleQueries') }}</span>
        </v-col>
      </v-row>
    </template>

    <template v-if="summary.mode === 'register'">
      <v-row dense>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold text-success">{{ summary.created_accounts }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.register.created') }}</span>
        </v-col>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold text-error">{{ summary.failed_registrations }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.failed') }}</span>
        </v-col>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-1 font-weight-bold">{{ summary.target_accounts }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.register.target') }}</span>
        </v-col>
      </v-row>
      <v-row dense class="mt-3">
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-2 font-weight-medium">{{ successRateText }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.successRate') }}</span>
        </v-col>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-2 font-weight-medium">{{ registerElapsedDisplay }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.elapsed') }}</span>
        </v-col>
        <v-col cols="4" class="d-flex flex-column align-center text-center">
          <span class="text-subtitle-2 font-weight-medium">{{ registerRemaining }}</span>
          <span class="text-overline text-medium-emphasis">{{ t('agents.run.remaining') }}</span>
        </v-col>
      </v-row>
    </template>
  </div>
</template>

<script setup lang="ts">
import type { RunSummary } from '~~bindings/smegg.me/thughunter/core/scraper/models.js'

const { t } = useI18n()
const props = defineProps<{ summary: RunSummary | null }>()

const refreshRemaining = computed(() => {
  const total = props.summary?.total_accounts ?? 0
  const done = (props.summary?.refreshed_accounts ?? 0) + (props.summary?.failed_accounts ?? 0)
  return Math.max(0, total - done)
})

const refreshProcessed = computed(() =>
  (props.summary?.refreshed_accounts ?? 0) + (props.summary?.failed_accounts ?? 0),
)

const refreshSuccessRateText = computed(() => {
  if (refreshProcessed.value === 0) return '-'
  return `${Math.round(((props.summary?.refreshed_accounts ?? 0) / refreshProcessed.value) * 100)}%`
})

const registerElapsedDisplay = computed(() => {
  const elapsed = formatDuration(props.summary?.duration_secs ?? 0)
  const max = props.summary?.max_duration_secs ?? 0
  if (max > 0) return `${elapsed} / ${formatDuration(max)}`
  return elapsed
})

const registerRemaining = computed(() => {
  const target = props.summary?.target_accounts ?? 0
  if (target <= 0) return '-'
  return Math.max(0, target - (props.summary?.created_accounts ?? 0))
})

const totalRegister = computed(() =>
  (props.summary?.created_accounts ?? 0) + (props.summary?.failed_registrations ?? 0),
)

const successRateText = computed(() => {
  if (totalRegister.value === 0) return '-'
  return `${Math.round(((props.summary?.created_accounts ?? 0) / totalRegister.value) * 100)}%`
})
</script>
