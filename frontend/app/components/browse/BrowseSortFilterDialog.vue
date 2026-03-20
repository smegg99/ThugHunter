<!-- app/components/browse/BrowseSortFilterDialog.vue -->
<template>
  <v-dialog v-model="model" max-width="520" scrollable>
    <v-card rounded="lg">
      <v-card-title class="d-flex align-center ga-2 px-4 pt-4 pb-0">
        <v-icon icon="mdi-tune-variant" size="20" />
        {{ t('browse.sortFilter.title') }}
        <v-spacer />
        <v-btn icon variant="text" density="compact" size="small" @click="model = false">
          <v-icon icon="mdi-close" size="18" />
        </v-btn>
      </v-card-title>

      <v-card-text class="px-4 pt-3 pb-1">
        <!-- ───── Sort ───── -->
        <div class="text-caption text-medium-emphasis font-weight-bold mb-2">
          {{ t('browse.sortFilter.sortBy') }}
        </div>
        <div class="d-flex ga-2 align-stretch">
          <v-select v-model="draft.sortBy" :items="sortFieldItems" item-title="label" item-value="value"
            density="compact" variant="solo" hide-details class="flex-grow-1" style="min-width: 0" />
          <v-btn-toggle v-if="showNoAuthOnly" v-model="draft.screenshotFilter" mandatory density="compact" divided
            color="primary" style="height: 40px; flex-shrink: 0">
            <v-btn value="has" :title="t('browse.sortFilter.screenshotHas')" size="small">
              <v-icon icon="mdi-image-check" size="18" class="mr-1" />
            </v-btn>
            <v-btn value="all" :title="t('browse.sortFilter.screenshotAll')" size="small">
              <v-icon icon="mdi-image-multiple" size="18" class="mr-1" />
            </v-btn>
            <v-btn value="none" :title="t('browse.sortFilter.screenshotNone')" size="small">
              <v-icon icon="mdi-image-off" size="18" class="mr-1" />
            </v-btn>
          </v-btn-toggle>

        </div>

        <!-- ───── Per page + Screenshot filter ───── -->
        <div class="text-caption text-medium-emphasis font-weight-bold mt-4 mb-2">
          {{ t('browse.sortFilter.perPage') }}
        </div>
        <div class="d-flex ga-3 align-stretch">
          <v-select v-model="draft.pageSize" :items="pageSizeOptions" density="compact" variant="solo" hide-details
            class="flex-grow-1" style="min-width: 0" />
          <v-btn-toggle v-model="draft.sortOrder" mandatory density="compact" divided color="primary"
            style="height: 40px; flex-shrink: 0">
            <v-btn value="asc" :title="t('browse.sortFilter.asc')" size="small">
              <v-icon icon="mdi-sort-ascending" size="18" />
            </v-btn>
            <v-btn value="desc" :title="t('browse.sortFilter.desc')" size="small">
              <v-icon icon="mdi-sort-descending" size="18" />
            </v-btn>
          </v-btn-toggle>
          <v-btn-toggle v-if="showNoAuthOnly" v-model="draft.authFilter" mandatory density="compact" divided
            color="error" style="height: 40px; flex-shrink: 0">
            <v-btn value="open" :title="t('browse.sortFilter.noAuthOnly')" size="small">
              <v-icon icon="mdi-lock-open-variant" size="18" />
            </v-btn>
            <v-btn value="all" :title="t('browse.sortFilter.authAll')" size="small">
              <v-icon icon="mdi-lock-open-check-outline" size="18" />
            </v-btn>
            <v-btn value="closed" :title="t('browse.sortFilter.authOnly')" size="small">
              <v-icon icon="mdi-lock" size="18" />
            </v-btn>
          </v-btn-toggle>
        </div>

        <v-divider class="my-4" />

        <!-- ───── Filters ───── -->
        <div class="text-caption text-medium-emphasis font-weight-bold mb-2">
          {{ t('browse.sortFilter.countries') }}
        </div>
        <v-combobox v-model="draft.countries" :items="filterOptions.countries" multiple chips closable-chips
          :placeholder="t('browse.sortFilter.countriesHint')" density="compact" variant="solo" hide-details>
          <template #chip="{ props: chipProps, item }">
            <v-chip v-bind="chipProps" size="small" variant="tonal" color="primary">
              <span class="mr-1">{{ countryFlag(item) }}</span>
              {{ item }}
            </v-chip>
          </template>
        </v-combobox>

        <div class="text-caption text-medium-emphasis font-weight-bold mt-4 mb-2">
          {{ t('browse.sortFilter.labels') }}
        </div>
        <v-combobox v-model="draft.labels" :items="filterOptions.labels" multiple chips closable-chips
          :placeholder="t('browse.sortFilter.labelsHint')" density="compact" variant="solo" hide-details>
          <template #chip="{ props: chipProps, item }">
            <v-chip v-bind="chipProps" size="small" variant="tonal" color="secondary">
              {{ item }}
            </v-chip>
          </template>
        </v-combobox>

        <div class="text-caption text-medium-emphasis font-weight-bold mt-4 mb-2">
          {{ t('browse.sortFilter.hardware') }}
        </div>
        <v-text-field v-model="draft.hardware" :placeholder="t('browse.sortFilter.hardwareHint')" density="compact"
          variant="solo" hide-details clearable />
      </v-card-text>

      <v-card-actions class="px-4 pb-3 pt-2">
        <v-btn variant="text" size="small" @click="onReset">
          {{ t('browse.sortFilter.reset') }}
        </v-btn>
        <v-spacer />
        <v-btn color="primary" variant="flat" size="small" @click="onApply">
          {{ t('browse.sortFilter.apply') }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import type { SortFilterParams, FilterOptions } from '~/types/scanner'
import { countryFlag } from '~/utils/country'

const model = defineModel<boolean>({ default: false })

const props = defineProps<{
  current: SortFilterParams
  sortFields: { value: string, label: string }[]
  filterOptions: FilterOptions
  showNoAuthOnly?: boolean
}>()

const emit = defineEmits<{
  apply: [params: SortFilterParams]
}>()

const { t } = useI18n()

const draft = reactive<SortFilterParams>({
  sortBy: props.current.sortBy,
  sortOrder: props.current.sortOrder,
  countries: [...props.current.countries],
  labels: [...props.current.labels],
  hardware: props.current.hardware,
  pageSize: props.current.pageSize,
  authFilter: props.current.authFilter,
  screenshotFilter: props.current.screenshotFilter,
})

const sortFieldItems = computed(() => props.sortFields)
const pageSizeOptions = [24, 48, 96]



watch(model, (open) => {
  if (open) {
    draft.sortBy = props.current.sortBy
    draft.sortOrder = props.current.sortOrder
    draft.countries = [...props.current.countries]
    draft.labels = [...props.current.labels]
    draft.hardware = props.current.hardware
    draft.pageSize = props.current.pageSize
    draft.authFilter = props.current.authFilter
    draft.screenshotFilter = props.current.screenshotFilter
  }
})

function onApply() {
  emit('apply', {
    sortBy: draft.sortBy,
    sortOrder: draft.sortOrder,
    countries: [...draft.countries],
    labels: [...draft.labels],
    hardware: draft.hardware,
    pageSize: draft.pageSize,
    authFilter: draft.authFilter,
    screenshotFilter: draft.screenshotFilter,
  })
  model.value = false
}

function onReset() {
  draft.sortBy = props.sortFields[0]?.value ?? 'ping_ms'
  draft.sortOrder = 'asc'
  draft.countries = []
  draft.labels = []
  draft.hardware = ''
  draft.pageSize = 24
  draft.authFilter = 'all'
  draft.screenshotFilter = 'all'
}
</script>


