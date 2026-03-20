<!-- app/components/browse/BrowseVNCList.vue -->
<template>
  <div>
    <Transition name="page" mode="out-in">
      <div v-if="loading" key="loading" class="d-flex justify-center pa-8">
        <v-progress-circular indeterminate color="primary" />
      </div>

      <div v-else-if="!items.length && !frozen" key="empty" class="text-center text-medium-emphasis pa-8">
        <v-icon icon="mdi-monitor-off" size="48" class="mb-3 d-block mx-auto opacity-40" />
        {{ t('browse.noServices') }}
      </div>

      <div v-else key="content">
        <div style="display: grid; grid-template-columns: repeat(auto-fill, 340px); grid-auto-rows: auto; gap: 10px; width: 100%; justify-content: center">
          <template v-if="frozen">
            <BrowseVNCCardSkeleton v-for="i in pageSize" :key="`sk-${i}`" />
          </template>
          <template v-else>
            <BrowseVNCCard v-for="svc in items" :key="svc.ID" :service="svc"
              :screenshot-src="screenshotMap.get(svc.ID) ?? ''" @favorite-changed="onFavoriteChanged" />
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
import type { VNCItem, SortFilterParams, ScreenshotResult } from '~/types/scanner'

const props = defineProps<{
  search?: string
  sortFilter: SortFilterParams
  savedPage?: number
  frozen?: boolean
}>()

const emit = defineEmits<{
  'update:page': [page: number]
}>()

const { t } = useI18n()
const { listVNC, refreshScreenshots } = useScanner()

const loading = ref(true)
const items = ref<VNCItem[]>([])
const total = ref(0)
const currentPage = ref(props.savedPage ?? 1)
const screenshotMap = ref<Map<number, string>>(new Map())

const pageSize = computed(() => props.sortFilter.pageSize)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

let prefetchAbort: AbortController | null = null
let searchTimer: ReturnType<typeof setTimeout> | null = null
let fetchInProgress = false
let alive = true

async function fetchPage() {
  if (fetchInProgress || !alive) return
  fetchInProgress = true
  // Cancel any in-flight screenshot loads from a previous fetch.
  if (prefetchAbort) prefetchAbort.abort()
  loading.value = true
  try {
    const sf = props.sortFilter
    const result = await listVNC(currentPage.value, pageSize.value, sf.sortBy, sf.sortOrder, props.search ?? '', sf.countries, sf.labels, sf.hardware, sf.authFilter, sf.screenshotFilter)
    items.value = result.items
    total.value = result.total

    // Clamp page if it exceeds total after data change
    if (currentPage.value > totalPages.value && totalPages.value > 0) {
      currentPage.value = totalPages.value
      emit('update:page', currentPage.value)
    }

    // Load cached screenshots from the DB for items that already have one.
    // No on-demand capture - only IDs with has_screenshot=true are sent,
    // so the backend returns stored data without spawning VNC connections.
    const cached = new Map<number, string>()
    const cachedIds = result.items.filter(s => s.has_screenshot).map(s => s.ID)
    if (cachedIds.length) {
      try {
        const results = await refreshScreenshots(cachedIds)
        for (const r of results) {
          if (r.screenshot) {
            cached.set(r.id, `data:${detectMime(r.screenshot)};base64,${r.screenshot}`)
          }
        }
      }
      catch { /* ignore */ }
    }
    screenshotMap.value = cached
  }
  finally {
    loading.value = false
    fetchInProgress = false
  }
}

async function loadScreenshots() {
  // Cancel any in-flight request.
  if (prefetchAbort) prefetchAbort.abort()
  prefetchAbort = new AbortController()
  const signal = prefetchAbort.signal

  // Collect IDs: current page items that need fresh screenshots.
  const currentIds = items.value.filter(s => s.no_auth).map(s => s.ID)
  if (!currentIds.length || signal.aborted) return

  // Fire batches in parallel so screenshots stream in progressively.
  const BATCH = 8
  const batches: number[][] = []
  for (let i = 0; i < currentIds.length; i += BATCH) {
    batches.push(currentIds.slice(i, i + BATCH))
  }

  await Promise.all(
    batches.map(async (batch) => {
      if (signal.aborted) return
      try {
        const results = await refreshScreenshots(batch)
        if (signal.aborted) return
        applyScreenshots(results)
      }
      catch { /* ignore */ }
    }),
  )
}

function detectMime(b64: string): string {
  if (b64.startsWith('/9j/')) return 'image/jpeg'
  if (b64.startsWith('iVBOR')) return 'image/png'
  return 'image/jpeg'
}

function applyScreenshots(results: ScreenshotResult[]) {
  const newMap = new Map(screenshotMap.value)
  for (const r of results) {
    if (r.screenshot) {
      newMap.set(r.id, `data:${detectMime(r.screenshot)};base64,${r.screenshot}`)
    }
  }
  screenshotMap.value = newMap
}

function onPageChange(page: number) {
  currentPage.value = page
  emit('update:page', page)
  if (!props.frozen) fetchPage()
  scrollToTop()
}

defineExpose({ reload: fetchPage })

function onFavoriteChanged(id: number, value: boolean) {
  const item = items.value.find(s => s.ID === id)
  if (item) item.is_favorite = value
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

onMounted(fetchPage)

onUnmounted(() => {
  alive = false
  if (prefetchAbort) prefetchAbort.abort()
  if (searchTimer) clearTimeout(searchTimer)
})
</script>

