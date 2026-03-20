<!-- app/components/layout/SecondaryNav.vue -->
<template>
  <div class="d-flex" style="height: 100%;">
    <v-sheet :width="width" color="transparent" class="flex-shrink-0 overflow-y-auto">
      <div v-if="title" class="text-subtitle-2 font-weight-bold text-medium-emphasis pa-4 pb-2">
        {{ title }}
      </div>
      <v-list nav density="compact" bg-color="transparent">
        <v-list-item v-for="item in items" :key="item.to" :to="item.to" :prepend-icon="item.icon" :title="item.title"
          rounded="lg" />
      </v-list>
    </v-sheet>
    <div class="flex-grow-1 overflow-y-auto">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SecondaryNavItem } from '~/types'

withDefaults(defineProps<{
  title?: string
  items: SecondaryNavItem[]
  width?: number | string
}>(), {
  width: 200,
})
</script>
