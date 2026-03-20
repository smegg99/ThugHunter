<!-- app/pages/settings/imap.vue -->
<template>
  <v-container>
    <SettingsCommonSection :title="t('settings.imap.title')">
      <SettingsImapConnectionAlert :verifying="verifying" :verified="verified" :verify-error="verifyError" />
      <SettingsImapForm @save="onSave" />
    </SettingsCommonSection>
  </v-container>
</template>

<script setup lang="ts">
const { config } = useConfigSync()
const { t } = useI18n()
const { verifying, verified, verifyError, verifyConnection, onSave } = useImapVerify(config)

onMounted(() => {
  verifyConnection()
})
</script>
