<!-- app/components/agents/AgentList.vue -->
<template>
  <div>
    <Transition name="fade">
      <v-row>
        <v-col v-for="(agent, index) in agents" :key="agent.name" cols="12" sm="6" lg="4" xl="3">
          <AgentsAgentCard :agent="agent" :data-index="index" />
        </v-col>
      </v-row>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import type { AgentInfo } from '~~bindings/smegg.me/thughunter/core/scraper/models.js'

const props = defineProps<{ agents: AgentInfo[] }>()
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
