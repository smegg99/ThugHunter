<!-- app/components/layout/AppDrawerNavList.vue -->
<template>
  <v-list nav density="compact" class="px-2">
    <v-list-item v-for="item in navItems" :key="item.to" :to="item.to" :title="t(item.titleKey)" rounded="lg">
      <template #prepend>
        <v-badge v-if="item.badgeColor" dot :color="item.badgeColor" offset-x="-2" offset-y="-2">
          <v-icon :icon="item.icon" />
        </v-badge>
        <v-icon v-else :icon="item.icon" />
      </template>
    </v-list-item>
  </v-list>
</template>

<script setup lang="ts">
const { t } = useI18n()
const { trayPhase } = useScraper()

// 0=idle, 1=starting(blue), 2=running(green), 3=stopping(orange)
const agentsBadgeColor = computed(() => {
  switch (trayPhase.value) {
    case 1: return 'info'
    case 2: return 'success'
    case 3: return 'warning'
    default: return null
  }
})

const navItems = computed(() => [
  { to: '/', icon: 'mdi-home-outline', titleKey: 'routes.home', badgeColor: null as string | null },
  { to: '/browse', icon: 'mdi-compass-outline', titleKey: 'routes.browse', badgeColor: null as string | null },
  { to: '/agents', icon: 'mdi-robot-outline', titleKey: 'routes.agents', badgeColor: agentsBadgeColor.value },
  { to: '/accounts', icon: 'mdi-account-group-outline', titleKey: 'routes.accounts', badgeColor: null as string | null },
])
</script>
