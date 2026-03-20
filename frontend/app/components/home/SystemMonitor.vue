<!-- app/components/home/SystemMonitor.vue -->
<template>
  <div class="pa-5">
    <!-- CPU Sparkline -->
    <div class="mb-8">
      <div class="d-flex align-center ga-3 mb-4">
        <div class="text-caption text-medium-emphasis">{{ t('home.monitor.cpuHistory') }}</div>
        <div class="d-flex align-baseline ga-1">
          <span class="text-h5 font-weight-bold" :style="{ color: `rgb(var(--v-theme-${cpuColor}))` }">{{ cpuTotal
          }}</span>
          <span class="text-caption text-medium-emphasis">%</span>
        </div>
      </div>
      <div class="sparkline-wrap" style="font-family: monospace">
        <div class="sparkline-y-axis">
          <span>100%</span>
          <span>75%</span>
          <span>50%</span>
          <span>25%</span>
          <span>0%</span>
        </div>
        <div class="sparkline-chart">
          <v-sparkline :model-value="cpuHistory" :gradient="['rgb(var(--v-theme-primary))']" :line-width="4"
            :padding="4" :smooth="2" :min="0" :max="100" type="bar" height="80" />
          <div class="sparkline-x-axis">
            <span>{{ t('home.monitor.timeAgo', { seconds: historySeconds }) }}</span>
            <span>{{ t('home.monitor.timeNow') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- RAM -->
    <div class="mb-4">
      <div class="d-flex justify-space-between align-center mb-1">
        <span class="text-caption text-medium-emphasis">{{ t('home.monitor.ram') }}</span>
        <span class="text-caption text-medium-emphasis">{{ formatBytes(ram.used) }} / {{ formatBytes(ram.total)
          }}</span>
      </div>
      <v-progress-linear :model-value="ram.usedPercent" :color="percentColor(ram.usedPercent, { danger: 85, warn: 60 })" height="8" rounded />
    </div>

    <!-- Swap -->
    <div v-if="swap.total > 0">
      <div class="d-flex justify-space-between align-center mb-1">
        <span class="text-caption text-medium-emphasis">{{ t('home.monitor.swap') }}</span>
        <span class="text-caption text-medium-emphasis">{{ formatBytes(swap.used) }} / {{ formatBytes(swap.total)
          }}</span>
      </div>
      <v-progress-linear :model-value="swap.usedPercent" :color="percentColor(swap.usedPercent, { danger: 85, warn: 60 })" height="8" rounded />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useThemeSync } from '~/composables/useThemeSync'
import { percentColor } from '~/utils/color'

const { t } = useI18n()
const { snapshot, cpuHistory } = useMonitor()
const { prefs } = useThemeSync()

const isWindows = computed(() => navigator.userAgent.includes('Windows'))
const useBinary = computed(() => {
  const unit = prefs.value.memory_unit || 'auto'
  if (unit === 'mb') return false
  if (unit === 'mib') return true
  return !isWindows.value
})

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0'
  if (useBinary.value) {
    const mib = bytes / (1024 * 1024)
    if (mib >= 1024) return `${(mib / 1024).toFixed(1)} GiB`
    return `${mib.toFixed(0)} MiB`
  }
  const mb = bytes / (1000 * 1000)
  if (mb >= 1000) return `${(mb / 1000).toFixed(1)} GB`
  return `${mb.toFixed(0)} MB`
}

const cpuTotal = computed(() => Math.round(snapshot.value?.cpu.totalPercent ?? 0))
const historySeconds = computed(() => Math.round(cpuHistory.value.length * 1.5))
const ram = computed(() => snapshot.value?.ram ?? { total: 0, used: 0, usedPercent: 0 })
const swap = computed(() => snapshot.value?.swap ?? { total: 0, used: 0, usedPercent: 0 })

const cpuColor = computed(() => percentColor(cpuTotal.value, { danger: 80, warn: 50 }, 'success'))
</script>

<style scoped>
.sparkline-wrap {
  display: flex;
  gap: 6px;
}

.sparkline-y-axis {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  font-size: 10px;
  opacity: 0.5;
  line-height: 1;
  padding-bottom: 16px;
  min-width: 30px;
  text-align: right;
}

.sparkline-chart {
  flex: 1;
  min-width: 0;
}

.sparkline-x-axis {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  opacity: 0.5;
  margin-top: 8px;
  padding: 0 4px;
}
</style>
