<!-- app/components/settings/appearance/MemoryUnitButtonGroup.vue -->
<template>
  <v-btn-toggle v-model="memoryUnit" mandatory density="comfortable" color="primary">
    <v-btn value="auto" size="small">{{ t('settings.memoryUnits.auto') }}</v-btn>
    <v-btn value="mb" size="small">MB</v-btn>
    <v-btn value="mib" size="small">MiB</v-btn>
  </v-btn-toggle>
</template>

<script setup lang="ts">
import { useThemeSync } from '~/composables/useThemeSync'

const { prefs } = useThemeSync()
const { t } = useI18n()

const memoryUnit = computed({
  get: () => prefs.value.memory_unit || 'auto',
  set: (v: string) => {
    prefs.value.memory_unit = v
  },
})
</script>
