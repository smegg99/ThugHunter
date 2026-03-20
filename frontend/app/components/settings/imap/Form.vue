<!-- app/components/settings/imap/Form.vue -->
<template>
  <p class="text-body-2 text-medium-emphasis mb-6">
    {{ t('settings.imap.description') }}
  </p>

  <SettingsCommonItem :title="t('settings.imap.host')">
    <v-text-field v-model="config.imap.host" density="compact" variant="solo" hide-details
      style="min-width: 280px; max-width: 400px" @blur="$emit('save')" />
  </SettingsCommonItem>

  <SettingsCommonItem :title="t('settings.imap.port')">
    <v-number-input v-model="config.imap.port" :min="1" :max="65535" density="compact" variant="solo" hide-details
      style="min-width: 120px; max-width: 240px" controlVariant="stacked" @update:model-value="$emit('save')" />
  </SettingsCommonItem>

  <SettingsCommonItem :title="t('settings.imap.catchAllUsername')">
    <v-text-field v-model="config.imap.catch_all_username" density="compact" variant="solo" hide-details
      style="min-width: 280px; max-width: 400px" @blur="$emit('save')" />
  </SettingsCommonItem>

  <SettingsCommonItem :title="t('settings.imap.catchAllPassword')">
    <v-text-field v-model="config.imap.catch_all_password" :type="showPassword ? 'text' : 'password'" density="compact"
      variant="solo" hide-details style="min-width: 280px; max-width: 400px"
      :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'" @click:append-inner="showPassword = !showPassword"
      @blur="$emit('save')" />
  </SettingsCommonItem>

  <SettingsCommonItem :title="t('settings.imap.mbox')">
    <v-text-field v-model="config.imap.mbox" density="compact" variant="solo" hide-details
      style="min-width: 180px; max-width: 280px" @blur="$emit('save')" />
  </SettingsCommonItem>

  <SettingsCommonItem :title="t('settings.imap.useTls')">
    <v-switch v-model="config.imap.use_tls" hide-details density="compact" color="primary"
      @update:model-value="$emit('save')" />
  </SettingsCommonItem>
</template>

<script setup lang="ts">
defineEmits<{
  save: []
}>()

const { config } = useConfigSync()
const { t } = useI18n()

const showPassword = ref(false)
</script>
