<!-- app/components/settings/appearance/VariantButtonGroup.vue -->
<template>
  <v-btn-toggle v-model="variant" mandatory density="comfortable" color="primary">
    <v-btn v-for="item in THEME_VARIANTS" :key="item.value" :value="item.value" size="small">
      <v-icon :icon="item.icon" start />
      {{ t(item.labelKey) }}
    </v-btn>
  </v-btn-toggle>
</template>

<script setup lang="ts">
import { useThemeSync } from '~/composables/useThemeSync'
import { THEME_VARIANTS, splitThemeMode, joinThemeMode } from '~/types/config'
import type { ThemeMode, ThemeVariant } from '~/types/config'

const { prefs } = useThemeSync()
const { t } = useI18n()

const variant = computed<ThemeVariant>({
  get: () => splitThemeMode((prefs.value.theme || 'light') as ThemeMode).variant,
  set: (v: ThemeVariant) => {
    const { base } = splitThemeMode((prefs.value.theme || 'light') as ThemeMode)
    prefs.value.theme = joinThemeMode(base, v)
  },
})
</script>
