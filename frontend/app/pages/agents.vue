<!-- app/pages/agents.vue -->
<template>
  <v-container>
    <SettingsCommonSection :title="t('agents.run.title')">
      <AgentsRunControl :running="running" :starting="starting" :stopping="stopping" :progress="progress"
        :can-start="hasAccounts" :can-register="imapReady" @start="start" @stop="stop" @refresh="refreshAccounts"
        @register="showRegisterDialog = true" />
    </SettingsCommonSection>

    <Transition name="section-fade" @after-leave="onSectionLeft" @after-enter="onSectionEntered">
      <SettingsCommonSection v-if="visibleSection === 'progress'" key="progress" :title="t('agents.run.progress')">
        <template #header-actions>
          <v-btn v-if="progress?.mode === 'scrape'" icon variant="text" density="compact"
            @click="showQueriesDialog = true">
            <v-icon icon="mdi-information-outline" />
          </v-btn>
        </template>
        <AgentsScrapeProgress :progress="progress!" :stopping="stopping" />
        <AgentsRefreshProgress :progress="progress!" :stopping="stopping" />
        <AgentsRegisterProgress :progress="progress!" :stopping="stopping" />
      </SettingsCommonSection>

      <SettingsCommonSection v-else-if="visibleSection === 'summary'" key="summary" :title="summaryTitle">
        <AgentsRunSummary :summary="summary!" />
      </SettingsCommonSection>
    </Transition>

    <Transition name="section-fade">
      <SettingsCommonSection v-if="showAgents" :title="t('agents.agentList')">
        <AgentsAgentList :agents="agents" />
      </SettingsCommonSection>
    </Transition>

    <AgentsRegisterLaunchDialog v-model="showRegisterDialog" @launch="onRegisterLaunch" />

    <v-dialog v-model="showQueriesDialog" max-width="950">
      <v-card>
        <v-card-title class="d-flex align-center pb-0">
          {{ t('agents.run.queriesDialog.title') }}
          <v-spacer />
          <v-btn icon variant="text" density="compact" @click="showQueriesDialog = false">
            <v-icon icon="mdi-close" />
          </v-btn>
        </v-card-title>
        <v-card-text class="text-body-2 py-0" style="max-height: 480px; overflow-y: auto">
          <div v-if="runQueries.length === 0" class="text-medium-emphasis">
            {{ t('agents.run.queriesDialog.empty') }}
          </div>
          <v-list v-else density="compact">
            <v-list-item v-for="(q, i) in runQueries" :key="i" class="px-0">
              <span class="text-body-2 text-mono text-medium-emphasis">{{ q }}</span>
            </v-list-item>
          </v-list>
        </v-card-text>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { Call } from '@wailsio/runtime'

const SERVICE = 'smegg.me/thughunter/services/scraper.Service'
const { t } = useI18n()
const {
  agents: rawAgents,
  accounts,
  running,
  starting,
  stopping,
  progress,
  summary,
  imapReady,
  start,
  stop,
  refreshAccounts,
  registerAccounts,
  canStartRun,
} = useScraper()

const showRegisterDialog = ref(false)
const showQueriesDialog = ref(false)
const runQueries = ref<string[]>([])
const hasAccounts = ref(true)

watch(showQueriesDialog, async (open) => {
  if (open) {
    try {
      runQueries.value = await Call.ByName(`${SERVICE}.GetQueries`) ?? []
    }
    catch {
      runQueries.value = []
    }
  }
})

async function checkAccounts() {
  hasAccounts.value = await canStartRun()
}

onMounted(checkAccounts)

watch(accounts, checkAccounts)

function onRegisterLaunch(opts: { targetAccounts: number, durationSecs: number }) {
  registerAccounts({
    target_accounts: opts.targetAccounts,
    duration_secs: opts.durationSecs,
  })
}

const agents = computed(() =>
  [...rawAgents.value].sort((a, b) => a.name.localeCompare(b.name)),
)

const summaryTitle = computed(() => {
  if (!summary.value) return t('agents.summary.title')
  const status = summary.value.stopped_early
    ? t('agents.summary.stoppedEarly')
    : t('agents.summary.completedFully')
  return `${t('agents.summary.title')} (${status})`
})

// Section transition state machine
type Section = 'progress' | 'summary' | null

const desiredSection = computed<Section>(() => {
  if (running.value && progress.value) return 'progress'
  if (!running.value && !starting.value && summary.value) return 'summary'
  return null
})

const visibleSection = ref<Section>(desiredSection.value)
let pendingSection: Section = null
let leaving = false
const sectionReady = ref(!!desiredSection.value)

const showAgents = computed(() => agents.value.length > 0 && sectionReady.value)

watch(desiredSection, (next) => {
  if (leaving) {
    // Leave animation in progress - just update the queued target
    pendingSection = next
  } else if (visibleSection.value && visibleSection.value !== next) {
    // Something visible needs to leave first
    pendingSection = next
    visibleSection.value = null
    leaving = true
  } else {
    visibleSection.value = next
  }
})

function onSectionLeft() {
  leaving = false
  visibleSection.value = pendingSection
  pendingSection = null
}

function onSectionEntered() {
  sectionReady.value = true
}

watch(visibleSection, (v) => {
  if (v === null) sectionReady.value = false
})
</script>

<style scoped>
.section-fade-enter-active {
  transition: opacity 0.5s cubic-bezier(0.4, 0, 0.2, 1), transform 0.5s cubic-bezier(0.4, 0, 0.2, 1);
}

.section-fade-leave-active {
  transition: opacity 0.35s cubic-bezier(0.4, 0, 0.2, 1), transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

.section-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.section-fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
