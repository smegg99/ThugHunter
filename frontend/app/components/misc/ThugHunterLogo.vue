<!-- app/components/misc/ThugHunterLogo.vue -->
<template>
  <component :is="LogoSvg" class="thughunter-logo" :style="{ '--logo-accent': effectiveAccent }" />
</template>

<script setup lang="ts">
import LogoSvg from '~/assets/thethughunter-logo.svg'

const { effectiveAccent } = useThemeSync()
</script>

<style scoped>
.thughunter-logo {
  display: block;
  height: 100%;
  width: auto;
}

.thughunter-logo :deep(.logo-accent) {
  fill: var(--logo-accent);
}

.thughunter-logo :deep(.logo-light) {
  fill: rgb(var(--v-theme-on-surface));
}

.thughunter-logo :deep(.logo-text-light) {
  fill: rgb(var(--v-theme-on-surface));
}

.thughunter-logo :deep(.logo-text-accent) {
  fill: var(--logo-accent);
}
</style>
