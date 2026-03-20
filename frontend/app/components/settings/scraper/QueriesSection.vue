<!-- app/components/settings/scraper/QueriesSection.vue -->
<template>
  <SettingsCommonSection :title="t('settings.scraper.customQueries')">
    <template #header-actions>
      <v-btn icon variant="text" density="compact" @click="helpOpen = true">
        <v-icon icon="mdi-help-circle-outline" />
      </v-btn>
    </template>

    <v-textarea v-model="queryText" :placeholder="t('settings.scraper.customQueriesPlaceholder')" variant="solo"
      density="compact" rows="6" auto-grow hide-details style="max-width: 950px" @blur="save" />
  </SettingsCommonSection>

  <v-dialog v-model="helpOpen" max-width="950">
    <v-card>
      <v-card-title class="d-flex align-center">
        {{ t('settings.scraper.customQueriesHelp.title') }}
        <v-spacer />
        <v-btn icon variant="text" density="compact" @click="helpOpen = false">
          <v-icon icon="mdi-close" />
        </v-btn>
      </v-card-title>
      <v-card-text class="text-medium-emphasis" style="white-space: pre-line">
        {{ t('settings.scraper.customQueriesHelp.body') }}
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
const emit = defineEmits<{ save: [] }>()
const { config } = useConfigSync()
const { t } = useI18n()

const helpOpen = ref(false)

const queryText = computed({
  get: () => (config.scraper.custom_query_strings ?? []).join('\n'),
  set: (val: string) => {
    config.scraper.custom_query_strings = val
      .split('\n')
      .map(l => l.trim())
      .filter(l => l.length > 0)
  },
})

function save() {
  emit('save')
}
</script>
