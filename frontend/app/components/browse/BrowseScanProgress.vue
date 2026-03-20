<!-- app/components/browse/BrowseScanProgress.vue -->
<template>
  <!-- Compact inline variant: shown in the app bar next to the search bar -->
  <Transition v-if="compact" name="browse-progress">
    <div v-if="progress" class="d-flex align-center text-caption text-medium-emphasis"
      style="gap: 16px; white-space: nowrap; font-family: monospace">
      <Transition name="browse-progress-icon" mode="out-in">
        <v-progress-circular v-if="running" key="spinner" indeterminate size="18" width="2" color="primary"
          class="flex-shrink-0" />
        <v-icon v-else key="check" icon="mdi-check-circle" size="18" class="flex-shrink-0" />
      </Transition>

      <!-- Stage indicator -->
      <template v-if="isScreenshotMode">
        <span class="font-weight-medium text-secondary d-inline-flex align-center" style="min-width: 7ch">
          <v-icon icon="mdi-camera" size="16" class="mr-1" />{{ progress.screenshot_done }}/{{ progress.screenshot_total
          }}
        </span>
      </template>
      <template v-else>
        <span class="font-weight-medium d-inline-flex align-center" style="min-width: 7ch" :class="running ? 'text-primary' : 'text-success'">
          {{ progress.scanned }}/{{ progress.total_hosts }}
        </span>
        <span class="d-inline-flex align-center" style="min-width: 7ch"><v-icon icon="mdi-table-tennis" size="18" class="mr-1" />{{ progress.ping_ok }}</span>
      </template>
      <span class="d-inline-flex align-center" style="min-width: 7ch"><v-icon icon="mdi-timer-outline" size="18" class="mr-1" />{{ Math.round(progress.elapsed_secs) }}s</span>
    </div>
  </Transition>

  <!-- Full banner variant -->
  <Transition v-else name="browse-progress">
    <v-alert v-if="progress" variant="tonal" :color="alertColor" density="compact" class="mb-4" rounded="lg">
      <div class="d-flex align-center ga-3 flex-wrap">
        <v-progress-circular v-if="running" indeterminate size="18" width="2" :color="alertColor"
          class="flex-shrink-0" />
        <v-icon v-else icon="mdi-check-circle" size="18" color="success" class="flex-shrink-0" />

        <span class="text-body-2 font-weight-medium">
          <template v-if="isScreenshotMode">
            {{ t('browse.screenshotProgress', { done: progress.screenshot_done, total: progress.screenshot_total }) }}
          </template>
          <template v-else-if="running">
            {{ t('browse.scanProgress', { scanned: progress.scanned, total: progress.total_hosts }) }}
          </template>
          <template v-else>
            {{ t('browse.scanDone', { saved: progress.saved }) }}
          </template>
        </span>

        <div class="d-flex ga-4 text-caption text-medium-emphasis ml-auto flex-wrap">
          <span v-if="!isScreenshotMode"><v-icon icon="mdi-wifi" size="12" class="mr-1" />{{ progress.ping_ok }}</span>
          <span v-if="!isScreenshotMode"><v-icon icon="mdi-database-check" size="12" class="mr-1" />{{ progress.saved }}</span>
          <span v-if="progress.screenshot_stage >= 1">
            <v-icon icon="mdi-camera" size="12" class="mr-1" />{{ progress.screenshot_done }}/{{
              progress.screenshot_total }}
          </span>
          <span v-if="progress.elapsed_secs > 0">
            <v-icon icon="mdi-timer-outline" size="12" class="mr-1" />{{ Math.round(progress.elapsed_secs) }}s
          </span>
        </div>
      </div>

      <!-- Progress bars -->
      <div v-if="running" class="mt-2 d-flex flex-column ga-1">
        <v-progress-linear v-if="!isScreenshotMode && progress.total_hosts > 0"
          :model-value="(progress.scanned / progress.total_hosts) * 100" color="primary" height="2" rounded />
        <v-progress-linear v-if="progress.screenshot_stage === 1 && progress.screenshot_total > 0"
          :model-value="(progress.screenshot_done / progress.screenshot_total) * 100" color="secondary" height="2"
          rounded />
      </div>
    </v-alert>
  </Transition>
</template>

<script setup lang="ts">
import type { ScanProgressData } from '~/types/scanner'

const props = withDefaults(defineProps<{
  progress: ScanProgressData | null
  running: boolean
  compact?: boolean
}>(), { compact: false })

const { t } = useI18n()

const isScreenshotMode = computed(() =>
  props.progress?.mode === 'screenshots' || (props.running && props.progress?.screenshot_stage === 1),
)

const alertColor = computed(() => {
  if (!props.running) return 'success'
  return isScreenshotMode.value ? 'secondary' : 'primary'
})
</script>

<style scoped>
/* Slide + fade in/out for the whole row */
.browse-progress-enter-active,
.browse-progress-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.browse-progress-enter-from,
.browse-progress-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* Crossfade for spinner <-> check icon */
.browse-progress-icon-enter-active,
.browse-progress-icon-leave-active {
  transition: opacity 0.2s ease;
}

.browse-progress-icon-enter-from,
.browse-progress-icon-leave-to {
  opacity: 0;
}
</style>