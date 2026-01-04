import { computed, type ComputedRef, type Ref } from 'vue'
import { useGameStore, type LoadingProgress } from '@/stores'
import { storeToRefs } from 'pinia'
import type { SpinResponse } from '@/types/global'

/**
 * Game state composable interface
 */
export interface UseGameState {
  credits: Ref<number>
  bet: Ref<number>
  currentWin: Ref<number>
  accumulatedWinAmount: Ref<number>
  isSpinning: Ref<boolean>
  showingWinOverlay: Ref<boolean>
  consecutiveWins: Ref<number>
  currentMultiplier: Ref<number>
  freeSpins: Ref<number>
  inFreeSpinMode: Ref<boolean>
  isShowAmountNotification: Ref<boolean>
  canSpin: ComputedRef<boolean>
  showStartScreen: Ref<boolean>
  loadingProgress: Ref<LoadingProgress>
  gameSound: Ref<boolean>
  animationComplete: Ref<boolean>
  spinResponse: Ref<SpinResponse | null>
  freeSpinSessionWinAmount: Ref<number>
}

export function useGameState(): UseGameState {
  const gameStore = useGameStore()

  // Use storeToRefs to maintain reactivity for state properties
  const {
    credits,
    bet,
    currentWin,
    accumulatedWinAmount,
    isSpinning,
    showingWinOverlay,
    consecutiveWins,
    currentMultiplier,
    freeSpins,
    inFreeSpinMode,
    isShowAmountNotification,
    showStartScreen,
    loadingProgress,
    gameSound,
    spinResponse,
    freeSpinSessionWinAmount,
  } = storeToRefs(gameStore)

  // Getters are already computed in the store, but we need to access them as refs
  const canSpin = computed(() => gameStore.canSpin)
  const animationComplete = computed(() => gameStore.animationComplete)

  return {
    credits,
    bet,
    currentWin,
    accumulatedWinAmount,
    isSpinning,
    showingWinOverlay,
    consecutiveWins,
    currentMultiplier,
    freeSpins,
    inFreeSpinMode,
    isShowAmountNotification,
    canSpin,
    showStartScreen,
    loadingProgress,
    gameSound,
    animationComplete,
    spinResponse,
    freeSpinSessionWinAmount
  }
}
