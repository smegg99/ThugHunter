<!-- app/components/browse/BrowseHostCard.vue -->
<template>
  <v-card variant="tonal" rounded="lg" class="h-100" style="cursor: pointer" role="button" tabindex="0"
    @click="dialogOpen = true" @keydown.enter="dialogOpen = true" @keydown.space.prevent="dialogOpen = true">
    <v-card-text class="d-flex flex-column ga-1 pa-3">
      <!-- Header row: IP + ping badge -->
      <div class="d-flex align-center justify-space-between ga-2">
        <span class="text-body-2 font-weight-bold text-truncate" style="font-family: monospace">{{ host.ip }}</span>
        <div class="d-flex align-center ga-1 flex-shrink-0">
          <v-icon v-if="host.is_favorite" icon="mdi-star" size="14" color="amber" />
          <v-chip v-if="host.ping_ms > 0" size="x-small" :color="hostPingColor" variant="tonal" class="font-weight-medium">
            {{ host.ping_ms.toFixed(0) }} ms
          </v-chip>
          <v-icon v-else icon="mdi-lan-disconnect" size="14" color="error" :title="t('browse.host.notResponding')"
            class="opacity-50" />
        </div>
      </div>

      <!-- Country flag + name -->
      <div class="d-flex align-center text-caption text-medium-emphasis">
        <template v-if="host.country_code">
          <span class="flex-shrink-0 mr-1">{{ flag }}</span>
          <span class="text-truncate">{{ name }}</span>
          <span v-if="host.city" class="text-truncate ml-2 opacity-60">{{ host.city }}</span>
        </template>
        <span v-else class="opacity-40">{{ t('browse.host.unknownLocation') }}</span>
      </div>

      <!-- OS -->
      <div class="d-flex align-center text-caption text-medium-emphasis">
        <v-icon icon="mdi-monitor" size="11" class="flex-shrink-0 mr-1" />
        <span v-if="host.os" class="text-truncate mr-2">{{ host.os }}</span>
        <span v-else class="text-truncate">{{ t('browse.host.unknownOS') }}</span>
      </div>

      <!-- Hardware -->
      <div class="d-flex align-center text-caption text-medium-emphasis">
        <v-icon icon="mdi-chip" size="11" class="flex-shrink-0 mr-1" />
        <span v-if="host.hardware" class="text-truncate mr-2 opacity-60">{{ host.hardware }}</span>
        <span v-else class="text-truncate">{{ t('browse.host.unknownHardware') }}</span>
      </div>

      <!-- Service indicators -->
      <div class="d-flex ga-1 flex-nowrap overflow-hidden mt-auto">
        <v-chip v-for="svc in visibleServices" :key="svc" size="x-small" color="primary" variant="tonal"
          class="font-weight-bold text-uppercase">
          {{ svc }}
        </v-chip>
        <v-chip v-if="extraCount > 0" size="x-small" variant="tonal" class="font-weight-bold text-medium-emphasis">
          +{{ extraCount }}
        </v-chip>
      </div>
    </v-card-text>
  </v-card>

  <BrowseHostDetailDialog v-model="dialogOpen" :host="host" @favorite-changed="onFavoriteChanged" />
</template>

<script setup lang="ts">
import type { HostItem } from '~/types/scanner'
import { countryFlag as toFlag, countryName as toName } from '~/utils/country'
import { pingColor as toPingColor } from '~/utils/color'

const props = defineProps<{ host: HostItem }>()
const emit = defineEmits<{ favoriteChanged: [id: number, value: boolean] }>()
const { t, locale } = useI18n()

const dialogOpen = ref(false)

function onFavoriteChanged(id: number, value: boolean) {
  emit('favoriteChanged', id, value)
}

const flag = computed(() => toFlag(props.host.country_code))
const name = computed(() => toName(props.host.country_code, locale.value))
const serviceNames = computed(() => Object.keys(props.host.services ?? {}))
const visibleServices = computed(() => serviceNames.value.slice(0, 4))
const extraCount = computed(() => Math.max(0, serviceNames.value.length - 4))
const hostPingColor = computed(() => toPingColor(props.host.ping_ms))
</script>
