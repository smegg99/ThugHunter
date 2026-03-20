<!-- app/components/agents/ScrapeProgress.vue -->
<template>
  <div v-if="progress && progress.mode === 'scrape'">
    <v-alert v-if="progress.accounts_exhausted" type="warning" variant="tonal" density="compact" class="mb-3">
      {{ t('agents.run.accountsExhausted') }}
    </v-alert>
    <v-row dense>
      <v-col class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold">{{ progress.completed_queries }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.completed') }}</span>
      </v-col>
      <v-col class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold text-warning">{{ progress.empty_queries }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.empty') }}</span>
      </v-col>
      <v-col class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold">{{ progress.total_queries }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.total') }}</span>
      </v-col>
      <v-col class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold">{{ remainingQueries }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.remaining') }}</span>
      </v-col>
      <v-col class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold text-success">{{ progress.total_hosts }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.hosts') }}</span>
      </v-col>
      <v-col class="d-flex flex-column align-center text-center">
        <span class="text-subtitle-1 font-weight-bold">{{ progress.used_credits }}</span>
        <span class="text-overline text-medium-emphasis">{{ t('agents.run.usedCredits') }}</span>
      </v-col>
    </v-row>

    <v-progress-linear v-if="progress.total_queries > 0"
      :model-value="(progress.completed_queries + progress.empty_queries) / progress.total_queries * 100"
      :color="scrapeColor" height="4" rounded class="mt-3" />
    <v-progress-linear v-else indeterminate color="primary" height="4" rounded class="mt-3" />
  </div>
</template>

<script setup lang="ts">
import type { RunProgress } from '~~bindings/smegg.me/thughunter/core/scraper/models.js'

const { t } = useI18n()
const props = defineProps<{ progress: RunProgress | null, stopping?: boolean }>()

const remainingQueries = computed(() => {
  const total = props.progress?.total_queries ?? 0
  const completed = props.progress?.completed_queries ?? 0
  const empty = props.progress?.empty_queries ?? 0
  return Math.max(0, total - completed - empty)
})

const scrapeColor = computed(() => {
  const completed = props.progress?.completed_queries ?? 0
  const empty = props.progress?.empty_queries ?? 0
  const total = completed + empty
  if (total === 0) return 'primary'
  const successRate = completed / total
  if (successRate >= 0.8) return 'primary'
  if (successRate >= 0.5) return 'warning'
  return 'error'
})
</script>
