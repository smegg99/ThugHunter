<!-- app/components/browse/BrowseHostCardSkeleton.vue -->
<template>
  <v-card variant="tonal" rounded="lg" class="h-100">
    <v-card-text class="d-flex flex-column ga-1 pa-3">
      <!-- Header row: IP + ping badge -->
      <div class="d-flex align-center justify-space-between ga-2">
        <span class="bone" style="width: 55%; height: 14px" />
        <span class="bone rounded-pill" style="width: 48px; height: 20px" />
      </div>

      <!-- Country flag + name -->
      <div class="d-flex align-center ga-1">
        <span class="bone rounded-circle" style="width: 16px; height: 16px; flex-shrink: 0" />
        <span class="bone" style="width: 65%; height: 12px" />
      </div>

      <!-- OS -->
      <div class="d-flex align-center ga-1">
        <span class="bone rounded-circle" style="width: 11px; height: 11px; flex-shrink: 0" />
        <span class="bone" style="width: 50%; height: 12px" />
      </div>

      <!-- Hardware -->
      <div class="d-flex align-center ga-1">
        <span class="bone rounded-circle" style="width: 11px; height: 11px; flex-shrink: 0" />
        <span class="bone" style="width: 40%; height: 12px" />
      </div>

      <!-- Service chips -->
      <div class="d-flex ga-1 mt-auto">
        <span class="bone rounded-pill" style="width: 36px; height: 18px" />
        <span class="bone rounded-pill" style="width: 36px; height: 18px" />
      </div>
    </v-card-text>
  </v-card>
</template>

