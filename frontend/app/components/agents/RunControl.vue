<!-- app/components/agents/RunControl.vue -->
<template>
  <div class="d-flex justify-center align-center">
    <Transition name="run-btn" mode="out-in">
      <div v-if="!running" key="idle" class="d-flex justify-center align-center ga-3">
        <v-btn color="info" variant="tonal" size="large" prepend-icon="mdi-refresh" rounded="lg" :loading="starting"
          :disabled="starting || !canStart" min-width="160" @click="$emit('refresh')">
          {{ t('agents.run.refreshAccounts') }}
        </v-btn>
        <v-btn color="success" variant="tonal" size="large" prepend-icon="mdi-play" rounded="lg" :loading="starting"
          :disabled="starting || !canStart" min-width="160" @click="$emit('start')">
          {{ t('agents.run.startScraping') }}
        </v-btn>
        <v-btn color="warning" variant="tonal" size="large" prepend-icon="mdi-account-plus" rounded="lg"
          :loading="starting" :disabled="starting || !canRegister" min-width="160" @click="$emit('register')">
          {{ t('agents.register.registerAccounts') }}
        </v-btn>
      </div>
      <div v-else key="running" class="d-flex justify-center align-center">
        <v-btn color="error" variant="tonal" size="large" prepend-icon="mdi-stop" rounded="lg" :loading="stopping"
          :disabled="stopping" width="160" @click="$emit('stop')">
          {{ t('agents.run.stop') }}
        </v-btn>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import type { RunProgress } from '~~bindings/smegg.me/thughunter/core/scraper/models.js'

const { t } = useI18n()

defineProps<{
  running: boolean
  starting: boolean
  stopping: boolean
  progress: RunProgress | null
  canStart: boolean
  canRegister: boolean
}>()

defineEmits<{
  start: []
  stop: []
  refresh: []
  register: []
}>()
</script>

<style scoped>
.run-btn-enter-active,
.run-btn-leave-active {
  transition: opacity 0.2s ease;
}

.run-btn-enter-from,
.run-btn-leave-to {
  opacity: 0;
}
</style>
