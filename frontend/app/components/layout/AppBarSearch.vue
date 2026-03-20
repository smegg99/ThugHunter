<!-- app/components/layout/AppBarSearch.vue -->
<template>
  <v-text-field v-model="model" :placeholder="placeholder" variant="solo-filled" density="compact" flat hide-details
    single-line prepend-inner-icon="mdi-magnify" clearable style="width: 400px; min-width: 240px; font-size: 14px" @click:clear="model = ''"
    @keydown.escape="model = ''" />
</template>

<script setup lang="ts">
defineProps<{
  placeholder?: string
}>()

const model = defineModel<string>({ default: '' })
</script>

