<!-- app/pages/settings/preferences.vue -->
<template>
  <v-container>
    <SettingsCommonSection :title="t('settings.general')">
      <SettingsCommonItem :title="t('settings.language')">
        <SettingsAppearanceLanguageSelect />
      </SettingsCommonItem>
      <SettingsCommonItem :title="t('settings.preferences.closeToTray')"
        :description="t('settings.preferences.closeToTrayDesc')">
        <v-switch v-model="closeToTray" hide-details density="compact" color="primary" />
      </SettingsCommonItem>
    </SettingsCommonSection>

    <SettingsCommonSection :title="t('settings.appearance')">
      <SettingsCommonItem :title="t('settings.theme')">
        <SettingsAppearanceThemeButtonGroup />
      </SettingsCommonItem>

      <SettingsCommonItem :title="t('settings.variant')">
        <SettingsAppearanceVariantButtonGroup />
      </SettingsCommonItem>

      <SettingsCommonItem :title="t('settings.accent')">
        <SettingsAppearanceAccentButtonGroup />
      </SettingsCommonItem>

      <SettingsCommonItem :title="t('settings.memoryUnit')">
        <SettingsAppearanceMemoryUnitButtonGroup />
      </SettingsCommonItem>
    </SettingsCommonSection>
  </v-container>
</template>

<script setup lang="ts">
const { t } = useI18n()
const { config } = useConfigSync()

const closeToTray = computed({
  get: () => config.preferences.close_to_tray,
  set: (v: boolean) => {
    config.preferences.close_to_tray = v
  },
})
</script>
