<!-- app/components/agents/RefreshProgress.vue -->
<template>
  <div v-if="progress && progress.mode === 'refresh'">
    <v-row dense>
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold text-success">{{ progress.refreshed_accounts }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.refreshed') }}</span>
      </v-col>
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold text-error">{{ progress.failed_accounts }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.failed') }}</span>
      </v-col>
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold text-warning">{{ progress.remaining_accounts }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.remaining') }}</span>
      </v-col>
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold">{{ progress.total_accounts }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.totalAccounts') }}</span>
      </v-col>
    </v-row>

    <v-progress-linear v-if="progress.total_accounts > 0" :model-value="progressPercent" :color="progressColor"
      height="4" rounded class="mt-3" />
    <v-progress-linear v-else indeterminate color="primary" height="4" rounded class="mt-3" />

    <v-row dense class="mt-3">
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ progress.active_agents }} / {{ progress.total_agents
        }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.agents') }}</span>
      </v-col>
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ successRateText }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.successRate') }}</span>
      </v-col>
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ elapsed }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.elapsed') }}</span>
      </v-col>
      <v-col cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ eta }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.eta') }}</span>
      </v-col>
    </v-row>

    <v-row dense class="mt-3">
      <v-col cols="6" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ progress.total_credits }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.totalCredits') }}</span>
      </v-col>
      <v-col cols="6" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ progress.possible_queries }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.possibleQueries') }}</span>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import type { RunProgress } from '~~bindings/smegg.me/thughunter/core/scraper/models.js'

const { t } = useI18n()
const props = defineProps<{ progress: RunProgress | null, stopping?: boolean }>()

const now = ref(Date.now())
let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  timer = setInterval(() => {
    if (!props.stopping) now.value = Date.now()
  }, 1000)
})
onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const processed = computed(() =>
  (props.progress?.refreshed_accounts ?? 0) + (props.progress?.failed_accounts ?? 0),
)

const progressPercent = computed(() => {
  const total = props.progress?.total_accounts ?? 0
  return total > 0 ? (processed.value / total) * 100 : 0
})

const successRate = computed(() => {
  const p = processed.value
  if (p === 0) return -1
  return (props.progress?.refreshed_accounts ?? 0) / p
})

const successRateText = computed(() =>
  successRate.value < 0 ? '-' : `${Math.round(successRate.value * 100)}%`,
)

const progressColor = computed(() => {
  const rate = successRate.value
  if (rate < 0) return 'primary'
  if (rate >= 0.8) return 'success'
  if (rate >= 0.5) return 'warning'
  return 'error'
})

const elapsedSeconds = computed(() => {
  const started = props.progress?.started_at
  if (!started) return 0
  return Math.max(0, (now.value - new Date(started).getTime()) / 1000)
})

const elapsed = computed(() => props.stopping ? '-' : formatDuration(elapsedSeconds.value))

const eta = computed(() => {
  if (props.stopping) return '-'
  const done = processed.value
  const remaining = props.progress?.remaining_accounts ?? 0
  const secs = elapsedSeconds.value
  if (done === 0 || remaining === 0 || secs === 0) return '-'
  return '~' + formatDuration((secs / done) * remaining)
})
</script>
