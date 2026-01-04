<template>
  <div class="game-container">
    <!-- Trial Mode Label -->
    <div v-if="authStore.isTrial" class="trial-label">
      TRIAL
    </div>

    <!-- Slot Machine - PixiJS canvas will be created and appended to body -->
    <!-- Background is now rendered inside PixiJS canvas -->
    <div class="gameView">
      <StartScreen />
    </div>

    <!-- Settings Menu -->
    <SettingsMenu />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { gsap } from 'gsap'
import { useSlotMachine } from '@/composables/useSlotMachine'
import { useSettingsStore, useAuthStore } from '@/stores'
import { audioManager } from '@/composables/audioManager'
import StartScreen from '@/components/StartScreen.vue'
import SettingsMenu from '@/components/SettingsMenu.vue'

const settingsStore = useSettingsStore()
const authStore = useAuthStore()
const isInitialized = ref(false)

// Background is now rendered inside PixiJS canvas - no CSS background needed

// Initialize audio manager
audioManager.setGameSoundEnabled(settingsStore.gameSound)

const {
  init,
  handleResize,
  handleKeydown,
  stopAnimation,
  pauseForVisibility,
  resumeFromVisibility,
} = useSlotMachine(null)  // No canvasRef needed - PixiJS creates its own canvas

// Watch gameSound state
watch(
  () => settingsStore.gameSound,
  (newValue) => {
    audioManager.setGameSoundEnabled(newValue)
  }
)

/**
 * Handle visibility change to pause/resume game when tab is hidden/visible
 * This prevents background tabs from consuming CPU/GPU resources
 */
const handleVisibilityChange = (): void => {
  if (document.hidden) {
    // Tab is now hidden - pause everything
    pauseForVisibility()
    gsap.globalTimeline.pause()
  } else {
    // Tab is now visible - resume everything
    resumeFromVisibility()
    gsap.globalTimeline.resume()
  }
}

onMounted(() => {
  if (isInitialized.value) {
    return
  }
  isInitialized.value = true

  init()
  window.addEventListener('resize', handleResize)
  // PixiJS handles all button clicks internally - no need for DOM listeners
  // Keeping only keyboard handler for accessibility
  document.addEventListener('keydown', handleKeydown)
  // Listen for tab visibility changes to pause/resume game
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onBeforeUnmount(() => {
  isInitialized.value = false
  stopAnimation()

  window.removeEventListener('resize', handleResize)
  document.removeEventListener('keydown', handleKeydown)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped lang="scss">
@use "@/assets/styles/views/GameView.scss";

.trial-label {
  position: fixed;
  top: 20px;
  left: -42px;
  z-index: 9999;
  background: linear-gradient(135deg, #ff6b35 0%, #f7931e 100%);
  color: white;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 2px;
  padding: 8px 50px;
  transform: rotate(-45deg);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  text-transform: uppercase;
  pointer-events: none;
}
</style>
