<!-- frontend/app/app.vue -->
<template>
  <div v-cloak>
    <NuxtLoadingIndicator />
    <NuxtLayout>
      <v-app>
        <NuxtPage />
      </v-app>
    </NuxtLayout>
  </div>
</template>

<style>
[v-cloak] {
  display: none;
}
</style>
