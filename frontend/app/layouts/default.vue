<!-- app/layouts/default.vue -->
<template>
  <v-app>
    <LayoutAppBar />
    <LayoutAppNavigationDrawer v-model="drawer" />
    <v-main class="d-flex flex-column">
      <div class="main-scroll-container">
        <slot />
      </div>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { useConfigSync } from '~/composables/useConfigSync'
import { useThemeSync } from '~/composables/useThemeSync'

useConfigSync()
useThemeSync()

const drawer = ref(true)
</script>

<style scoped>
.main-scroll-container {
  flex: 1 1 0;
  overflow-y: auto;
  min-height: 0;
}
</style>