<!-- app/components/settings/common/Section.vue -->
<template>
  <section class="mb-8">
    <div class="d-flex align-center mb-1">
      <h2 class="text-overline text-medium-emphasis font-weight-bold ma-0">
        {{ title }}
      </h2>
      <v-spacer />
      <slot name="header-actions" />
    </div>
    <v-divider class="mb-2" />
    <slot />
  </section>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>
