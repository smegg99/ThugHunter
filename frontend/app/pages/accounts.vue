<!-- app/pages/accounts.vue -->
<template>
  <div>
  <Teleport to="#app-bar-search-slot">
    <LayoutAppBarSearch v-model="search" :placeholder="t('accounts.search')" />
  </Teleport>
  <Teleport to="#app-bar-actions">
    <v-btn icon density="comfortable" color="primary" variant="text" @click="showAddDialog = true">
      <v-icon icon="mdi-plus" />
    </v-btn>
    <v-btn icon density="comfortable" variant="text" :loading="refreshLoading" @click="onManualRefresh">
      <v-icon icon="mdi-refresh" />
    </v-btn>
  </Teleport>

  <v-container style="max-width: 920px">
    <AccountsAccountTable :items="rows" :total="totalAccounts" :page="page" :page-size="pageSize" :loading="loading"
      :sort-by="sortBy" :search="search" @update:page="onPageUpdate" @update:page-size="onPageSizeUpdate"
      @update:sort-by="onSortUpdate" @update:search="onSearchUpdate" @edit-password="onEditPassword" @delete="onDelete"
      @add="showAddDialog = true" />

    <!-- Dialogs -->
    <AccountsAddAccountDialog v-model="showAddDialog" @added="reloadPage" />

    <AccountsEditPasswordDialog v-if="editTarget" v-model="showEditDialog" :account-id="editTarget.ID"
      :account-email="editTarget.email" @saved="reloadPage" />

    <AccountsDeleteAccountDialog v-if="deleteTarget" v-model="showDeleteDialog" :account-id="deleteTarget.ID"
      :account-email="deleteTarget.email" @deleted="reloadPage" />
  </v-container>
  </div>
</template>

<script setup lang="ts">
import type { AccountRow } from '~/types'

const { t } = useI18n()
const { loadAccountPage } = useScraper()

const page = ref(1)
const pageSize = ref(12)
const sortBy = ref<{ key: string; order: 'asc' | 'desc' }[]>([{ key: 'credits_amount', order: 'desc' }])
const search = ref('')
const loading = ref(false)
const refreshLoading = ref(false)
const rows = ref<AccountRow[]>([])
const totalAccounts = ref(0)

const showAddDialog = ref(false)
const showEditDialog = ref(false)
const showDeleteDialog = ref(false)
const editTarget = ref<AccountRow | null>(null)
const deleteTarget = ref<AccountRow | null>(null)

let searchTimer: ReturnType<typeof setTimeout> | null = null

const MIN_LOADING_MS = 500

async function onManualRefresh() {
  refreshLoading.value = true
  try {
    await fetchPage()
  } finally {
    refreshLoading.value = false
  }
}

async function fetchPage() {
  loading.value = true
  const start = Date.now()
  try {
    const sort = sortBy.value[0]
    const result = await loadAccountPage(
      page.value,
      pageSize.value,
      sort?.key ?? 'credits_amount',
      sort?.order ?? 'desc',
      search.value,
    )
    rows.value = result.items.map(a => ({ ...a, status: accountStatus(a) }))
    totalAccounts.value = result.total
  }
  finally {
    const elapsed = Date.now() - start
    if (elapsed < MIN_LOADING_MS) {
      await new Promise(r => setTimeout(r, MIN_LOADING_MS - elapsed))
    }
    loading.value = false
  }
}

function onPageUpdate(p: number) {
  page.value = p
  fetchPage()
  scrollToTop()
}

function onPageSizeUpdate(s: number) {
  pageSize.value = s
  page.value = 1
  fetchPage()
  scrollToTop()
}

function onSortUpdate(s: { key: string; order: 'asc' | 'desc' }[]) {
  sortBy.value = s
  page.value = 1
  fetchPage()
  scrollToTop()
}

function onSearchUpdate(q: string) {
  search.value = q
}

watch(search, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    fetchPage()
  }, 300)
})

function onEditPassword(account: AccountRow) {
  editTarget.value = account
  showEditDialog.value = true
}

function onDelete(account: AccountRow) {
  deleteTarget.value = account
  showDeleteDialog.value = true
}

function reloadPage() {
  fetchPage()
}

onMounted(() => {
  fetchPage()
})

onUnmounted(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>
