<!-- app/components/accounts/DeleteAccountDialog.vue -->
<template>
  <v-dialog v-model="visible" max-width="400">
    <v-card>
      <v-card-title>{{ t('accounts.dialogs.deleteTitle') }}</v-card-title>
      <v-card-text>
        <div class="text-body-1">{{ t('accounts.dialogs.deleteConfirm') }}</div>
        <div class="text-body-2 font-weight-medium mt-2">{{ accountEmail }}</div>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="close">{{ t('accounts.dialogs.cancel') }}</v-btn>
        <v-btn color="error" variant="tonal" :loading="loading" @click="submit">
          {{ t('accounts.dialogs.delete') }}
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
const emit = defineEmits<{ deleted: [] }>()

const { t } = useI18n()
const visible = defineModel<boolean>({ default: false })
const loading = ref(false)
const log = useLogger()

async function submit() {
  loading.value = true
  try {
    const { deleteAccount } = useScraper()
    await deleteAccount(props.accountId)
    emit('deleted')
    close()
  } catch (err) {
    log.error('failed to delete account', { error: String(err) })
  } finally {
    loading.value = false
  }
}

function close() {
  visible.value = false
}
</script>
