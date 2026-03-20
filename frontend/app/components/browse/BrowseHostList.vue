<!-- app/components/browse/BrowseHostList.vue -->
<template>
  <div>
    <Transition name="page" mode="out-in">
      <div v-if="loading" key="loading" class="d-flex justify-center pa-8">
        <v-progress-circular indeterminate color="primary" />
      </div>

      <div v-else-if="!items.length && !frozen" key="empty" class="text-center text-medium-emphasis pa-8">
        <v-icon icon="mdi-server-off" size="48" class="mb-3 d-block mx-auto opacity-40" />
        {{ t('browse.noHosts') }}
      </div>

      <div v-else key="content">
        <div style="display: grid; grid-template-columns: repeat(auto-fill, 340px); grid-auto-rows: 140px; gap: 10px; width: 100%; justify-content: center">
          <template v-if="frozen">
            <BrowseHostCardSkeleton v-for="i in pageSize" :key="`sk-${i}`" />
          </template>
          <template v-else>
            <BrowseHostCard v-for="host in items" :key="host.ID" :host="host" @favorite-changed="onFavoriteChanged" />
          </template>
        </div>

        <div class="d-flex align-center justify-center mt-4 ga-4 flex-wrap">
          <v-pagination v-if="totalPages > 1" v-model="currentPage" :length="totalPages" :total-visible="7"
            show-first-last-page density="comfortable" rounded="lg" :disabled="frozen"
            @update:model-value="onPageChange" />
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import type { HostItem, SortFilterParams } from '~/types/scanner'

const props = defineProps<{
  search: string
  sortFilter: SortFilterParams
  savedPage?: number
  frozen?: boolean
}>()

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const { t } = useI18n()
const { listHosts } = useScanner()

const loading = ref(true)
const items = ref<HostItem[]>([])
const total = ref(0)
const currentPage = ref(props.savedPage ?? 1)

const pageSize = computed(() => props.sortFilter.pageSize)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

let searchTimer: ReturnType<typeof setTimeout> | null = null

async function fetchPage() {
  loading.value = true
  try {
    const sf = props.sortFilter
    const result = await listHosts(currentPage.value, pageSize.value, sf.sortBy, sf.sortOrder, props.search, sf.countries, sf.labels, sf.hardware)
    items.value = result.items
    total.value = result.total
    // Clamp page if it exceeds total after data change
    if (currentPage.value > totalPages.value && totalPages.value > 0) {
      currentPage.value = totalPages.value
      emit('update:page', currentPage.value)
    }
  }
  finally {
    loading.value = false
  }
}

function onPageChange(page: number) {
  currentPage.value = page
  emit('update:page', page)
  if (!props.frozen) fetchPage()
  scrollToTop()
}

watch(() => props.search, () => {
  if (props.frozen) return
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    emit('update:page', 1)
    fetchPage()
  }, 350)
})

watch(() => props.sortFilter, () => {
  if (props.frozen) return
  currentPage.value = 1
  emit('update:page', 1)
  fetchPage()
}, { deep: true })

// Expose reload so parent can trigger a refresh after scan completes.
defineExpose({ reload: fetchPage })

function onFavoriteChanged(id: number, value: boolean) {
  const item = items.value.find(h => h.ID === id)
  if (item) item.is_favorite = value
}

onMounted(fetchPage)
</script>

