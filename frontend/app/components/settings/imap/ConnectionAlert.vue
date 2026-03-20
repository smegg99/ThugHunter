<!-- app/components/settings/imap/ConnectionAlert.vue -->
<template>
  <v-fade-transition mode="out-in">
    <v-alert v-if="verifyError" key="error" type="error" variant="tonal" density="compact" icon="mdi-alert-circle"
      class="mb-6">
      {{ verifyError }}
    </v-alert>
    <v-alert v-else-if="verified" key="success" type="success" variant="tonal" density="compact" icon="mdi-check-circle"
      class="mb-6">
      {{ t('settings.imap.verifySuccess') }}
    </v-alert>
    <v-alert v-else-if="verifying" key="verifying" type="info" variant="tonal" density="compact" class="mb-6">
      <template #prepend>
        <v-progress-circular indeterminate size="20" width="2" />
      </template>
      {{ t('settings.imap.verifying') }}
    </v-alert>
  </v-fade-transition>
</template>

<script setup lang="ts">
defineProps<{
  verifying: boolean
  verified: boolean
  verifyError: string
}>()

const { t } = useI18n()
</script>
