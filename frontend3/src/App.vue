<template>
  <div id="app">
    <!-- App-level loading screen while assets initialize -->
    <AppLoader v-if="!isAppReady" :progress="loadingProgress" :status="loadingStatus" :error="initError" />

    <!-- Main app content (only shown after assets are loaded) -->
    <div v-else class="app-content">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore, useUIStore } from '@/stores'
import { loadAllAssets } from '@/utils/imageLoader'
import AppLoader from '@/components/AppLoader.vue'

const authStore = useAuthStore()
const uiStore = useUIStore()

// App initialization state
const isAppReady = ref(false)
const loadingProgress = ref(0)
const initError = ref<string | null>(null)

const loadingStatus = computed(() => {
  if (initError.value) return 'error'
  if (isAppReady.value) return 'complete'
  return 'loading'
})

onMounted(async () => {
  // Initialize and preload all game assets from API
  try {
    // Load all assets with progress tracking
    // This fetches asset URLs from API and preloads all images into memory
    await loadAllAssets((loaded, total) => {
      // Map loading progress to 10-90% range (reserve 90-100% for auth init)
      const assetProgress = total > 0 ? (loaded / total) * 80 : 0
      loadingProgress.value = 10 + assetProgress
    })

    loadingProgress.value = 90

    // Try to restore session on app mount
    await authStore.init()
    loadingProgress.value = 100

    // Small delay to show 100% before hiding loader
    await new Promise(resolve => setTimeout(resolve, 200))
    isAppReady.value = true
  } catch (error) {
    console.error('Failed to initialize app:', error)
    initError.value = 'Failed to load game assets. Please refresh the page.'
    uiStore.setInitializationError(initError.value)
  }
})
</script>

<style lang="scss">
@use "@/assets/styles/App.scss";
</style>
