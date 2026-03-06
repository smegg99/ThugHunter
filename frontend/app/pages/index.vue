<!-- frontend/app/pages/index.vue -->
<template>
  <v-main class="pa-6">
    <h1 class="text-h5 mb-4">
      Nuxt + Wails + Vuetify
    </h1>

    <div class="mt-4 d-flex align-center ga-3">
      <v-icon icon="mdi-numeric" />
      <v-icon icon="mdi-home" />
      <v-icon icon="mdi-cog" />
    </div>

    <v-divider class="my-6" />
    <h2 class="text-h6 mb-3">
      {{ t('settings.theme') }}
    </h2>

    <v-btn-toggle v-model="prefs.theme" mandatory density="comfortable" color="primary" divided>
      <v-btn v-for="thm in themeOptions" :key="thm" :value="thm" size="small">
        {{ t(`settings.themes.${thm}`) }}
      </v-btn>
    </v-btn-toggle>

    <h2 class="text-h6 mt-6 mb-3">
      {{ t('settings.language') }}
    </h2>

    <v-select v-model="prefs.language" :items="localeItems" item-title="name" item-value="code" density="comfortable"
      style="max-width: 240px" />

    <v-divider class="my-6" />

    <h2 class="text-h6 mb-3">
      {{ t('settings.accent') }}
    </h2>

    <v-btn-toggle v-model="prefs.accent_mode" mandatory density="comfortable" color="primary" divided class="mb-4">
      <v-btn v-for="mode in accentModes" :key="mode" :value="mode" size="small"
        :disabled="mode === ACCENT_MODE.AUTO && !isAccentWatchSupported">
        {{ t(`settings.accentModes.${mode}`) }}
      </v-btn>
    </v-btn-toggle>

    <v-alert v-if="!isAccentWatchSupported" type="info" density="compact" class="mb-4">
      {{ t('settings.accentWatchNotSupported') }}
    </v-alert>

    <div v-if="prefs.accent_mode === ACCENT_MODE.CUSTOM" class="mt-4">
      <v-color-picker v-model="customAccentColor" mode="hex" :modes="['hex', 'hsl', 'rgb']" />
    </div>

    <div class="mt-4 d-flex align-center ga-3">
      <div class="accent-preview" :style="{ backgroundColor: effectiveAccent }" />
      <span class="text-body-2 text-medium-emphasis">{{ t('settings.currentAccent') }}: {{ effectiveAccent }}</span>
    </div>
  </v-main>
</template>

<script setup lang="ts">
import { useThemeSync } from '~/composables/useThemeSync'
import { ACCENT_MODE, THEME_MODES, ACCENT_MODES } from '~/types/config'
import type { ThemeMode, AccentMode } from '~/types/config'

const { prefs, localeItems, effectiveAccent, isAccentWatchSupported } = useThemeSync()
const { t } = useI18n()

const themeOptions: ThemeMode[] = THEME_MODES
const accentModes: AccentMode[] = ACCENT_MODES

// Two-way binding for the color picker that syncs to the config
const customAccentColor = computed({
  get: () => prefs.value.accent_color || '#7B5CDE',
  set: (v: string) => {
    prefs.value.accent_color = v
  },
})
</script>

<style scoped>
.accent-preview {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
</style>
