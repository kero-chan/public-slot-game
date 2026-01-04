/**
 * Betting Store
 * Manages bet amount and related actions
 */

import { defineStore } from 'pinia'
import { CONFIG } from '@/config/constants'
import { useAuthStore } from '../user/auth'
import { useSettingsStore } from '../user/settingsStore'
import { useGameFlowStore, GAME_STATES } from './gameFlowStore'

export interface BettingState {
  // State is empty - bet amount is now in settingsStore
}

/**
 * Betting Store - manages bet amount logic
 * Actual bet value is persisted in settingsStore
 */
export const useBettingStore = defineStore('betting', {
  state: (): BettingState => ({
    // Empty state - using settingsStore for persistence
  }),

  getters: {
    /**
     * Get current bet amount from settingsStore
     */
    bet(): number {
      const settingsStore = useSettingsStore()
      return settingsStore.betAmount
    },

    /**
     * Check if player can increase bet
     */
    canIncreaseBet(): boolean {
      const gameFlowStore = useGameFlowStore()
      const settingsStore = useSettingsStore()
      return gameFlowStore.isIdle && settingsStore.betAmount < CONFIG.game.maxBet
    },

    /**
     * Check if player can decrease bet
     */
    canDecreaseBet(): boolean {
      const gameFlowStore = useGameFlowStore()
      const settingsStore = useSettingsStore()
      return gameFlowStore.isIdle && settingsStore.betAmount > CONFIG.game.minBet
    },

    /**
     * Check if balance is sufficient for current bet
     */
    hasSufficientBalance(): boolean {
      const authStore = useAuthStore()
      const settingsStore = useSettingsStore()
      const balance = authStore.balance || 0
      return balance >= settingsStore.betAmount
    }
  },

  actions: {
    /**
     * Increase bet amount
     */
    increaseBet(): void {
      if (this.canIncreaseBet) {
        const settingsStore = useSettingsStore()
        settingsStore.increaseBetAmount()
      }
    },

    /**
     * Decrease bet amount
     */
    decreaseBet(): void {
      if (this.canDecreaseBet) {
        const settingsStore = useSettingsStore()
        settingsStore.decreaseBetAmount()
      }
    },

    /**
     * Set bet to specific amount
     */
    setBet(amount: number): void {
      const settingsStore = useSettingsStore()
      settingsStore.setBetAmount(amount)
    },

    /**
     * Reset bet to minimum
     */
    resetBet(): void {
      const settingsStore = useSettingsStore()
      settingsStore.resetBetAmount()
    }
  }
})
