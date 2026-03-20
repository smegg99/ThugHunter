<!-- app/components/layout/AppBar.vue -->
<template>
  <v-app-bar flat height="64" class="app-bar-border">
    <div class="logo-container">
      <MiscThugHunterLogo />
    </div>
    <div class="app-bar-center">
      <div id="app-bar-search-slot" />
      <div id="app-bar-actions" class="app-bar-actions" />
    </div>
  </v-app-bar>
</template>

<style scoped>
.app-bar-border {
  border-bottom: thin solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.logo-container {
  height: 48px;
  margin-left: 8px;
  display: flex;
  align-items: center;
  overflow: hidden;
}

.app-bar-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
}

.app-bar-actions {
  position: absolute;
  left: 100%;
  margin-left: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>