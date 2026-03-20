<!-- app/components/layout/AppNavigationDrawerFooter.vue -->
<template>
  <div class="drawer-footer mt-auto flex-shrink-0 text-center">
    <Transition name="fade">
      <div v-show="!rail" class="d-flex align-center justify-center ga-3 pa-4">
        <a href="#" class="github-link d-inline-flex align-center ga-1" @click.prevent="openGitHub">
          <v-icon icon="mdi-github" size="16" />
          <span class="text-caption font-weight-medium" style="letter-spacing: 0.3px; line-height: 16px">GitHub</span>
        </a>

        <div class="vertical-separator" />

        <span class="text-caption text-primary font-weight-bold" style="letter-spacing: 0.5px">v2.0.0</span>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import * as ProgramService from '~~bindings/smegg.me/thughunter/services/program/service.js'

defineProps<{ rail: boolean }>()

function openGitHub() {
  ProgramService.OpenURL('https://github.com/smegg99/ThugHunter')
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease-in-out;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.github-link {
  color: rgba(var(--v-theme-on-background), 0.5);
  text-decoration: none;
  transition: color 0.2s ease-in-out, transform 0.2s ease-in-out;
}

.github-link:hover {
  color: rgb(var(--v-theme-primary));
  transform: translateY(-1px);
}

.vertical-separator {
  width: 1px;
  height: 14px;
  background: rgba(var(--v-theme-primary), 0.3);
  flex-shrink: 0;
}
</style>
