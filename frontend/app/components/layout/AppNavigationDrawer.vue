<!-- app/components/layout/AppNavigationDrawer.vue -->
<template>
  <v-navigation-drawer v-model="model" :rail="isRail" permanent @mouseenter="isRail = false"
    @mouseleave="isRail = true">
    <LayoutAppDrawerNavList />
    <template #append>
      <LayoutAppNavigationDrawerFooter :rail="isRail" />
      <LayoutAppDrawerSettingsLink />
    </template>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
const model = defineModel<boolean>({ default: true })
const isRail = ref(true)
</script>
