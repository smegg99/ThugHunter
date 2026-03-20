<!-- app/components/settings/scraper/TemplatesSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scraper.templates')">
    <template #header-actions>
      <v-btn icon variant="text" density="compact" @click="helpOpen = true">
        <v-icon icon="mdi-help-circle-outline" />
      </v-btn>
    </template>

    <SettingsCommonItem :title="t('settings.scraper.emailTemplate')">
      <v-text-field v-model="config.scraper.agents.templates.email" density="compact" variant="solo" hide-details
        style="min-width: 280px; width: 380px; max-width: 420px" @blur="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scraper.passwordTemplate')">
      <v-text-field v-model="config.scraper.agents.templates.password" density="compact" variant="solo" hide-details
        style="min-width: 280px; width: 380px; max-width: 420px" @blur="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scraper.organizationTemplate')">
      <v-text-field v-model="config.scraper.agents.templates.organization" density="compact" variant="solo" hide-details
        style="min-width: 280px; width: 380px; max-width: 420px" @blur="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scraper.firstNameTemplate')">
      <v-text-field v-model="config.scraper.agents.templates.first_name" density="compact" variant="solo" hide-details
        style="min-width: 280px; width: 380px; max-width: 420px" @blur="$emit('save')" />
    </SettingsCommonItem>

    <SettingsCommonItem :title="t('settings.scraper.lastNameTemplate')">
      <v-text-field v-model="config.scraper.agents.templates.last_name" density="compact" variant="solo" hide-details
        style="min-width: 280px; width: 380px; max-width: 420px" @blur="$emit('save')" />
    </SettingsCommonItem>
  </SettingsCommonSection>

  <v-dialog v-model="helpOpen" max-width="950">
    <v-card>
      <v-card-title class="d-flex align-center">
        {{ t('settings.scraper.templatesHelp.title') }}
        <v-spacer />
        <v-btn icon variant="text" density="compact" @click="helpOpen = false">
          <v-icon icon="mdi-close" />
        </v-btn>
      </v-card-title>
      <v-card-text class="text-medium-emphasis" style="white-space: pre-line;">
        {{ t('settings.scraper.templatesHelp.body') }}
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
defineEmits<{
  save: []
}>()

const { config } = useConfigSync()
const { t } = useI18n()

const helpOpen = ref(false)
</script>
