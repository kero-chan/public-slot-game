<template>
  <!-- Background is now rendered in PixiJS canvas -->
  <div v-if="gameState.showStartScreen.value" class="start-screen" :style="screenStyle">
    <div class="content">
      <!-- Loading Progress -->
      <div v-if="isLoading" class="loading-container">
        <div class="progress-bar-bg">
          <div class="progress-bar-fill" :style="{ width: `${loadingPercent}%` }"></div>
        </div>
        <div class="progress-text">
          <span v-if="loadingPercent < 100">Loading Resources...</span>
          <span v-else-if="!isHowlerReady">Preparing Sound System...</span>
          <span v-else>Loading Complete</span>
        </div>
      </div>

      <!-- Start Button -->
      <div v-else class="start-button-container" @click="handleStart">
        <img :src="assetImages.startBtn" alt="Start Button" class="start-button"
          :class="{ 'button-loading': isUnlockingAudio }" />

        <!-- Loading Spinner Overlay -->
        <div v-if="isUnlockingAudio" class="loading-spinner">
          <div class="spinner"></div>
          <div class="loading-text">Preparing...</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
// @ts-nocheck
import { computed, ref, watch, onMounted, onBeforeUnmount, type Ref, type ComputedRef } from "vue";
import { useGameLogic, type UseGameLogic } from '@/composables/slotMachine/useGameLogic'
import { useGameState } from "@/composables/slotMachine/useGameState";
import { audioManager } from "@/composables/audioManager";
import { howlerAudio } from "@/composables/useHowlerAudio";
import { audioEvents, AUDIO_EVENTS } from "@/composables/audioEventBus";
import { Howler } from "howler";
import { useGameStore, useSettingsStore, useGridStore, useFreeSpinsStore } from "@/stores";
import { CONFIG } from '@/config/constants';
import { ASSETS } from '@/config/assets';

interface LoadingProgress {
  loaded: number;
  total: number;
}

interface ScreenStyle {
  width: string;
  height: string;
}

const gameState = useGameState();
const gameStore = useGameStore();
const settingsStore = useSettingsStore();
const gridStore = useGridStore();
const backgroundMusic = audioManager.initialize();
// Set initial game sound state immediately after audio manager initialization
audioManager.setGameSoundEnabled(settingsStore.gameSound);
const gameLogic = useGameLogic(gameState, gridStore, null, null, null)

// Asset image URLs from API (ASSETS is reactive)
// Note: backgroundStart is now rendered in PixiJS canvas, not in HTML
const assetImages = computed(() => ({
  startBtn: ASSETS.imagePaths.startBtn || '',
}));

const isHowlerReady: Ref<boolean> = ref(false);

const isLoading: ComputedRef<boolean> = computed(() => {
  const progress: LoadingProgress = gameState.loadingProgress?.value || { loaded: 0, total: 1 };
  const assetsLoaded: boolean = progress.loaded >= progress.total;

  // Show loading until both assets are loaded AND Howler is ready
  return !assetsLoaded || !isHowlerReady.value;
});

const loadingPercent: ComputedRef<number> = computed(() => {
  const progress: LoadingProgress = gameState.loadingProgress?.value || { loaded: 0, total: 1 };
  const percent: number =
    progress.total > 0 ? (progress.loaded / progress.total) * 100 : 0;
  return Math.floor(percent);
});

const isUnlockingAudio: Ref<boolean> = ref(false);

// Reactive window dimensions for resize handling
const windowWidth: Ref<number> = ref(window.innerWidth);
const windowHeight: Ref<number> = ref(window.innerHeight);

const handleResize = (): void => {
  windowWidth.value = window.innerWidth;
  windowHeight.value = window.innerHeight;
};

onMounted(() => {
  window.addEventListener('resize', handleResize);
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize);
});

// Calculate screen dimensions to match canvas positioning
const screenStyle: ComputedRef<ScreenStyle> = computed(() => {
  const vw: number = windowWidth.value;
  const vh: number = windowHeight.value;
  const viewportRatio: number = vw / vh;
  const targetRatio: number = CONFIG.canvas.aspectRatio;

  // On smartphones (narrow viewports), fill viewport
  if (viewportRatio < targetRatio) {
    return {
      width: `${vw}px`,
      height: `${vh}px`,
    };
  }

  // On PC/tablets (wider viewports), match canvas dimensions (configured aspect ratio centered)
  const height: number = vh;
  const width: number = Math.floor(height * CONFIG.canvas.aspectRatio);

  return {
    width: `${width}px`,
    height: `${height}px`,
  };
});

// Check if Howler is ready
const checkHowlerReady = (): void => {
  if (howlerAudio.isReady()) {
    isHowlerReady.value = true;
  } else {
    // Keep checking every 100ms until ready
    setTimeout(checkHowlerReady, 100);
  }
};

// Start checking when assets finish loading
const startHowlerCheck = (): void => {
  const progress: LoadingProgress = gameState.loadingProgress?.value || { loaded: 0, total: 1 };
  if (progress.loaded >= progress.total && !isHowlerReady.value) {
    checkHowlerReady();
  }
};

// Watch loading progress
watch(() => gameState.loadingProgress?.value, (newProgress: LoadingProgress | undefined) => {
  if (newProgress && newProgress.loaded >= newProgress.total) {
    startHowlerCheck();
  }
}, { immediate: true, deep: true });

const handleStart = async (): Promise<void> => {
  if (isUnlockingAudio.value) {
    return;
  }

  // Play start button sound
  audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'start_button', volume: 0.7 });

  isUnlockingAudio.value = true;

  // Fetch initial grid from backend when user starts the game
  // This ensures zero frontend RNG and fresh grid on game start
  try {
    await gridStore.fetchInitialGrid()
  } catch {
    // Continue anyway - grid store has fallback logic
  }

  // Clean up any active sessions before starting the game
  try {
    const { useBackendGameStore } = await import('@/stores')
    const { gameApi } = await import('@/api/game')
    const backendGame = useBackendGameStore()

    // Reset frontend state
    backendGame.reset()

    // Check for active free spin session FIRST (before ending any session)
    let hasActiveFreeSpins = false
    try {
      const freeSpinsStatus = await gameApi.getFreeSpinsStatus()

      if (freeSpinsStatus.active && freeSpinsStatus.free_spins_session_id) {
        hasActiveFreeSpins = true
        // Player has active free spins!
        // Set the free spins session ID in backend game store
        backendGame.currentFreeSpinsSessionId = freeSpinsStatus.free_spins_session_id

        // Also set the session ID to preserve provably fair session
        if (freeSpinsStatus.session_id) {
          backendGame.setCurrentSessionId(freeSpinsStatus.session_id)
        }

        // Activate free spin mode in game store (WITHOUT starting jackpot music yet)
        // Music will be started after AudioContext is unlocked
        gameStore.setFreeSpins(freeSpinsStatus.remaining_spins)
        gameStore.setPendingFreeSpinSessionWinAmount(freeSpinsStatus.total_won)
        gameStore.setFreeSpinSessionWinAmount(freeSpinsStatus.total_won)
        // Set inFreeSpinMode flag directly without triggering music
        useFreeSpinsStore().inFreeSpinMode = true
        useFreeSpinsStore().finalJackpotResultShown = false
      }
    } catch {
      // No active free spin session or error checking
    }

    // Only end active backend sessions if NO active free spins
    // Ending the session would also end the provably fair session needed for free spins
    if (!hasActiveFreeSpins) {
      try {
        const history = await gameApi.getSessionHistory(1, 1)
        if (history.sessions && history.sessions.length > 0) {
          const lastSession = history.sessions[0]
          if (!lastSession.ended_at) {
            await gameApi.endSession(lastSession.id)
          }
        }
      } catch {
        // No active sessions to clean up
      }
    }
  } catch {
    // Failed to cleanup sessions - continue anyway
  }

  // Timeout protection: force unlock after 3 seconds
  const unlockTimeout: NodeJS.Timeout = setTimeout(() => {
    if (isUnlockingAudio.value) {
      // Force start game even if unlock failed
      audioManager.setGameSoundEnabled(settingsStore.gameSound);
      // Start appropriate music based on mode
      if (gameStore.inFreeSpinMode) {
        audioManager.switchToJackpotMusic()
      } else {
        backgroundMusic.start();
      }
      gameStore.hideStartScreen();
    }
  }, 3000);

  try {
    // Step 1: Unlock AudioContext (required for mobile browsers)
    await howlerAudio.unlockAudioContext();

    // Step 2: Verify AudioContext is running
    const ctx = Howler.ctx;
    if (ctx) {
      if (ctx.state === 'suspended') {
        await ctx.resume();
      }
    }

    // Step 3: Additional delay to ensure everything is settled
    await new Promise<void>(resolve => setTimeout(resolve, 100));

    // Clear timeout since we succeeded
    clearTimeout(unlockTimeout);

    // Step 4: Set audio state
    audioManager.setGameSoundEnabled(settingsStore.gameSound);

    // Step 5: Start background music with retry
    // If in free spin mode, start jackpot music instead of normal music
    let musicStarted: boolean = false;

    if (gameStore.inFreeSpinMode) {
      // Resume AudioContext before free spins
      await howlerAudio.resumeAudioContext()

      // Start jackpot music for free spin mode
      audioManager.switchToJackpotMusic()
      musicStarted = true
    } else {
      // Normal game start - play normal background music
      for (let attempt = 1; attempt <= 3; attempt++) {
        try {
          const result: boolean = await backgroundMusic.start();

          if (result) {
            musicStarted = true;
            break;
          } else {
            if (attempt < 3) {
              await new Promise<void>(resolve => setTimeout(resolve, 200));
            }
          }
        } catch (err) {
          if (attempt < 3) {
            await new Promise<void>(resolve => setTimeout(resolve, 200));
          }
        }
      }

    }

    // Step 6: Hide start screen
    gameStore.hideStartScreen();
  } catch {
    // Clear timeout
    clearTimeout(unlockTimeout);

    // Force start game anyway
    audioManager.setGameSoundEnabled(settingsStore.gameSound);

    // Try to start appropriate music based on mode
    if (gameStore.inFreeSpinMode) {
      audioManager.switchToJackpotMusic()
    } else {
      try {
        await backgroundMusic.start();
      } catch {
        // Background music failed - continue anyway
      }
    }

    gameStore.hideStartScreen();
  }
};
</script>

<style scoped lang="scss">
@use "@/assets/styles/components/StartScreen.scss";
</style>
