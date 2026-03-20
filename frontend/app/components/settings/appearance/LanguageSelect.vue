<!-- app/components/settings/appearance/LanguageSelect.vue -->
<template>
  <v-select v-model="language" :items="localeItems" item-title="name" item-value="code" density="compact" variant="solo"
    hide-details style="max-width: 240px" />
</template>

<script setup lang="ts">
import { useThemeSync } from '~/composables/useThemeSync'

const { prefs, localeItems } = useThemeSync()
const { t, locale } = useI18n()

const language = computed({
  get: () => prefs.value.language || locale.value || 'en',
  set: (v: string) => {
    prefs.value.language = v
  },
})
</script>
