<!-- app/components/accounts/AccountTable.vue -->
<template>
  <div>
    <v-data-table-server :sort-by="sortBy" :headers="headers" :items="items" :items-length="total"
      :items-per-page="pageSize" :items-per-page-text="t('accounts.pagination.rowsPerPage')"
      :items-per-page-options="[12, 24, 36, 48, 60]" :page="page" :loading="loading" density="comfortable"
      @update:page="onPageUpdate" @update:items-per-page="onPageSizeUpdate" @update:sort-by="onSortUpdate">
      <template #item.email="{ item }">
        <div class="d-flex align-center ga-2">
          <v-icon :icon="statusIcon(item)" :style="{ color: statusIconColor(item) }" size="16" />
          <span :style="{ color: emailColor(item), fontStyle: item.user_added ? 'italic' : undefined }">{{ item.email }}</span>
        </div>
      </template>

      <template #item.credits_expire_at="{ item }">
        <span :title="item.credits_expire_at ?? ''">
          {{ item.credits_expire_at ? timeAgo(item.credits_expire_at) : '-' }}
        </span>
      </template>

      <template #item.refreshed_credits_at="{ item }">
        <span :title="item.refreshed_credits_at ?? ''">
          {{ item.refreshed_credits_at ? timeAgo(item.refreshed_credits_at) : '-' }}
        </span>
      </template>

      <template #item.actions="{ item }">
        <div class="d-flex ga-1 justify-end">
          <v-btn icon variant="text" size="x-small" color="warning" @click="emit('editPassword', item)">
            <v-icon icon="mdi-key-variant" size="16" />
          </v-btn>
          <v-btn icon variant="text" size="x-small" color="error" @click="emit('delete', item)">
            <v-icon icon="mdi-delete-outline" size="16" />
          </v-btn>
        </div>
      </template>

      <template #no-data>
        <div class="d-flex flex-column align-center text-center py-12 px-4">
          <v-icon icon="mdi-account-off-outline" size="40" color="medium-emphasis" class="mb-2" />
          <div class="text-body-2 text-medium-emphasis">{{ t('accounts.empty') }}</div>
        </div>
      </template>
    </v-data-table-server>
  </div>
</template>

<script setup lang="ts">
import type { AccountRow, SortItem } from '~/types'

defineProps<{
  items: AccountRow[]
  total: number
  page: number
  pageSize: number
  loading: boolean
  sortBy: SortItem[]
  search: string
}>()

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [size: number]
  'update:sortBy': [sortBy: SortItem[]]
  'update:search': [search: string]
  editPassword: [account: AccountRow]
  delete: [account: AccountRow]
  add: []
}>()

const { t } = useI18n()
const { timeAgo } = useTimeAgo()

const headers = computed(() => [
  { title: t('accounts.columns.email'), key: 'email', sortable: true, width: '150px' },
  { title: t('accounts.columns.credits'), key: 'credits_amount', sortable: true, align: 'end' as const },
  { title: t('accounts.columns.creditsExpire'), key: 'credits_expire_at', sortable: true },
  { title: t('accounts.columns.refreshed'), key: 'refreshed_credits_at', sortable: true },
  { title: t('accounts.columns.actions'), key: 'actions', sortable: false, align: 'end' as const, width: '90px' },
])

function emailColor(item: AccountRow): string | undefined {
  // if (item.user_added) return undefined
  switch (item.status) {
    case 'active': return 'rgb(var(--v-theme-success))'
    case 'expired': return 'rgb(var(--v-theme-error))'
    default: return 'rgb(var(--v-theme-warning))'
  }
}

function statusIcon(item: AccountRow): string {
  switch (item.status) {
    case 'active': return 'mdi-check-circle'
    case 'expired': return 'mdi-close-circle'
    default: return 'mdi-alert-circle'
  }
}

function statusIconColor(item: AccountRow): string {
  switch (item.status) {
    case 'active': return 'rgb(var(--v-theme-success))'
    case 'expired': return 'rgb(var(--v-theme-error))'
    default: return 'rgb(var(--v-theme-warning))'
  }
}

function onPageUpdate(p: number) {
  emit('update:page', p)
}

function onPageSizeUpdate(s: number) {
  emit('update:pageSize', s)
}

function onSortUpdate(s: SortItem[]) {
  emit('update:sortBy', s)
}
</script>

