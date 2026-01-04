/**
 * Free Spins Store
 * Manages free spins state and mode
 */

import { defineStore } from 'pinia'
import { audioManager } from '@/composables/audioManager'
import { useGameFlowStore, GAME_STATES } from './gameFlowStore'
import { useSpinWinsStore } from './spinWinsStore'

export interface FreeSpinsState {
  /** Number of free spins remaining */
  freeSpins: number
  /** Whether in free spin mode */
  inFreeSpinMode: boolean
  /** Whether final jackpot result has been shown */
  finalJackpotResultShown: boolean
  /** Pending retrigger additional spins (to show overlay) */
  pendingRetriggerSpins: number
  /** Pending total free spins count (to apply after retrigger overlay) */
  pendingFreeSpinsTotal: number | null
  /** Current free spin session win amount */
  freeSpinSessionWinAmount: number

  /** Pending free spin session win amount (to apply when start spin) */
  pendingFreeSpinSessionWinAmount: number

  /** Flag indicating free spin session should resume (set on reload) */
  pendingFreeSpinResume: boolean
}

/**
 * Free Spins Store - manages free spin mode
 */
export const useFreeSpinsStore = defineStore('freeSpins', {
  state: (): FreeSpinsState => ({
    freeSpins: 0,
    inFreeSpinMode: false,
    finalJackpotResultShown: false,
    pendingRetriggerSpins: 0,
    pendingFreeSpinsTotal: null,
    freeSpinSessionWinAmount: 0,
    pendingFreeSpinSessionWinAmount: 0,
    pendingFreeSpinResume: false,
  }),

  getters: {
    /**
     * Check if there are free spins remaining
     */
    hasSpinsRemaining(): boolean {
      return this.freeSpins > 0
    },

    /**
     * Check if free spin session is complete
     */
    isSessionComplete(): boolean {
      return this.inFreeSpinMode && this.freeSpins <= 0
    }
  },

  actions: {
    /**
     * Set free spins count (from server authority)
     */
    setFreeSpins(count: number): void {
      this.freeSpins = count
    },

    decreaseFreeSpins(): void {
      this.freeSpins--
    },

    /**
     * Enter free spin mode
     */
    enterFreeSpinMode(): void {
      this.inFreeSpinMode = true
      this.finalJackpotResultShown = false
      this.freeSpinSessionWinAmount = 0

      // Resume AudioContext before free spins
      import('@/composables/useHowlerAudio').then(async ({ howlerAudio }) => {
        if (howlerAudio.isReady()) {
          await howlerAudio.resumeAudioContext()
        }
      })
    },

    /**
     * Exit free spin mode
     */
    exitFreeSpinMode(): void {
      this.inFreeSpinMode = false
      this.freeSpins = 0
      this.finalJackpotResultShown = false
      this.freeSpinSessionWinAmount = 0
      // Switch back to normal background music
      audioManager.switchToNormalMusic()
    },

    /**
     * Mark final jackpot result as shown
     */
    markFinalJackpotResultShown(): void {
      this.finalJackpotResultShown = true
    },

    /**
     * Set pending retrigger spins (triggers overlay display)
     */
    setPendingRetrigger(additionalSpins: number): void {
      this.pendingRetriggerSpins = additionalSpins
      console.log('🎰 Free spins retrigger pending:', additionalSpins)
    },

    /**
     * Clear pending retrigger (after overlay shown)
     */
    clearPendingRetrigger(): void {
      this.pendingRetriggerSpins = 0
    },

    /**
     * Set pending free spins total (to apply after retrigger overlay)
     */
    setPendingFreeSpinsTotal(total: number): void {
      this.pendingFreeSpinsTotal = total
      console.log('🎰 Pending free spins total stored:', total)
    },

    /**
     * Apply pending free spins total (after retrigger overlay dismisses)
     */
    applyPendingFreeSpinsTotal(): void {
      if (this.pendingFreeSpinsTotal !== null) {
        console.log('🎰 Applying pending free spins total:', this.pendingFreeSpinsTotal)
        this.freeSpins = this.pendingFreeSpinsTotal
        this.pendingFreeSpinsTotal = null
      }
    },

    /**
     * Check if there's a pending retrigger
     */
    hasPendingRetrigger(): boolean {
      return this.pendingRetriggerSpins > 0
    },

    increaseFreeSpinSessionWinAmount(amount: number): void {
      this.freeSpinSessionWinAmount += amount
    },

    setFreeSpinSessionWinAmount(amount: number): void {
      this.freeSpinSessionWinAmount = amount
    },

    setPendingFreeSpinSessionWinAmount(amount: number): void {
      this.pendingFreeSpinSessionWinAmount = amount
    },

    /**
     * Set pending free spin resume flag (when resuming from reload)
     */
    setPendingFreeSpinResume(pending: boolean): void {
      this.pendingFreeSpinResume = pending
    },

    /**
     * Clear pending free spin resume flag
     */
    clearPendingFreeSpinResume(): void {
      this.pendingFreeSpinResume = false
    },

    /**
     * Reset free spins state
     */
    reset(): void {
      this.freeSpins = 0
      this.inFreeSpinMode = false
      this.finalJackpotResultShown = false
      this.pendingRetriggerSpins = 0
      this.pendingFreeSpinsTotal = null
      this.freeSpinSessionWinAmount = 0
      this.pendingFreeSpinResume = false
    },

    // ==================== Bonus Flow Actions ====================

    /**
     * Start checking for bonus
     */
    startCheckingBonus(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.SPIN_COMPLETE) return
      gameFlowStore.transitionTo(GAME_STATES.CHECKING_BONUS)
    },

    /**
     * Set bonus results and handle state transition
     */
    setBonusResults(freeSpinsTriggered: boolean, freeSpinsAwarded?: number): void {
      const gameFlowStore = useGameFlowStore()
      const spinWinsStore = useSpinWinsStore()

      if (freeSpinsTriggered && !this.inFreeSpinMode) {
        if (freeSpinsAwarded === undefined || freeSpinsAwarded <= 0) {
          console.error('❌ Server did not provide free spins count for bonus trigger')
          gameFlowStore.transitionTo(GAME_STATES.CHECKING_WINS)
          return
        }

        // Check if there are cascades with wins that need to play first
        // Only check this on the FIRST call (when freeSpins is not yet set)
        // After cascades complete, completeNoWins/completeWinOverlay will call CHECKING_BONUS again
        // At that point, freeSpins is already set, so we go directly to bonus
        if (this.freeSpins === 0) {
          // First time - store free spins and check for cascade wins
          this.setFreeSpins(freeSpinsAwarded)

          const response = spinWinsStore.spinResponse
          if (response?.cascades && response.cascades.length > 0) {
            const hasWins = response.cascades.some(c => c.wins && c.wins.length > 0)
            if (hasWins) {
              console.log('🎰 Free spins triggered but cascades have wins - playing cascade first')
              gameFlowStore.transitionTo(GAME_STATES.CHECKING_WINS)
              return
            }
          }
        }
        // Either no cascade wins, or cascades already played - trigger bonus now
        gameFlowStore.transitionTo(GAME_STATES.POPPING_BONUS_TILES)
      } else {
        gameFlowStore.transitionTo(GAME_STATES.CHECKING_WINS)
      }
    },

    /**
     * Complete bonus tile pop animation
     */
    completeBonusTilePop(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.POPPING_BONUS_TILES) return
      gameFlowStore.transitionTo(GAME_STATES.SHOWING_JACKPOT_ANIMATION)
    },

    /**
     * Complete jackpot animation
     */
    completeJackpotAnimation(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.SHOWING_JACKPOT_ANIMATION) return
      gameFlowStore.transitionTo(GAME_STATES.SHOWING_BONUS_OVERLAY)
    },

    /**
     * Complete bonus overlay and enter free spin mode
     */
    completeBonusOverlay(): void {
      const gameFlowStore = useGameFlowStore()
      const spinWinsStore = useSpinWinsStore()

      if (gameFlowStore.gameFlowState !== GAME_STATES.SHOWING_BONUS_OVERLAY) return

      this.enterFreeSpinMode()
      spinWinsStore.resetAccumulated()
      gameFlowStore.transitionTo(GAME_STATES.FREE_SPINS_ACTIVE)
    },

    /**
     * Start a free spin round
     */
    startFreeSpinRound(): boolean {
      if (!this.inFreeSpinMode) return false
      if (this.freeSpins <= 0) {
        this.exitFreeSpinMode()
        return false
      }
      return true
    },

    /**
     * Show the retrigger overlay (called when free spins are retriggered)
     */
    showRetriggerOverlay(): void {
      const gameFlowStore = useGameFlowStore()
      gameFlowStore.transitionTo(GAME_STATES.SHOWING_RETRIGGER_OVERLAY)
    },

    /**
     * Complete the retrigger overlay and continue free spins
     */
    completeRetriggerOverlay(): void {
      const gameFlowStore = useGameFlowStore()

      if (gameFlowStore.gameFlowState !== GAME_STATES.SHOWING_RETRIGGER_OVERLAY) return

      this.clearPendingRetrigger()
      // Apply the pending free spins total NOW (after overlay dismisses)
      this.applyPendingFreeSpinsTotal()
      gameFlowStore.transitionTo(GAME_STATES.FREE_SPINS_ACTIVE)
    },

    /**
     * Complete final jackpot result overlay
     */
    completeFinalJackpotResult(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.SHOWING_FINAL_JACKPOT_RESULT) return
      gameFlowStore.endSpinCycle()
    }
  }
})
