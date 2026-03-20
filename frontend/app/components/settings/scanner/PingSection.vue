<!-- app/components/settings/scanner/PingSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scanner.ping')">
    <SettingsCommonItem :title="t('settings.scanner.pingMode')" :description="t('settings.scanner.pingModeDesc')">
      <v-btn-toggle v-model="config.scanner.ping_mode" density="compact" variant="flat" color="primary" mandatory
        @update:model-value="$emit('save')">
        <v-btn value="soft">{{ t('settings.scanner.pingModes.soft') }}</v-btn>
        <v-btn value="strict">{{ t('settings.scanner.pingModes.strict') }}</v-btn>
      </v-btn-toggle>
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scanner.icmpPing')" :description="t('settings.scanner.icmpPingDesc')">
      <v-switch v-model="config.scanner.icmp_ping" hide-details density="compact" color="primary"
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
