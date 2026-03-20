<!-- app/components/agents/RegisterLaunchDialog.vue -->
<template>
  <v-dialog v-model="model" max-width="420" persistent>
    <v-card>
      <v-card-title>{{ t('agents.register.launchTitle') }}</v-card-title>
      <v-card-text>
        <v-radio-group v-model="limitMode" inline class="mb-4">
          <v-radio :label="t('agents.register.byCount')" value="count" />
          <v-radio :label="t('agents.register.byDuration')" value="duration" />
        </v-radio-group>

        <div v-if="limitMode === 'count'">
          <v-number-input v-model="accountCount" :label="t('agents.register.accountCount')" :min="1" :max="99999"
            variant="solo" density="comfortable" hide-details />
        </div>

        <div v-if="limitMode === 'duration'">
          <v-number-input v-model="durationMinutes" :label="t('agents.register.durationMinutes')" :min="1" :max="1440"
            variant="solo" density="comfortable" hide-details />
        </div>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="model = false">
          {{ t('agents.register.cancel') }}
        </v-btn>
        <v-btn color="primary" variant="flat" @click="launch">
          {{ t('agents.register.start') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
const { t } = useI18n()

const model = defineModel<boolean>({ default: false })
const emit = defineEmits<{
  launch: [opts: { targetAccounts: number, durationSecs: number }]
}>()

const limitMode = ref<'count' | 'duration'>('count')
const accountCount = ref(10)
const durationMinutes = ref(30)

function launch() {
  const opts = {
    targetAccounts: limitMode.value === 'count' ? accountCount.value : 0,
    durationSecs: limitMode.value === 'duration' ? durationMinutes.value * 60 : 0,
  }
  emit('launch', opts)
  model.value = false
}
</script>
