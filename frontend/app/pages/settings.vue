<!-- app/pages/settings.vue -->
<template>
  <LayoutSecondaryNav :items="settingsNavItems">
    <v-container style="max-width: 720px">
      <NuxtPage />
    </v-container>
  </LayoutSecondaryNav>
</template>

<script setup lang="ts">
const { t } = useI18n()

const settingsNavItems = computed(() => [
  { to: '/settings/preferences', icon: 'mdi-palette-outline', title: t('settings.nav.preferences') },
  { to: '/settings/imap', icon: 'mdi-email-outline', title: t('settings.nav.imap') },
  { to: '/settings/scraper', icon: 'mdi-spider-web', title: t('settings.nav.scraper') },
  { to: '/settings/scanner', icon: 'mdi-radar', title: t('settings.nav.scanner') },
  // { to: '/settings/data-export', icon: 'mdi-database-export-outline', title: t('settings.nav.dataExport') },
  // { to: '/settings/data-import', icon: 'mdi-database-import-outline', title: t('settings.nav.dataImport') },
])
</script>
