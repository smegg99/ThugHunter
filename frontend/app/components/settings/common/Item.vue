<!-- app/components/settings/common/Item.vue -->
<template>
  <div class="d-flex align-start justify-space-between ga-6 py-3 mb-2">
    <div class="flex-1-1-0" style="min-width: 0">
      <div class="text-subtitle-2 font-weight-medium">
        {{ title }}
      </div>
      <div v-if="description" class="text-body-2 text-medium-emphasis mt-1">
        {{ description }}
      </div>
    </div>
    <div class="flex-shrink-0">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>
