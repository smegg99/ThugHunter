<!-- app/components/settings/scanner/ScreenshotSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scanner.screenshots')">
    <SettingsCommonItem :title="t('settings.scanner.screenshotMaxWorkers')"
      :description="t('settings.scanner.screenshotMaxWorkersDesc')">
      <v-number-input v-model="config.scanner.workers.screenshot_max_workers" :min="1" :max="3500" density="compact"
        variant="solo" hide-details style="min-width: 120px; max-width: 240px" controlVariant="stacked"
        @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.screenshotTimeout')"
      :description="t('settings.scanner.screenshotTimeoutDesc')">
      <v-number-input v-model="config.scanner.workers.screenshot_timeout_seconds" :min="1" :max="240" suffix="s"
        density="compact" variant="solo" hide-details style="min-width: 120px; max-width: 240px"
        controlVariant="stacked" @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.screenshotDelay')"
      :description="t('settings.scanner.screenshotDelayDesc')">
      <v-number-input v-model="config.scanner.workers.screenshot_delay_seconds" :min="1" :max="240" suffix="s"
        density="compact" variant="solo" hide-details style="min-width: 120px; max-width: 240px"
        controlVariant="stacked" @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.screenshotPause')"
      :description="t('settings.scanner.screenshotPauseDesc')">
      <v-number-input v-model="config.scanner.workers.screenshot_pause_seconds" :min="1" :max="240" suffix="s"
        density="compact" variant="solo" hide-details style="min-width: 120px; max-width: 240px"
        controlVariant="stacked" @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.rejectBlankScreenshots')"
      :description="t('settings.scanner.rejectBlankScreenshotsDesc')">
      <v-switch v-model="config.scanner.reject_blank_screenshots" hide-details density="compact" color="primary"
        @update:model-value="$emit('save')" />
    </SettingsCommonItem>
  </SettingsCommonSection>
</template>

<script setup lang="ts">
defineEmits<{
  save: []
}>()

const { config } = useConfigSync()
const { t } = useI18n()
</script>
