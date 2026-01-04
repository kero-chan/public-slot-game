/**
 * UI Store
 * Manages UI state like overlays, loading, and start screen
 */

import { defineStore } from 'pinia'
import type { SettingsMenuContentType } from '@/types/settingsMenu'

export interface LoadingProgress {
  /** Number of assets loaded */
  loaded: number
  /** Total number of assets to load */
  total: number
}

export type { SettingsMenuContentType } from '@/types/settingsMenu'

export interface UIState {
  /** Whether to show start screen */
  showStartScreen: boolean
  /** Whether win overlay is being shown */
  showingWinOverlay: boolean
  /** Whether to show amount notification */
  isShowAmountNotification: boolean
  /** Loading progress for assets */
  loadingProgress: LoadingProgress
  /** Anticipation mode (near-miss feature for bonus tiles) */
  anticipationMode: boolean
  /** Initialization error message (null if no error) */
  initializationError: string | null
  /** Settings menu visibility */
  isSettingsMenuOpen: boolean
  /** Settings menu active content */
  settingsMenuContent: SettingsMenuContentType
}

/**
 * UI Store - manages UI state
 */
export const useUIStore = defineStore('ui', {
  state: (): UIState => ({
    showStartScreen: true,
    showingWinOverlay: false,
    isShowAmountNotification: false,
    loadingProgress: { loaded: 0, total: 1 },
    anticipationMode: false,
    initializationError: null,
    isSettingsMenuOpen: false,
    settingsMenuContent: null
  }),

  getters: {
    /**
     * Get loading percentage
     */
    loadingPercentage(): number {
      if (this.loadingProgress.total === 0) return 0
      return Math.round((this.loadingProgress.loaded / this.loadingProgress.total) * 100)
    },

    /**
     * Check if loading is complete
     */
    isLoadingComplete(): boolean {
      return this.loadingProgress.loaded >= this.loadingProgress.total
    }
  },

  actions: {
    /**
     * Hide start screen
     */
    hideStartScreen(): void {
      this.showStartScreen = false
    },

    /**
     * Show start screen again
     */
    showStartScreenAgain(): void {
      this.showStartScreen = true
    },

    /**
     * Show win overlay
     */
    showWinOverlay(): void {
      this.showingWinOverlay = true
    },

    /**
     * Hide win overlay
     */
    hideWinOverlay(): void {
      this.showingWinOverlay = false
    },

    /**
     * Set amount notification visibility
     */
    setShowAmountNotification(value: boolean): void {
      this.isShowAmountNotification = value
    },

    /**
     * Update loading progress
     */
    updateLoadingProgress(loaded: number, total: number): void {
      this.loadingProgress = { loaded, total }
    },

    /**
     * Activate anticipation mode
     */
    activateAnticipationMode(): void {
      this.anticipationMode = true

      // Resume AudioContext before anticipation mode
      import('@/composables/useHowlerAudio').then(async ({ howlerAudio }) => {
        if (howlerAudio.isReady()) {
          await howlerAudio.resumeAudioContext()
        }
      })
    },

    /**
     * Deactivate anticipation mode
     */
    deactivateAnticipationMode(): void {
      this.anticipationMode = false
    },

    /**
     * Set initialization error
     */
    setInitializationError(error: string | null): void {
      this.initializationError = error
    },

    /**
     * Clear initialization error
     */
    clearInitializationError(): void {
      this.initializationError = null
    },

    /**
     * Open settings menu
     */
    openSettings(): void {
      this.isSettingsMenuOpen = true
      this.settingsMenuContent = null
    },

    /**
     * Close settings menu
     */
    closeSettings(): void {
      this.isSettingsMenuOpen = false
      this.settingsMenuContent = null
    },

    /**
     * Toggle settings menu
     */
    toggleSettings(): void {
      this.isSettingsMenuOpen = !this.isSettingsMenuOpen
      if (!this.isSettingsMenuOpen) {
        this.settingsMenuContent = null
      }
    },

    /**
     * Set settings menu content
     */
    setSettingsMenuContent(content: SettingsMenuContentType): void {
      this.settingsMenuContent = content
    },

    /**
     * Clear settings menu content
     */
    clearSettingsMenuContent(): void {
      this.settingsMenuContent = null
    },

    /**
     * Reset UI state
     */
    reset(): void {
      this.showStartScreen = true
      this.showingWinOverlay = false
      this.isShowAmountNotification = false
      this.anticipationMode = false
      this.initializationError = null
      this.isSettingsMenuOpen = false
      this.settingsMenuContent = null
    }
  }
})
