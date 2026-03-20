<!-- app/components/settings/scraper/AgentsSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scraper.agents')">
    <SettingsCommonItem :title="t('settings.scraper.maxAgents')" :description="t('settings.scraper.maxAgentsDesc')">
      <v-number-input v-model="config.scraper.agents.max_agents" :min="1" density="compact" variant="solo" hide-details
        style="min-width: 120px; max-width: 240px" controlVariant="stacked" @update:model-value="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scraper.maxRegisterRetries')"
      :description="t('settings.scraper.maxRegisterRetriesDesc')">
      <v-number-input v-model="config.scraper.agents.max_register_retries" :min="0" density="compact" variant="solo"
        hide-details style="min-width: 120px; max-width: 240px" controlVariant="stacked"
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
