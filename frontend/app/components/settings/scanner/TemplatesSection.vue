<!-- app/components/settings/scanner/TemplatesSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scanner.templates')">
    <template #header-actions>
      <v-btn icon variant="text" density="compact" @click="helpOpen = true">
        <v-icon icon="mdi-help-circle-outline" />
      </v-btn>
    </template>

    <SettingsCommonItem v-for="p in protocols" :key="p.key" :title="t(`settings.scanner.${p.key}`)"
      :description="t(`settings.scanner.${p.key}Desc`)">
      <v-text-field v-model="(config.scanner.templates as any)[p.field]" density="compact" variant="solo" flat
        hide-details style="min-width: 280px; width: 420px; max-width: 460px" @blur="$emit('save')" />
    </SettingsCommonItem>
  </SettingsCommonSection>

  <v-dialog v-model="helpOpen" max-width="750">
    <v-card>
      <v-card-title class="d-flex align-center">
        {{ t('settings.scanner.templatesHelp.title') }}
        <v-spacer />
        <v-btn icon variant="text" density="compact" @click="helpOpen = false">
          <v-icon icon="mdi-close" />
        </v-btn>
      </v-card-title>
      <v-card-text class="text-medium-emphasis" style="white-space: pre-line;">
        {{ t('settings.scanner.templatesHelp.body') }}
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

const protocols = [
  { key: 'vncCommand', field: 'vnc_command' },
  { key: 'rdpCommand', field: 'rdp_command' },
  { key: 'spiceCommand', field: 'spice_command' },
  { key: 'sshCommand', field: 'ssh_command' },
  { key: 'httpCommand', field: 'http_command' },
  { key: 'httpsCommand', field: 'https_command' },
  { key: 'screenshotCommand', field: 'screenshot_command' },
]
</script>
