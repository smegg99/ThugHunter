<!-- app/pages/settings/scanner.vue -->
<template>
  <v-container>
    <SettingsScannerPingSection />
    <SettingsScannerWorkersSection />
    <SettingsScannerTemplatesSection />
    <SettingsScannerScreenshotSection />
  </v-container>
</template>

<script setup lang="ts">
const { config } = useConfigSync()
</script>
