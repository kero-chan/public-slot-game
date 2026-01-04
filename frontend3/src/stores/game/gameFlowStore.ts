/**
 * Game Flow Store
 * Manages the game state machine and flow transitions
 */

import { defineStore } from 'pinia'
import { useBettingStore } from './bettingStore'
import { useFreeSpinsStore } from './freeSpinsStore'
import { useSpinWinsStore } from './spinWinsStore'
import { useUIStore } from '../ui/uiStore'
import { useAuthStore } from '../user/auth'

/**
 * Game state machine states
 */
export const GAME_STATES = {
  IDLE: 'idle',
  SPINNING: 'spinning',
  SPIN_COMPLETE: 'spin_complete',
  CHECKING_BONUS: 'checking_bonus',
  POPPING_BONUS_TILES: 'popping_bonus_tiles',
  SHOWING_JACKPOT_ANIMATION: 'showing_jackpot_animation',
  SHOWING_BONUS_OVERLAY: 'showing_bonus_overlay',
  FREE_SPINS_ACTIVE: 'free_spins_active',
  CHECKING_WINS: 'checking_wins',
  NO_WINS: 'no_wins',
  HIGHLIGHTING_WINS: 'highlighting_wins',
  TRANSFORMING_GOLD: 'transforming_gold',
  WAITING_AFTER_GOLD: 'waiting_after_gold',
  DISAPPEARING_TILES: 'disappearing_tiles',
  CASCADING: 'cascading',
  WAITING_AFTER_CASCADE: 'waiting_after_cascade',
  SHOWING_WIN_OVERLAY: 'showing_win_overlay',
  SHOWING_FINAL_JACKPOT_RESULT: 'showing_final_jackpot_result',
  SHOWING_RETRIGGER_OVERLAY: 'showing_retrigger_overlay'
} as const

export type GameState = typeof GAME_STATES[keyof typeof GAME_STATES]

export interface GameFlowState {
  /** Current game flow state */
  gameFlowState: GameState
  /** Previous game flow state (for debugging) */
  previousGameFlowState: GameState | null
  /** Whether animation is complete (for flow control) */
  animationComplete: boolean
}

/**
 * Game Flow Store - manages state machine transitions
 */
export const useGameFlowStore = defineStore('gameFlow', {
  state: (): GameFlowState => ({
    gameFlowState: GAME_STATES.IDLE,
    previousGameFlowState: null,
    animationComplete: false
  }),

  getters: {
    /**
     * Check if game is in progress (not idle)
     */
    isInProgress(): boolean {
      return this.gameFlowState !== GAME_STATES.IDLE
    },

    /**
     * Check if currently spinning
     */
    isSpinning(): boolean {
      return this.gameFlowState === GAME_STATES.SPINNING
    },

    /**
     * Check if in idle state
     */
    isIdle(): boolean {
      return this.gameFlowState === GAME_STATES.IDLE
    }
  },

  actions: {
    /**
     * Transition to a new game state
     */
    transitionTo(newState: GameState): void {
      this.previousGameFlowState = this.gameFlowState
      this.gameFlowState = newState
      this.animationComplete = false
    },

    /**
     * Mark current animation as complete
     */
    markAnimationComplete(): void {
      this.animationComplete = true
    },

    /**
     * Reset to idle state
     */
    resetToIdle(): void {
      this.gameFlowState = GAME_STATES.IDLE
      this.previousGameFlowState = null
      this.animationComplete = false
    },

    // ==================== Spin Cycle Actions ====================

    /**
     * Start a new spin cycle
     * Returns true if spin was started successfully
     */
    startSpinCycle(): boolean {
      const freeSpinsStore = useFreeSpinsStore()
      const spinWinsStore = useSpinWinsStore()
      const uiStore = useUIStore()
      const authStore = useAuthStore()
      const bettingStore = useBettingStore()
      const balance = authStore.balance || 0

      // CRITICAL: Check and set state atomically to prevent race condition
      if (!this.isIdle) {
        console.warn('⚠️ startSpinCycle blocked: gameFlowState is', this.gameFlowState)
        return false
      }

      // IMMEDIATELY transition to SPINNING to block duplicate calls
      this.transitionTo(GAME_STATES.SPINNING)
      uiStore.setShowAmountNotification(false)
      spinWinsStore.clearSpinResponse()

      if (freeSpinsStore.inFreeSpinMode) {
        if (freeSpinsStore.freeSpins <= 0) {
          this.transitionTo(GAME_STATES.IDLE)
          return false
        }
        // Reset multiplier to base free spin value (2x)
        freeSpinsStore.decreaseFreeSpins()
        // Sync to backend's authoritative value at start of each spin
        if (freeSpinsStore.pendingFreeSpinSessionWinAmount > 0) {
          freeSpinsStore.setFreeSpinSessionWinAmount(freeSpinsStore.pendingFreeSpinSessionWinAmount)
          freeSpinsStore.setPendingFreeSpinSessionWinAmount(0)
        }
        spinWinsStore.resetMultiplier(true)
      } else {
        if (balance < bettingStore.bet) {
          // Rollback state change
          this.transitionTo(GAME_STATES.IDLE)
          return false
        }
        // Backend will deduct credits - we just reset win tracking
        spinWinsStore.resetForNewSpin()
      }

      spinWinsStore.resetConsecutiveWins()
      uiStore.deactivateAnticipationMode()

      return true
    },

    /**
     * Complete the spin animation
     */
    completeSpinAnimation(): void {
      if (this.gameFlowState !== GAME_STATES.SPINNING) return
      this.transitionTo(GAME_STATES.SPIN_COMPLETE)
    },

    /**
     * End the current spin cycle
     */
    endSpinCycle(): void {
      const freeSpinsStore = useFreeSpinsStore()
      const uiStore = useUIStore()
      const spinWinsStore = useSpinWinsStore()

      if (freeSpinsStore.inFreeSpinMode && freeSpinsStore.freeSpins <= 0) {
        // Exit free spin mode and reset accumulated win amount to 0
        // (the total was already shown in the final jackpot result overlay)
        freeSpinsStore.exitFreeSpinMode()
        spinWinsStore.setAccumulatedWinAmount(0)
      }

      uiStore.hideWinOverlay()
      spinWinsStore.resetMultiplier(freeSpinsStore.inFreeSpinMode)
      this.transitionTo(GAME_STATES.IDLE)
      spinWinsStore.clearCurrentWins()
      spinWinsStore.clearSpinResponse()
    },

    /**
     * Legacy start spin (for backward compatibility)
     */
    startSpin(): void {
      const freeSpinsStore = useFreeSpinsStore()
      const spinWinsStore = useSpinWinsStore()
      const uiStore = useUIStore()

      if (freeSpinsStore.inFreeSpinMode) {
        spinWinsStore.resetForFreeSpin()
      } else {
        spinWinsStore.resetForNewSpin()
      }
      uiStore.setShowAmountNotification(false)
    }
  }
})
