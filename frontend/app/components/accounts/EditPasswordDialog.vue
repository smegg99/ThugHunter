<!-- app/components/accounts/EditPasswordDialog.vue -->
<template>
  <v-dialog v-model="visible" max-width="440" persistent>
    <v-card>
      <v-card-title>{{ t('accounts.dialogs.editPasswordTitle') }}</v-card-title>
      <v-card-text>
        <div class="text-body-2 text-medium-emphasis mb-3">{{ accountEmail }}</div>
        <v-text-field v-model="newPassword" :label="t('accounts.dialogs.newPassword')" type="password" variant="solo"
          density="compact" autofocus />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="close">{{ t('accounts.dialogs.cancel') }}</v-btn>
        <v-btn color="primary" variant="tonal" :disabled="!newPassword" :loading="loading" @click="submit">
          {{ t('accounts.dialogs.save') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
const props = defineProps<{
  accountId: number
  accountEmail: string
}>()
const emit = defineEmits<{ saved: [] }>()

const { t } = useI18n()
const visible = defineModel<boolean>({ default: false })
const newPassword = ref('')
const loading = ref(false)
const log = useLogger()

async function submit() {
  loading.value = true
  try {
    const { updateAccountPassword } = useScraper()
    await updateAccountPassword(props.accountId, newPassword.value)
    emit('saved')
    close()
  } catch (err) {
    log.error('failed to update password', { error: String(err) })
  } finally {
    loading.value = false
  }
}

function close() {
  visible.value = false
  newPassword.value = ''
}
</script>
