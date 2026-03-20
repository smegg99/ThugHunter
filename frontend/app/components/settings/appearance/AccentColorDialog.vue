<!-- app/components/settings/appearance/AccentColorDialog.vue -->
<template>
  <v-dialog v-model="model" max-width="310">
    <v-card rounded="lg" class="pa-2">
      <v-color-picker v-model="draft" mode="hex" elevation="0" width="100%" hide-inputs />

      <v-card-actions class="pt-0">
        <v-spacer />
        <v-btn variant="text" size="small" @click="cancel">
          {{ t('common.cancel', 'Cancel') }}
        </v-btn>
        <v-btn color="primary" variant="tonal" size="small" @click="confirm">
          OK
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
const model = defineModel<boolean>({ default: false })
const color = defineModel<string>('color', { required: true })

const { t } = useI18n()

const draft = ref(color.value)

watch(model, (open) => {
  if (open) draft.value = color.value
})

function confirm() {
  color.value = draft.value
  model.value = false
}

function cancel() {
  model.value = false
}
</script>
