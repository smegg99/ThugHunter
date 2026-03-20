<!-- app/components/settings/scanner/WorkersSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scanner.workers')">
    <SettingsCommonItem :title="t('settings.scanner.maxWorkers')" :description="t('settings.scanner.maxWorkersDesc')">
      <v-number-input v-model="config.scanner.workers.max_workers" :min="1" :max="1000" density="compact" variant="solo"
        hide-details style="min-width: 120px; max-width: 240px" controlVariant="stacked"
        @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.pingTimeout')" :description="t('settings.scanner.pingTimeoutDesc')">
      <v-number-input v-model="config.scanner.workers.ping_timeout_seconds" :min="1" suffix="s" density="compact"
        variant="solo" hide-details style="min-width: 120px; max-width: 240px" controlVariant="stacked"
        @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.connectTimeout')"
      :description="t('settings.scanner.connectTimeoutDesc')">
      <v-number-input v-model="config.scanner.workers.connect_timeout_seconds" :min="1" suffix="s" density="compact"
        variant="solo" hide-details style="min-width: 120px; max-width: 240px" controlVariant="stacked"
        @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.bannerTimeout')"
      :description="t('settings.scanner.bannerTimeoutDesc')">
      <v-number-input v-model="config.scanner.workers.banner_timeout_seconds" :min="1" suffix="s" density="compact"
        variant="solo" hide-details style="min-width: 120px; max-width: 240px" controlVariant="stacked"
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
