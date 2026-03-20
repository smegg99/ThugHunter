<!-- app/components/agents/AgentCard.vue -->
<template>
  <div class="agent-row d-flex align-center ga-2 py-2 px-3 rounded">
    <v-icon :icon="display.icon" :color="display.color" size="20" />

    <div class="flex-1-1-0 d-flex flex-column overflow-hidden">
      <div class="d-flex align-center ga-1">
        <span class="text-body-2 font-weight-medium text-no-wrap">{{ agent.name }}</span>
        <span v-if="agent.account" class="text-caption text-medium-emphasis text-truncate">{{ agent.account }}</span>
      </div>
      <span v-if="display.text" class="text-caption" :class="`text-${display.color}`">
        {{ display.text }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AgentInfo } from '~~bindings/smegg.me/thughunter/core/scraper/models.js'

const props = defineProps<{
  agent: AgentInfo
}>()

const { t } = useI18n()

interface StatusSnapshot {
  text: string
  color: string
  icon: string
}

function statusColor(status: string): { color: string; icon: string } {
  switch (status) {
    case 'idle': return { color: 'success', icon: 'mdi-check-circle' }
    case 'busy': return { color: 'warning', icon: 'mdi-progress-clock' }
    case 'waiting': return { color: 'info', icon: 'mdi-timer-sand' }
    case 'error': return { color: 'error', icon: 'mdi-alert-circle' }
    default: return { color: 'grey', icon: 'mdi-circle-off-outline' }
  }
}

function fallbackText(status: string): string {
  const key = `agents.fallback.${status}`
  const translated = t(key)
  return translated !== key ? translated : t('agents.fallback.starting')
}

function snapshot(agent: AgentInfo): StatusSnapshot {
  const { color, icon } = statusColor(agent.status)
  const text = agent.status_text || fallbackText(agent.status)
  return { text, color, icon }
}

const MIN_DISPLAY_MS = 2500
const display = ref<StatusSnapshot>(snapshot(props.agent))
const queue: StatusSnapshot[] = []
let drainTimer: ReturnType<typeof setTimeout> | null = null

function showNext() {
  if (queue.length === 0) {
    drainTimer = null
    return
  }
  display.value = queue.shift()!
  drainTimer = setTimeout(showNext, MIN_DISPLAY_MS)
}

watch(() => props.agent, (agent) => {
  const next = snapshot(agent)
  const last = queue.length > 0 ? queue[queue.length - 1] : display.value
  if (last && next.text === last.text && next.color === last.color) return

  if (!drainTimer) {
    display.value = next
    drainTimer = setTimeout(showNext, MIN_DISPLAY_MS)
  } else {
    queue.push(next)
  }
}, { deep: true })

onUnmounted(() => {
  if (drainTimer) clearTimeout(drainTimer)
})
</script>
