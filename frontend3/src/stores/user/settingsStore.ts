/**
 * Settings Store
 * Manages user preferences persisted to localStorage
 */

import { defineStore } from 'pinia'
import { CONFIG } from '@/config/constants'

/**
 * Settings store state interface
 */
export interface SettingsState {
  /** Whether game sound is enabled */
  gameSound: boolean
  /** Current bet amount */
  betAmount: number
  /** Whether fast spin mode is enabled (only for normal spins, not jackpot/free spins) */
  fastSpin: boolean
}

/**
 * Settings store - persisted to localStorage
 * Only contains user preferences that should persist across sessions
 */
export const useSettingsStore = defineStore('settings', {
  state: (): SettingsState => ({
    gameSound: true,
    betAmount: CONFIG.game.minBet,
    fastSpin: false
  }),

  actions: {
    /**
     * Toggle game sound on/off
     */
    toggleGameSound(): void {
      this.gameSound = !this.gameSound
    },

    /**
     * Set game sound to specific value
     * @param value - New game sound value
     */
    setGameSound(value: boolean): void {
      this.gameSound = value
    },

    /**
     * Set bet amount
     * @param amount - New bet amount
     */
    setBetAmount(amount: number): void {
      if (amount >= CONFIG.game.minBet && amount <= CONFIG.game.maxBet) {
        this.betAmount = amount
      }
    },

    /**
     * Increase bet amount by step
     */
    increaseBetAmount(): void {
      const newAmount = this.betAmount + CONFIG.game.betStep
      if (newAmount <= CONFIG.game.maxBet) {
        this.betAmount = newAmount
      }
    },

    /**
     * Decrease bet amount by step
     */
    decreaseBetAmount(): void {
      const newAmount = this.betAmount - CONFIG.game.betStep
      if (newAmount >= CONFIG.game.minBet) {
        this.betAmount = newAmount
      }
    },

    /**
     * Reset bet amount to minimum
     */
    resetBetAmount(): void {
      this.betAmount = CONFIG.game.minBet
    },

    /**
     * Toggle fast spin mode on/off
     */
    toggleFastSpin(): void {
      this.fastSpin = !this.fastSpin
    },

    /**
     * Set fast spin mode to specific value
     * @param value - New fast spin mode value
     */
    setFastSpin(value: boolean): void {
      this.fastSpin = value
    }
  },

  // Persist all settings to localStorage
  persist: {
    key: 'slot-game-settings',
    storage: localStorage
  }
})
