<!-- app/components/settings/appearance/AccentButtonGroup.vue -->
<template>
  <div class="d-flex align-center ga-3 flex-wrap">
    <v-btn-toggle v-model="accentMode" mandatory density="comfortable" color="primary">
      <v-btn :value="ACCENT_MODE.AUTO" size="small" :disabled="!isAccentWatchSupported">
        <v-icon icon="mdi-monitor-shimmer" start />
        {{ t('settings.accentModes.auto') }}
      </v-btn>
      <v-btn :value="ACCENT_MODE.CUSTOM" size="small">
        <v-icon icon="mdi-palette" start />
        {{ t('settings.accentModes.custom') }}
      </v-btn>
    </v-btn-toggle>

    <button class="accent-dot rounded-circle flex-shrink-0 border pa-0" :style="{ backgroundColor: effectiveAccent }"
      :aria-label="t('settings.accentModes.custom')" :disabled="accentMode !== ACCENT_MODE.CUSTOM"
      @click="showPicker = true" />

    <SettingsAppearanceAccentColorDialog v-model="showPicker" v-model:color="customAccentColor" />
  </div>
</template>

<script setup lang="ts">
import { useThemeSync } from '~/composables/useThemeSync'
import { ACCENT_MODE } from '~/types/config'
import type { AccentMode } from '~/types/config'

const { prefs, effectiveAccent, isAccentWatchSupported } = useThemeSync()
const { t } = useI18n()

const showPicker = ref(false)

const accentMode = computed<AccentMode>({
  get: () => (prefs.value.accent_mode || ACCENT_MODE.CUSTOM) as AccentMode,
  set: (v: AccentMode) => {
    prefs.value.accent_mode = v
  },
})

const customAccentColor = computed<string>({
  get: () => prefs.value.accent_color || '#7B5CDE',
  set: (v: string) => {
    prefs.value.accent_color = v
  },
})
</script>

<style scoped>
.accent-dot {
  width: 28px;
  height: 28px;
  background: none;
  transition: transform 0.15s ease, box-shadow 0.15s ease, opacity 0.15s ease;
}

.accent-dot:disabled {
  opacity: 0.45;
}

.accent-dot:not(:disabled) {
  cursor: pointer;
}

.accent-dot:not(:disabled):hover {
  transform: scale(1.15);
  box-shadow: 0 0 0 2px rgb(var(--v-theme-primary));
}
</style>
