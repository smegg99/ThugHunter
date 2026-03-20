<!-- app/components/browse/BrowseVNCCardSkeleton.vue -->
<template>
  <v-card variant="tonal" rounded="t-lg" class="h-100">
    <!-- 16:9 image placeholder -->
    <div class="vnc-thumb bg-surface-light d-flex align-center justify-center">
      <span class="bone" style="width: 100%; height: 100%" />
    </div>

    <!-- Body -->
    <v-card-text class="d-flex flex-column ga-1 pa-3 flex-grow-1">
      <!-- Header: IP:Port + lock icon + latency -->
      <div class="d-flex align-center justify-space-between ga-2" style="min-height: 20px">
        <span class="bone" style="width: 55%; height: 14px" />
        <div class="d-flex align-center ga-1 flex-shrink-0">
          <span class="bone rounded-circle" style="width: 14px; height: 14px" />
          <span class="bone rounded-pill" style="width: 48px; height: 20px" />
        </div>
      </div>

      <!-- Host info: OS + hardware -->
      <div class="d-flex align-center ga-1" style="min-height: 18px">
        <span class="bone" style="width: 40%; height: 12px" />
        <span class="bone" style="width: 30%; height: 12px" />
      </div>

      <!-- Country + city -->
      <div class="d-flex align-center ga-1" style="min-height: 18px">
        <span class="bone rounded-circle" style="width: 16px; height: 16px; flex-shrink: 0" />
        <span class="bone" style="width: 50%; height: 12px" />
      </div>
    </v-card-text>
  </v-card>
</template>

<style scoped>
.vnc-thumb {
  aspect-ratio: 16 / 9;
  width: 100%;
  flex-shrink: 0;
  overflow: hidden;
}
</style>
