<!-- app/components/accounts/AddAccountDialog.vue -->
<template>
  <v-dialog v-model="visible" max-width="440" persistent>
    <v-card>
      <v-card-title>{{ t('accounts.dialogs.addTitle') }}</v-card-title>
      <v-card-text>
        <v-text-field v-model="email" :label="t('accounts.dialogs.email')" :error-messages="emailError" type="email"
          variant="solo" density="compact" class="mb-3" autofocus :rules="[emailRule]"
          @update:model-value="errorMsg = ''" />
        <v-text-field v-model="password" :label="t('accounts.dialogs.password')" type="password" variant="solo"
          density="compact" />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="close">{{ t('accounts.dialogs.cancel') }}</v-btn>
        <v-btn color="primary" variant="tonal" :disabled="!canSubmit" :loading="loading" @click="submit">
          {{ t('accounts.dialogs.add') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
const emit = defineEmits<{ added: [] }>()

const { t } = useI18n()
const visible = defineModel<boolean>({ default: false })
const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function emailRule(v: string): true | string {
  if (!v.trim()) return true
  return emailPattern.test(v.trim()) || t('accounts.dialogs.invalidEmail')
}

const emailValid = computed(() => emailPattern.test(email.value.trim()))
const canSubmit = computed(() => emailValid.value && password.value !== '')

const emailError = computed(() => errorMsg.value ? [errorMsg.value] : [])

async function submit() {
  loading.value = true
  errorMsg.value = ''
  try {
    const { addAccount } = useScraper()
    await addAccount(email.value.trim(), password.value)
    emit('added')
    close()
  } catch (err: any) {
    const msg = String(err?.message ?? err ?? '')
    if (msg.includes('already exists')) {
      errorMsg.value = t('accounts.dialogs.emailExists')
    } else {
      errorMsg.value = msg
    }
  } finally {
    loading.value = false
  }
}

function close() {
  visible.value = false
  email.value = ''
  password.value = ''
  errorMsg.value = ''
}
</script>
