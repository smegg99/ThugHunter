<!-- app/components/settings/scraper/BrowserSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scraper.browser')">
    <SettingsCommonItem :title="t('settings.scraper.browserBinaryPath')"
      :description="t('settings.scraper.browserBinaryPathDesc')">
      <v-text-field v-model="config.scraper.browser_binary_path" density="compact" variant="solo" hide-details
        style="min-width: 280px; max-width: 400px" @blur="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scraper.virtualDisplay')"
      :description="virtualDisplayAvailable ? t('settings.scraper.virtualDisplayDesc') : t('settings.scraper.virtualDisplayUnavailable')">
      <v-switch v-model="config.scraper.virtual_display" hide-details density="compact" color="primary"
        :disabled="!virtualDisplayAvailable" @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scraper.minimalBrowser')"
      :description="t('settings.scraper.minimalBrowserDesc')">
      <v-switch v-model="config.scraper.minimal_browser" hide-details density="compact" color="primary"
        @update:model-value="$emit('save')" />
    </SettingsCommonItem>
  </SettingsCommonSection>
</template>

<script setup lang="ts">
import * as ConfigService from '~~bindings/smegg.me/thughunter/services/config/service.js'

defineEmits<{
  save: []
}>()

const { config } = useConfigSync()
const { t } = useI18n()

const virtualDisplayAvailable = ref(false)

onMounted(async () => {
  try {
    virtualDisplayAvailable.value = await ConfigService.IsVirtualDisplayAvailable()
  }
  catch {
    virtualDisplayAvailable.value = false
  }
})
</script>
