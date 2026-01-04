/**
 * Game Store
 * Main game store that composes all domain-specific stores
 * Provides backward-compatible API while delegating to specialized stores
 */

import { defineStore } from 'pinia'
import { useSettingsStore } from '../user/settingsStore'
import { useAuthStore } from '../user/auth'
import { useGameFlowStore, GAME_STATES, type GameState } from './gameFlowStore'
import { useBettingStore } from './bettingStore'
import { useFreeSpinsStore } from './freeSpinsStore'
import { useUIStore, type LoadingProgress } from '../ui/uiStore'
import { useSpinWinsStore } from './spinWinsStore'
import type { WinCombination } from '@/features/spin/types'
import type { SpinResponse } from '@/types/global'

// Re-export for backward compatibility
export { GAME_STATES, type GameState } from './gameFlowStore'
export { type LoadingProgress } from '../ui/uiStore'

/**
 * Game store state interface (for backward compatibility)
 */
export interface GameStoreState {
  bet: number
  currentWin: number
  isSpinning: boolean
  showingWinOverlay: boolean
  gameFlowState: GameState
  previousGameFlowState: GameState | null
  consecutiveWins: number
  currentMultiplier: number
  currentWins: WinCombination[] | null
  accumulatedWinAmount: number
  allWinsThisSpin: WinCombination[]
  freeSpins: number
  inFreeSpinMode: boolean
  finalJackpotResultShown: boolean
  isShowAmountNotification: boolean
  anticipationMode: boolean
  showStartScreen: boolean
  animationComplete: boolean
  loadingProgress: LoadingProgress
  spinResponse: SpinResponse | null
}

/**
 * Game Store - main game orchestrator
 * Composes domain-specific stores and provides unified API
 */
export const useGameStore = defineStore('game', {
  state: (): { _placeholder: boolean } => ({
    // All state is now managed by composable stores
    // This placeholder prevents empty state errors
    _placeholder: true
  }),

  getters: {
    // ==================== State Getters (delegated to sub-stores) ====================

    bet(): number {
      return useBettingStore().bet
    },

    currentWin(): number {
      return useSpinWinsStore().currentWin
    },

    isSpinning(): boolean {
      return useGameFlowStore().isSpinning
    },

    showingWinOverlay(): boolean {
      return useUIStore().showingWinOverlay
    },

    gameFlowState(): GameState {
      return useGameFlowStore().gameFlowState
    },

    previousGameFlowState(): GameState | null {
      return useGameFlowStore().previousGameFlowState
    },

    consecutiveWins(): number {
      return useSpinWinsStore().consecutiveWins
    },

    currentMultiplier(): number {
      return useSpinWinsStore().currentMultiplier
    },

    currentWins(): WinCombination[] | null {
      return useSpinWinsStore().currentWins
    },

    accumulatedWinAmount(): number {
      return useSpinWinsStore().accumulatedWinAmount
    },

    allWinsThisSpin(): WinCombination[] {
      return useSpinWinsStore().allWinsThisSpin
    },

    freeSpins(): number {
      return useFreeSpinsStore().freeSpins
    },

    inFreeSpinMode(): boolean {
      return useFreeSpinsStore().inFreeSpinMode
    },

    freeSpinSessionWinAmount(): number {
      return useFreeSpinsStore().freeSpinSessionWinAmount
    },

    finalJackpotResultShown(): boolean {
      return useFreeSpinsStore().finalJackpotResultShown
    },

    isShowAmountNotification(): boolean {
      return useUIStore().isShowAmountNotification
    },

    anticipationMode(): boolean {
      return useUIStore().anticipationMode
    },

    showStartScreen(): boolean {
      return useUIStore().showStartScreen
    },

    animationComplete(): boolean {
      return useGameFlowStore().animationComplete
    },

    loadingProgress(): LoadingProgress {
      return useUIStore().loadingProgress
    },

    initializationError(): string | null {
      return useUIStore().initializationError
    },

    spinResponse(): SpinResponse | null {
      return useSpinWinsStore().spinResponse
    },

    highValueWinSymbol(): string | null {
      return useSpinWinsStore().highValueWinSymbol
    },

    // ==================== Computed Getters ====================

    credits(): number {
      const authStore = useAuthStore()
      return authStore.balance || 0
    },

    gameSound(): boolean {
      const settingsStore = useSettingsStore()
      return settingsStore.gameSound
    },

    canSpin(): boolean {
      const gameFlowStore = useGameFlowStore()
      const uiStore = useUIStore()
      const bettingStore = useBettingStore()
      const freeSpinsStore = useFreeSpinsStore()
      const authStore = useAuthStore()
      const balance = authStore.balance || 0

      return gameFlowStore.isIdle &&
             !uiStore.showingWinOverlay &&
             (balance >= bettingStore.bet || freeSpinsStore.inFreeSpinMode)
    },

    canIncreaseBet(): boolean {
      return useBettingStore().canIncreaseBet
    },

    canDecreaseBet(): boolean {
      return useBettingStore().canDecreaseBet
    },

    isInProgress(): boolean {
      return useGameFlowStore().isInProgress
    }
  },

  actions: {
    // ==================== State Machine Actions ====================

    transitionTo(newState: GameState): void {
      useGameFlowStore().transitionTo(newState)
    },

    // ==================== Spin Cycle Actions (delegated to gameFlowStore) ====================

    startSpinCycle(): boolean {
      return useGameFlowStore().startSpinCycle()
    },

    completeSpinAnimation(): void {
      useGameFlowStore().completeSpinAnimation()
    },

    endSpinCycle(): void {
      useGameFlowStore().endSpinCycle()
    },

    startSpin(): void {
      useGameFlowStore().startSpin()
    },

    endSpin(): void {
      // No longer needed - state managed by gameFlowStore
    },

    // ==================== Bonus Actions (delegated to freeSpinsStore) ====================

    startCheckingBonus(): void {
      useFreeSpinsStore().startCheckingBonus()
    },

    setBonusResults(freeSpinsTriggered: boolean, freeSpinsAwarded?: number): void {
      useFreeSpinsStore().setBonusResults(freeSpinsTriggered, freeSpinsAwarded)
    },

    completeBonusTilePop(): void {
      useFreeSpinsStore().completeBonusTilePop()
    },

    completeJackpotAnimation(): void {
      useFreeSpinsStore().completeJackpotAnimation()
    },

    completeBonusOverlay(): void {
      useFreeSpinsStore().completeBonusOverlay()
    },

    startFreeSpinRound(): boolean {
      return useFreeSpinsStore().startFreeSpinRound()
    },

    showRetriggerOverlay(): void {
      useFreeSpinsStore().showRetriggerOverlay()
    },

    completeRetriggerOverlay(): void {
      useFreeSpinsStore().completeRetriggerOverlay()
    },

    completeFinalJackpotResult(): void {
      useFreeSpinsStore().completeFinalJackpotResult()
    },

    // ==================== Win Actions (delegated to spinWinsStore) ====================

    startCheckingWins(): void {
      useSpinWinsStore().startCheckingWins()
    },

    setWinResults(wins: WinCombination[] | null): void {
      useSpinWinsStore().setWinResultsAndTransition(wins)
    },

    setHighValueWinSymbol(symbol: string | null): void {
      useSpinWinsStore().setHighValueWinSymbol(symbol)
    },

    completeHighlighting(): void {
      useSpinWinsStore().completeHighlighting()
    },

    completeGoldTransformation(): void {
      useSpinWinsStore().completeGoldTransformation()
    },

    completeGoldWait(): void {
      useSpinWinsStore().completeGoldWait()
    },

    completeDisappearing(): void {
      useSpinWinsStore().completeDisappearing()
    },

    completeCascade(): void {
      useSpinWinsStore().completeCascade()
    },

    completeCascadeWait(): void {
      useSpinWinsStore().completeCascadeWait()
    },

    completeNoWins(): void {
      useSpinWinsStore().completeNoWins()
    },

    completeWinOverlay(): void {
      useSpinWinsStore().completeWinOverlay()
    },

    markAnimationComplete(): void {
      useGameFlowStore().markAnimationComplete()
    },

    // ==================== Spin Response Actions (delegated to spinWinsStore) ====================

    setSpinResponse(response: SpinResponse): void {
      useSpinWinsStore().processSpinResponse(response)
    },

    // ==================== Betting Actions ====================

    increaseBet(): void {
      useBettingStore().increaseBet()
    },

    decreaseBet(): void {
      useBettingStore().decreaseBet()
    },

    setBet(amount: number): void {
      useBettingStore().setBet(amount)
    },

    // ==================== Credits Actions ====================

    addCredits(amount: number): void {
      const authStore = useAuthStore()
      authStore.updateBalance((authStore.balance || 0) + amount)
    },

    deductBet(): boolean {
      const authStore = useAuthStore()
      const bettingStore = useBettingStore()
      const balance = authStore.balance || 0
      return balance >= bettingStore.bet
    },

    // ==================== Win Amount Actions ====================

    setCurrentWin(amount: number): void {
      useSpinWinsStore().setCurrentWin(amount)
      if (this.inFreeSpinMode) {
        useFreeSpinsStore().increaseFreeSpinSessionWinAmount(amount)
      }
    },

    setFreeSpinSessionWinAmount(amount: number): void {
      useFreeSpinsStore().setFreeSpinSessionWinAmount(amount)
    },

    setPendingFreeSpinSessionWinAmount(amount: number): void {
      useFreeSpinsStore().setPendingFreeSpinSessionWinAmount(amount)
    },

    incrementConsecutiveWins(): void {
      useSpinWinsStore().incrementConsecutiveWins()
    },

    resetConsecutiveWins(): void {
      useSpinWinsStore().resetConsecutiveWins()
    },

    setCurrentMultiplier(multiplier: number): void {
      useSpinWinsStore().setCurrentMultiplier(multiplier)
    },

    setShowAmountNotification(value: boolean): void {
      useUIStore().setShowAmountNotification(value)
    },

    // ==================== Win Overlay Actions ====================

    showWinOverlay(): void {
      useUIStore().showWinOverlay()
    },

    hideWinOverlay(): void {
      useUIStore().hideWinOverlay()
    },

    // ==================== Free Spins Actions ====================

    setFreeSpins(count: number): void {
      useFreeSpinsStore().setFreeSpins(count)
    },

    enterFreeSpinMode(): void {
      useFreeSpinsStore().enterFreeSpinMode()
    },

    exitFreeSpinMode(): void {
      useFreeSpinsStore().exitFreeSpinMode()
    },

    // ==================== UI Actions ====================

    hideStartScreen(): void {
      useUIStore().hideStartScreen()
    },

    showStartScreenAgain(): void {
      useUIStore().showStartScreenAgain()
    },

    updateLoadingProgress(loaded: number, total: number): void {
      useUIStore().updateLoadingProgress(loaded, total)
    },

    toggleGameSound(): void {
      useSettingsStore().toggleGameSound()
    },

    activateAnticipationMode(): void {
      useUIStore().activateAnticipationMode()
    },

    deactivateAnticipationMode(): void {
      useUIStore().deactivateAnticipationMode()
    },

    // ==================== Reset Actions ====================

    resetGame(): void {
      useBettingStore().resetBet()
      useSpinWinsStore().reset()
      useFreeSpinsStore().reset()
      useUIStore().reset()
      useGameFlowStore().resetToIdle()
    }
  }
})
