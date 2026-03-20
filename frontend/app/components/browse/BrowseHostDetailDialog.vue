<!-- app/components/browse/BrowseHostDetailDialog.vue -->
<template>
  <v-dialog v-model="model" max-width="600" scrollable>
    <v-card rounded="lg">
      <v-card-title class="d-flex align-center">
        <v-icon icon="mdi-server" size="20" class="mr-2" />
        <span style="font-family: monospace">{{ host.ip }}</span>
        <v-chip v-if="host.ping_ms > 0" size="x-small" :color="hostPingColor" variant="tonal"
          class="font-weight-medium ml-3">
          {{ host.ping_ms.toFixed(0) }} ms
        </v-chip>
        <v-spacer />
        <v-btn :icon="isFavorite ? 'mdi-star' : 'mdi-star-outline'" :color="isFavorite ? 'amber' : undefined"
          variant="text" density="compact" class="mr-1" @click="toggleFavorite" />
        <v-btn icon variant="text" density="compact" @click="model = false">
          <v-icon icon="mdi-close" />
        </v-btn>
      </v-card-title>
      <v-divider />

      <v-card-text class="pt-4">
        <v-row dense>
          <!-- Country -->
          <v-col cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.country') }}</div>
            <div class="text-body-2 d-flex align-center">
              <template v-if="host.country_code">
                <span class="mr-1 text-body-1">{{ flag }}</span>{{ name }}
              </template>
              <span v-else class="opacity-40">{{ t('browse.host.unknownLocation') }}</span>
            </div>
          </v-col>

          <!-- Location -->
          <v-col cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.location') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon icon="mdi-map-marker" size="14" class="mr-1" />
              {{ location || t('browse.host.unknownLocation') }}
            </div>
          </v-col>

          <!-- OS -->
          <v-col v-if="host.os" cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.os') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon icon="mdi-monitor" size="14" class="mr-1" />{{ host.os }}
            </div>
          </v-col>

          <!-- Hardware -->
          <v-col v-if="host.hardware" cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.hardware') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon icon="mdi-chip" size="14" class="mr-1" />{{ host.hardware }}
            </div>
          </v-col>

          <!-- Last seen -->
          <v-col v-if="lastSeen" cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.lastSeen') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon icon="mdi-clock-outline" size="14" class="mr-1" />{{ lastSeen }}
            </div>
          </v-col>

          <!-- Services -->
          <v-col v-if="serviceNames.length" cols="12">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.services') }}</div>
            <div class="d-flex flex-column ga-2 mt-1">
              <div v-for="svc in serviceNames" :key="svc" class="d-flex align-center ga-1 flex-wrap">
                <v-chip size="small" color="primary" variant="tonal" class="font-weight-bold text-uppercase">
                  {{ svc }}
                </v-chip>
                <v-chip v-for="port in (host.services[svc] ?? [])" :key="port" size="small" variant="tonal"
                  class="font-weight-medium cursor-pointer" style="font-family: monospace"
                  @click="openService(svc, port)">
                  {{ port }}
                </v-chip>
              </div>
            </div>
          </v-col>

          <!-- Software -->
          <v-col v-if="host.software?.length" cols="12">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.software') }}</div>
            <div class="d-flex ga-1 flex-wrap mt-1">
              <v-chip v-for="sw in host.software" :key="sw" size="small" variant="tonal">{{ sw }}</v-chip>
            </div>
          </v-col>

          <!-- Labels -->
          <v-col v-if="host.labels?.length" cols="12">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.labels') }}</div>
            <div class="d-flex ga-1 flex-wrap mt-1">
              <v-chip v-for="label in host.labels" :key="label" size="small" color="secondary" variant="tonal">
                {{ label }}
              </v-chip>
            </div>
          </v-col>
        </v-row>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import type { HostItem } from '~/types/scanner'
import * as ProgramService from '~~bindings/smegg.me/thughunter/services/program/service.js'
import { countryFlag as toFlag, countryName as toName } from '~/utils/country'
import { pingColor as toPingColor } from '~/utils/color'

const model = defineModel<boolean>({ default: false })
const props = defineProps<{ host: HostItem }>()
const emit = defineEmits<{ favoriteChanged: [id: number, value: boolean] }>()
const { t, locale } = useI18n()
const dateLocale = useDateLocale()
const { toggleHostFavorite } = useScanner()

const isFavorite = ref(props.host.is_favorite)

watch(() => props.host, (h) => {
  isFavorite.value = h.is_favorite
})

async function toggleFavorite() {
  const newVal = await toggleHostFavorite(props.host.ID)
  isFavorite.value = newVal
  emit('favoriteChanged', props.host.ID, newVal)
}

const location = computed(() => {
  const parts = [props.host.city, props.host.region].filter(Boolean)
  return parts.join(', ')
})

const flag = computed(() => toFlag(props.host.country_code))
const name = computed(() => toName(props.host.country_code, locale.value))
const serviceNames = computed(() => Object.keys(props.host.services ?? {}))

const lastSeen = computed(() => {
  if (!props.host.UpdatedAt) return ''
  try {
    return new Date(props.host.UpdatedAt).toLocaleString(dateLocale.value, {
      dateStyle: 'medium',
      timeStyle: 'short',
    })
  }
  catch {
    return ''
  }
})

const hostPingColor = computed(() => toPingColor(props.host.ping_ms))

function openService(service: string, port: string) {
  ProgramService.OpenService(service, props.host.ip, port)
}
</script>
