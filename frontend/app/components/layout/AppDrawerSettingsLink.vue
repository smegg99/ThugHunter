<!-- app/components/layout/AppDrawerSettingsLink.vue -->
<template>
  <v-divider class="mx-3" />
  <v-list nav density="compact" class="px-2">
    <v-list-item to="/settings" prepend-icon="mdi-cog-outline" :title="t('routes.settings')" rounded="lg" />
  </v-list>
</template>

<script setup lang="ts">
const { t } = useI18n()
</script>
