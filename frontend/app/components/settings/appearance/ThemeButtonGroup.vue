<!-- app/components/settings/appearance/ThemeButtonGroup.vue -->
<template>
  <v-btn-toggle v-model="baseTheme" mandatory density="comfortable" color="primary">
    <v-btn v-for="item in BASE_THEMES" :key="item.value" :value="item.value" size="small">
      <v-icon :icon="item.icon" start />
      {{ t(item.labelKey) }}
    </v-btn>
  </v-btn-toggle>
</template>

<script setup lang="ts">
import { useThemeSync } from '~/composables/useThemeSync'
import { BASE_THEMES, splitThemeMode, joinThemeMode } from '~/types/config'
import type { ThemeMode, BaseTheme } from '~/types/config'

const { prefs } = useThemeSync()
const { t } = useI18n()

const baseTheme = computed<BaseTheme>({
  get: () => splitThemeMode((prefs.value.theme || 'light') as ThemeMode).base,
  set: (base: BaseTheme) => {
    const { variant } = splitThemeMode((prefs.value.theme || 'light') as ThemeMode)
    prefs.value.theme = joinThemeMode(base, variant)
  },
})
</script>
