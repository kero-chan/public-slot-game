/**
 * Spin Wins Store
 * Manages current spin wins and accumulated amounts
 */

import { defineStore } from 'pinia'
import type { WinCombination } from '@/features/spin/types'
import type { SpinResponse } from '@/types/global'
import { useGameFlowStore, GAME_STATES } from './gameFlowStore'
import { useFreeSpinsStore } from './freeSpinsStore'

export interface SpinWinsState {
  /** Current win amount for this spin */
  currentWin: number
  /** Number of consecutive wins in current cascade chain */
  consecutiveWins: number
  /** Current multiplier from backend (authoritative) */
  currentMultiplier: number
  /** Current winning combinations */
  currentWins: WinCombination[] | null
  /** Accumulated win amount for entire spin (including cascades) */
  accumulatedWinAmount: number
  /** All wins collected during this spin */
  allWinsThisSpin: WinCombination[]
  /** Spin response from backend */
  spinResponse: SpinResponse | null
  /** High-value symbol that won (fa, zhong, bai, bawan) - shown after all cascades */
  highValueWinSymbol: string | null
}

/**
 * Spin Wins Store - manages win tracking during a spin
 */
export const useSpinWinsStore = defineStore('spinWins', {
  state: (): SpinWinsState => ({
    currentWin: 0,
    consecutiveWins: 0,
    currentMultiplier: 1,
    currentWins: null,
    accumulatedWinAmount: 0,
    allWinsThisSpin: [],
    spinResponse: null,
    highValueWinSymbol: null
  }),

  getters: {
    /**
     * Check if current spin has any wins
     */
    hasWins(): boolean {
      return this.currentWins !== null && this.currentWins.length > 0
    },

    /**
     * Get total win amount for current spin session
     */
    totalWinAmount(): number {
      return this.allWinsThisSpin.reduce((sum, win) => sum + (win.payout || 0), 0)
    }
  },

  actions: {
    /**
     * Set current win amount and add to accumulated
     */
    setCurrentWin(amount: number): void {
      this.currentWin = amount
      this.accumulatedWinAmount += amount
    },

    setAccumulatedWinAmount(amount: number): void {
      this.accumulatedWinAmount = amount
    },

    /**
     * Set win results from backend
     */
    setWinResults(wins: WinCombination[] | null): void {
      this.currentWins = wins
      if (wins && wins.length > 0) {
        this.consecutiveWins++
        this.allWinsThisSpin.push(...wins)
      }
    },

    /**
     * Clear current wins
     */
    clearCurrentWins(): void {
      this.currentWins = null
    },

    /**
     * Increment consecutive wins counter
     */
    incrementConsecutiveWins(): void {
      this.consecutiveWins++
    },

    /**
     * Reset consecutive wins counter
     */
    resetConsecutiveWins(): void {
      this.consecutiveWins = 0
    },

    /**
     * Set current multiplier from backend cascade data
     */
    setCurrentMultiplier(multiplier: number): void {
      this.currentMultiplier = multiplier
    },

    /**
     * Reset multiplier to base value
     */
    resetMultiplier(isFreeSpin: boolean): void {
      this.currentMultiplier = isFreeSpin ? 2 : 1
    },

    /**
     * Set spin response from backend
     */
    setSpinResponse(response: SpinResponse): void {
      this.spinResponse = response
    },

    /**
     * Clear spin response
     */
    clearSpinResponse(): void {
      this.spinResponse = null
    },

    /**
     * Set high-value win symbol (fa, zhong, bai, bawan) for animation after cascades
     */
    setHighValueWinSymbol(symbol: string | null): void {
      this.highValueWinSymbol = symbol
    },

    /**
     * Reset for new spin (normal mode)
     */
    resetForNewSpin(): void {
      this.resetSpin()
      this.currentMultiplier = 1
    },

    /**
     * Reset for free spin (keeps accumulated amount across the session)
     * Do NOT zero accumulatedWinAmount here; it is used to decide
     * whether to show the final jackpot result overlay at session end.
     * Accumulated is cleared only when entering free spin mode via resetAccumulated().
     */
    resetForFreeSpin(): void {
      this.resetSpin()
      this.currentMultiplier = 2
    },

    resetSpin(): void {
      this.currentWin = 0
      this.consecutiveWins = 0
      this.currentWins = null
      this.accumulatedWinAmount = 0
      this.allWinsThisSpin = []
      this.spinResponse = null
      this.highValueWinSymbol = null
    },

    /**
     * Reset accumulated amounts (for free spin session start)
     */
    resetAccumulated(): void {
      this.accumulatedWinAmount = 0
      this.allWinsThisSpin = []
    },

    /**
     * Full reset
     */
    reset(): void {
      this.currentWin = 0
      this.consecutiveWins = 0
      this.currentMultiplier = 1
      this.currentWins = null
      this.accumulatedWinAmount = 0
      this.allWinsThisSpin = []
      this.spinResponse = null
      this.highValueWinSymbol = null
    },

    // ==================== Win Flow Actions ====================

    /**
     * Start checking for wins
     */
    startCheckingWins(): void {
      const gameFlowStore = useGameFlowStore()
      const state = gameFlowStore.gameFlowState

      if (state !== GAME_STATES.SPIN_COMPLETE &&
          state !== GAME_STATES.WAITING_AFTER_CASCADE &&
          state !== GAME_STATES.CHECKING_BONUS) return

      gameFlowStore.transitionTo(GAME_STATES.CHECKING_WINS)
    },

    /**
     * Set win results and transition to appropriate state
     */
    setWinResultsAndTransition(wins: WinCombination[] | null): void {
      const gameFlowStore = useGameFlowStore()

      this.setWinResults(wins)

      if (!wins || wins.length === 0) {
        gameFlowStore.transitionTo(GAME_STATES.NO_WINS)
      } else {
        gameFlowStore.transitionTo(GAME_STATES.HIGHLIGHTING_WINS)
      }
    },

    /**
     * Complete win highlighting
     */
    completeHighlighting(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.HIGHLIGHTING_WINS) return
      gameFlowStore.transitionTo(GAME_STATES.TRANSFORMING_GOLD)
    },

    /**
     * Complete gold transformation
     */
    completeGoldTransformation(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.TRANSFORMING_GOLD) return
      gameFlowStore.transitionTo(GAME_STATES.WAITING_AFTER_GOLD)
    },

    /**
     * Complete gold wait
     */
    completeGoldWait(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.WAITING_AFTER_GOLD) return
      gameFlowStore.transitionTo(GAME_STATES.DISAPPEARING_TILES)
    },

    /**
     * Complete tile disappearing
     */
    completeDisappearing(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.DISAPPEARING_TILES) return
      gameFlowStore.transitionTo(GAME_STATES.CASCADING)
    },

    /**
     * Complete cascade animation
     */
    completeCascade(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.CASCADING) return
      gameFlowStore.transitionTo(GAME_STATES.WAITING_AFTER_CASCADE)
    },

    /**
     * Complete cascade wait
     */
    completeCascadeWait(): void {
      const gameFlowStore = useGameFlowStore()
      if (gameFlowStore.gameFlowState !== GAME_STATES.WAITING_AFTER_CASCADE) return
      gameFlowStore.transitionTo(GAME_STATES.CHECKING_WINS)
    },

    /**
     * Complete no wins state
     */
    completeNoWins(): void {
      const gameFlowStore = useGameFlowStore()
      const freeSpinsStore = useFreeSpinsStore()

      if (gameFlowStore.gameFlowState !== GAME_STATES.NO_WINS) return

      if (freeSpinsStore.inFreeSpinMode) {
        const currentSpinWinAmount = this.allWinsThisSpin.reduce(
          (sum, win) => sum + (win.payout || 0), 0
        )

        if (currentSpinWinAmount > 0) {
          gameFlowStore.transitionTo(GAME_STATES.SHOWING_WIN_OVERLAY)
        } else if (freeSpinsStore.pendingRetriggerSpins > 0) {
          // Retrigger with no wins - show retrigger overlay
          gameFlowStore.transitionTo(GAME_STATES.SHOWING_RETRIGGER_OVERLAY)
        } else if (freeSpinsStore.freeSpins > 0) {
          gameFlowStore.transitionTo(GAME_STATES.FREE_SPINS_ACTIVE)
        } else {
          // Use freeSpinSessionWinAmount for the full session win amount
          if (freeSpinsStore.freeSpinSessionWinAmount > 0) {
            freeSpinsStore.markFinalJackpotResultShown()
            gameFlowStore.transitionTo(GAME_STATES.SHOWING_FINAL_JACKPOT_RESULT)
          } else {
            gameFlowStore.endSpinCycle()
          }
        }
      } else {
        // Not in free spin mode - check if cascades triggered free spins
        const response = this.spinResponse
        if (response?.free_spins_triggered && response?.free_spins_remaining_spins) {
          // Bonus was triggered during cascades - go to bonus flow
          gameFlowStore.transitionTo(GAME_STATES.CHECKING_BONUS)
        } else if (this.accumulatedWinAmount > 0) {
          gameFlowStore.transitionTo(GAME_STATES.SHOWING_WIN_OVERLAY)
        } else {
          gameFlowStore.endSpinCycle()
        }
      }
    },

    /**
     * Complete win overlay
     */
    completeWinOverlay(): void {
      const gameFlowStore = useGameFlowStore()
      const freeSpinsStore = useFreeSpinsStore()

      if (gameFlowStore.gameFlowState !== GAME_STATES.SHOWING_WIN_OVERLAY) return

      // Check for pending retrigger - show overlay before continuing
      if (freeSpinsStore.inFreeSpinMode && freeSpinsStore.pendingRetriggerSpins > 0) {
        this.resetForFreeSpin()
        gameFlowStore.transitionTo(GAME_STATES.SHOWING_RETRIGGER_OVERLAY)
      } else if (freeSpinsStore.inFreeSpinMode && freeSpinsStore.freeSpins > 0) {
        this.resetForFreeSpin()
        gameFlowStore.transitionTo(GAME_STATES.FREE_SPINS_ACTIVE)
      } else if (freeSpinsStore.inFreeSpinMode && freeSpinsStore.freeSpins <= 0 && !freeSpinsStore.finalJackpotResultShown) {
        freeSpinsStore.markFinalJackpotResultShown()
        gameFlowStore.transitionTo(GAME_STATES.SHOWING_FINAL_JACKPOT_RESULT)
      } else if (!freeSpinsStore.inFreeSpinMode) {
        // Not in free spin mode - check if cascades triggered free spins
        const response = this.spinResponse
        if (response?.free_spins_triggered && response?.free_spins_remaining_spins) {
          // Bonus was triggered during cascades - go to bonus flow
          gameFlowStore.transitionTo(GAME_STATES.CHECKING_BONUS)
        } else {
          gameFlowStore.endSpinCycle()
        }
      } else {
        gameFlowStore.endSpinCycle()
      }
    },

    // ==================== Spin Response Actions ====================

    /**
     * Process and store spin response from backend
     */
    processSpinResponse(response: SpinResponse): void {
      const freeSpinsStore = useFreeSpinsStore()

      this.setSpinResponse(response)

      // Sync freeSpinSessionWinAmount with backend's free_session_total_win (authoritative source)
      // Store as pending - will be applied when cascades/animations complete
      if (freeSpinsStore.inFreeSpinMode && response.free_session_total_win !== undefined) {
        console.log(`💰 Server free_session_total_win: ${response.free_session_total_win} | Client freeSpinSessionWinAmount: ${freeSpinsStore.freeSpinSessionWinAmount}`)
        freeSpinsStore.setPendingFreeSpinSessionWinAmount(response.free_session_total_win)
      }

      // Check for free spins retrigger
      const isRetrigger = response.free_spins_retriggered && response.free_spins_additional

      if (isRetrigger) {
        // Store the new total to apply AFTER retrigger overlay dismisses
        // This ensures the footer shows the old count until the overlay is dismissed
        if (response.free_spins_remaining_spins !== undefined) {
          freeSpinsStore.setPendingFreeSpinsTotal(response.free_spins_remaining_spins)
        }
        freeSpinsStore.setPendingRetrigger(response.free_spins_additional)
      } else {
        // SECURITY: Sync free spins count from backend (authoritative source)
        // Only update immediately if NOT a retrigger
        if (freeSpinsStore.inFreeSpinMode && response.free_spins_remaining_spins !== undefined) {
          freeSpinsStore.setFreeSpins(response.free_spins_remaining_spins)
        }
      }
    }
  }
})
