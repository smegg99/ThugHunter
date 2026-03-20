<!-- app/components/settings/common/PageHeader.vue -->
<template>
  <div class="mb-4">
    <h1 class="text-h5 font-weight-bold">
      {{ title }}
    </h1>
    <p v-if="description" class="text-body-2 text-medium-emphasis mt-1">
      {{ description }}
    </p>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>
