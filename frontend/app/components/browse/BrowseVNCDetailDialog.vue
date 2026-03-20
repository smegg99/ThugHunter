<!-- app/components/browse/BrowseVNCDetailDialog.vue -->
<template>
  <v-dialog v-model="model" max-width="600" scrollable>
    <v-card rounded="lg">
      <v-card-title class="d-flex align-center">
        <v-icon icon="mdi-monitor" size="20" class="mr-2" />
        <span style="font-family: monospace">{{ service.ip }}:{{ service.port }}</span>
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
          <v-col v-if="service.country_code" cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.country') }}</div>
            <div class="text-body-2 d-flex align-center">
              <span class="mr-1 text-body-1">{{ flag }}</span>{{ name }}
            </div>
          </v-col>

          <!-- City -->
          <v-col v-if="service.city" cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.location') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon icon="mdi-map-marker" size="14" class="mr-1" />{{ service.city }}
            </div>
          </v-col>

          <!-- Latency -->
          <v-col v-if="service.latency_ms > 0" cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.vnc.latency') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon icon="mdi-timer-outline" size="14" class="mr-1" />{{ service.latency_ms.toFixed(0) }} ms
            </div>
          </v-col>

          <!-- Auth type -->
          <v-col cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.vnc.auth') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon :icon="service.no_auth ? 'mdi-lock-open-variant' : 'mdi-lock'" size="14" class="mr-1" />
              {{ service.no_auth ? t('browse.vnc.noAuth') : t('browse.vnc.auth') }}
            </div>
          </v-col>

          <!-- Last seen -->
          <v-col v-if="lastSeen" cols="12" sm="6">
            <div class="text-overline text-medium-emphasis mb-1">{{ t('browse.host.lastSeen') }}</div>
            <div class="text-body-2 d-flex align-center">
              <v-icon icon="mdi-clock-outline" size="14" class="mr-1" />{{ lastSeen }}
            </div>
          </v-col>
        </v-row>
      </v-card-text>

      <v-divider />
      <v-card-actions class="px-4 py-2">
        <v-btn variant="tonal" size="small" prepend-icon="mdi-server" :loading="loadingHost" @click="openHostDetail">
          {{ t('browse.vnc.viewHost') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>

  <BrowseHostDetailDialog v-if="hostData" v-model="hostDialogOpen" :host="hostData" />
</template>

<script setup lang="ts">
import type { VNCItem, HostItem } from '~/types/scanner'
import { countryFlag as toFlag, countryName as toName } from '~/utils/country'

const model = defineModel<boolean>({ default: false })
const props = defineProps<{ service: VNCItem }>()
const emit = defineEmits<{ favoriteChanged: [id: number, value: boolean] }>()
const { t, locale } = useI18n()
const dateLocale = useDateLocale()
const { getHostByIP, toggleVNCFavorite } = useScanner()

const isFavorite = ref(props.service.is_favorite)

watch(() => props.service, (svc) => {
  isFavorite.value = svc.is_favorite
})

async function toggleFavorite() {
  const newVal = await toggleVNCFavorite(props.service.ID)
  isFavorite.value = newVal
  emit('favoriteChanged', props.service.ID, newVal)
}

const loadingHost = ref(false)
const hostDialogOpen = ref(false)
const hostData = ref<HostItem | null>(null)

async function openHostDetail() {
  loadingHost.value = true
  try {
    hostData.value = await getHostByIP(props.service.ip)
    if (hostData.value) {
      hostDialogOpen.value = true
    }
  }
  finally {
    loadingHost.value = false
  }
}

const flag = computed(() => toFlag(props.service.country_code))
const name = computed(() => toName(props.service.country_code, locale.value))

const lastSeen = computed(() => {
  if (!props.service.UpdatedAt) return ''
  try {
    return new Date(props.service.UpdatedAt).toLocaleString(dateLocale.value, {
      dateStyle: 'medium',
      timeStyle: 'short',
    })
  }
  catch {
    return ''
  }
})
</script>
