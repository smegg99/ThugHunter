<!-- app/pages/settings/scraper.vue -->
<template>
  <v-container>
    <SettingsScraperBrowserSection />
    <SettingsScraperAgentsSection />
    <SettingsScraperTemplatesSection />
    <SettingsScraperQueriesSection />
  </v-container>
</template>

<script setup lang="ts">
const { config } = useConfigSync()
</script>
