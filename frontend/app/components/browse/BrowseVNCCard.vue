<!-- app/components/browse/BrowseVNCCard.vue -->
<template>
  <v-card variant="tonal" rounded="t-lg" class="h-100" style="cursor: pointer" :style="{ background: ipBg }"
    role="button" tabindex="0" @click="dialogOpen = true" @keydown.enter="dialogOpen = true"
    @keydown.space.prevent="dialogOpen = true">
    <!-- 16:9 image placeholder - click launches VNC client -->
    <div class="vnc-thumb d-flex align-center justify-center bg-surface-light" @click.stop="openVNC">
      <v-img v-if="props.screenshotSrc" :src="props.screenshotSrc" cover rounded="t-lg" alt="VNC screenshot"
        :class="{ 'vnc-stale': service.stale_screenshot && service.has_screenshot }" />
      <v-progress-circular v-else-if="service.no_auth && !service.has_screenshot" indeterminate size="24" width="2"
        color="grey-lighten-1" />
      <v-icon v-else icon="mdi-monitor-screenshot" size="32" class="text-medium-emphasis" />
      <div v-if="service.stale_screenshot && service.has_screenshot && props.screenshotSrc" class="vnc-stale-badge"
        :title="t('browse.vnc.outdated')">
        <v-icon icon="mdi-clock-alert-outline" size="14" color="white" />
      </div>
    </div>

    <!-- Body -->
    <v-card-text class="d-flex flex-column ga-1 pa-3 flex-grow-1">
      <!-- Header: IP:Port + lock icon + latency -->
      <div class="d-flex align-center justify-space-between ga-2">
        <span class="text-body-2 font-weight-bold text-truncate" style="font-family: monospace">
          {{ service.ip }}:{{ service.port }}
        </span>
        <div class="d-flex align-center ga-1 flex-shrink-0">
          <v-icon v-if="service.is_favorite" icon="mdi-star" size="14" color="warning" />
          <v-icon :icon="authIcon" :color="authColor" size="14" />
          <v-chip v-if="service.latency_ms > 0" size="x-small" color="info" variant="tonal" class="font-weight-medium">
            {{ service.latency_ms.toFixed(0) }} ms
          </v-chip>
          <v-icon v-else icon="mdi-lan-disconnect" size="14" color="error" :title="t('browse.vnc.notResponding')"
            class="opacity-50" />
        </div>
      </div>

      <!-- Screenshot timestamp -->
      <div v-if="screenshotTime" class="text-caption text-medium-emphasis opacity-60">
        <v-icon icon="mdi-camera" size="12" class="mr-1" />{{ screenshotTime }}
      </div>

      <!-- Host info: OS + hardware -->
      <div v-if="service.os || service.hardware" class="d-flex align-center ga-1 text-caption text-medium-emphasis">
        <span v-if="service.os" class="text-truncate">{{ service.os }}</span>
        <span v-if="service.os && service.hardware" class="opacity-40">&middot;</span>
        <span v-if="service.hardware" class="text-truncate opacity-60">{{ service.hardware }}</span>
      </div>

      <!-- Country + city -->
      <div class="d-flex align-center text-caption text-medium-emphasis">
        <template v-if="service.country_code">
          <span class="flex-shrink-0 mr-1">{{ flag }}</span>
          <span class="text-truncate">{{ name }}</span>
          <span v-if="service.city" class="text-truncate ml-2 opacity-60">{{ service.city }}</span>
        </template>
        <span v-else class="opacity-40">{{ t('browse.host.unknownLocation') }}</span>
      </div>
    </v-card-text>
  </v-card>

  <BrowseVNCDetailDialog v-model="dialogOpen" :service="service" @favorite-changed="onFavoriteChanged" />
</template>

<script setup lang="ts">
import type { VNCItem } from '~/types/scanner'
import * as ProgramService from '~~bindings/smegg.me/thughunter/services/program/service.js'
import { countryFlag as toFlag, countryName as toName } from '~/utils/country'

const props = defineProps<{ service: VNCItem; screenshotSrc?: string }>()
const emit = defineEmits<{ favoriteChanged: [id: number, value: boolean] }>()
const { t, locale } = useI18n()
const dateLocale = useDateLocale()

function onFavoriteChanged(id: number, value: boolean) {
  emit('favoriteChanged', id, value)
}

function openVNC() {
  ProgramService.OpenService('vnc', props.service.ip, String(props.service.port))
}

const dialogOpen = ref(false)

const screenshotTime = computed(() => {
  const ts = props.service.screenshot_at
  if (!ts) return ''
  try {
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ''
    return d.toLocaleString(dateLocale.value, { dateStyle: 'medium', timeStyle: 'short' })
  }
  catch { return '' }
})

const authIcon = computed(() => {
  if (props.service.no_auth) return 'mdi-lock-open-variant'
  if (props.service.has_screenshot) return 'mdi-lock-question'
  return 'mdi-lock'
})

const authColor = computed(() => {
  if (props.service.no_auth) return 'error'
  if (props.service.has_screenshot) return 'warning'
  return undefined
})

const flag = computed(() => toFlag(props.service.country_code))
const name = computed(() => toName(props.service.country_code, locale.value))

const ipBg = computed(() => {
  const parts = props.service.ip.split('.')
  if (parts.length !== 4) return undefined
  const n = (Number(parts[0]) * 7 + Number(parts[1]) * 13 + Number(parts[2]) * 17 + Number(parts[3]) * 3) % 360
  return `hsla(${n}, 50%, 50%, 0.08)`
})
</script>

<style scoped>
.vnc-thumb {
  aspect-ratio: 16 / 9;
  width: 100%;
  flex-shrink: 0;
  overflow: hidden;
  position: relative;
}

.vnc-stale {
  filter: saturate(0.8);
}

.vnc-stale-badge {
  position: absolute;
  top: 6px;
  right: 6px;
  pointer-events: none;
  background: rgba(var(--v-theme-warning), 0.85);
  border-radius: 50%;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
