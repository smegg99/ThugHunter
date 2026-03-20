<!-- app/components/agents/RegisterProgress.vue -->
<template>
  <div v-if="progress && progress.mode === 'register'">
    <v-row dense>
      <v-col :cols="hasTarget ? 4 : 6" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold text-success">{{ progress.created_accounts }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.register.created') }}</span>
      </v-col>
      <v-col :cols="hasTarget ? 4 : 6" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold text-error">{{ progress.failed_registrations }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.failed') }}</span>
      </v-col>
      <v-col v-if="hasTarget" cols="4" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold">{{ progress.target_accounts }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.register.target') }}</span>
      </v-col>
    </v-row>

    <v-progress-linear v-if="progress.target_accounts > 0" :model-value="progressPercent" :color="progressColor"
      height="4" rounded class="mt-3" />
    <v-progress-linear v-else indeterminate :color="progressColor" height="4" rounded class="mt-3" />

    <v-row dense class="mt-3">
      <v-col :cols="hasDuration ? 4 : 3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ progress.active_agents }} / {{ progress.total_agents
          }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.agents') }}</span>
      </v-col>
      <v-col :cols="hasDuration ? 4 : 3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ successRateText }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.successRate') }}</span>
      </v-col>
      <v-col :cols="hasDuration ? 4 : 3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ timeDisplay }}</span>
        <span class="text-overline text-medium-emphasis">{{ timeLabel }}</span>
      </v-col>
      <v-col v-if="!hasDuration" cols="3" class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-2 font-weight-medium">{{ etaDisplay }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.eta') }}</span>
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

const total = computed(() =>
  (props.progress?.created_accounts ?? 0) + (props.progress?.failed_registrations ?? 0),
)

const progressPercent = computed(() => {
  const target = props.progress?.target_accounts ?? 0
  return target > 0 ? ((props.progress?.created_accounts ?? 0) / target) * 100 : 0
})

const successRate = computed(() => {
  const t = total.value
  if (t === 0) return -1
  return (props.progress?.created_accounts ?? 0) / t
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

const elapsed = computed(() => formatDuration(elapsedSeconds.value))
const hasDuration = computed(() => (props.progress?.duration_secs ?? 0) > 0)
const hasTarget = computed(() => (props.progress?.target_accounts ?? 0) > 0)

const remainingSeconds = computed(() => {
  const dur = props.progress?.duration_secs ?? 0
  if (dur <= 0) return 0
  return Math.max(0, dur - elapsedSeconds.value)
})

const timeDisplay = computed(() =>
  props.stopping ? '-' : hasDuration.value ? formatDuration(remainingSeconds.value) : elapsed.value,
)

const timeLabel = computed(() =>
  hasDuration.value ? t('agents.run.remaining') : t('agents.run.elapsed'),
)

const etaDisplay = computed(() => {
  if (props.stopping) return '-'
  if (hasDuration.value) return formatDuration(remainingSeconds.value)
  const target = props.progress?.target_accounts ?? 0
  const done = total.value
  if (target <= 0 || done <= 0) return '-'
  const rate = done / elapsedSeconds.value
  if (rate <= 0) return '-'
  return formatDuration(Math.max(0, target - (props.progress?.created_accounts ?? 0)) / rate)
})
</script>
