/**
 * Grid Store
 * Manages the game grid, reel strips, and spin animation state
 */

import { defineStore } from 'pinia'
import { createEmptyGrid, createReelStrips } from '@/utils/gameHelpers'
import { CONFIG } from '@/config/constants'
import { gameApi } from '@/api/game'
import { mapBackendGridToFrontend } from '@/utils/gridMapper'
import type { Grid } from '@/features/spin/types'
import type { WinCombination } from '@/features/spin/types'

/**
 * Animation state for highlight and disappear effects
 */
export interface AnimationState {
  /** Animation start timestamp */
  start: number
  /** Animation duration in ms */
  duration: number
}

/**
 * Grid store state interface
 */
export interface GridStoreState {
  /** Main grid: 5 columns × 6 rows (backend sends all 6 rows) */
  grid: Grid
  /** Loading state for grid fetch */
  gridLoading: boolean
  /** Error state for grid fetch */
  gridError: Error | null

  /** Reel strips for spinning animation (longer strip = smoother animation) */
  reelStrips: string[][]
  /** Top index for each reel strip */
  reelTopIndex: number[]

  /** Spin animation offset for each column */
  spinOffsets: number[]
  /** Spin velocity for each column */
  spinVelocities: number[]
  /** Explicit completion flags (true = use grid, false = use reelStrips) */
  columnsCompleted: boolean[]
  /** Per-column convergence flag */
  convergenceMode: boolean[]

  /** DEPRECATED - Win highlight animation (now managed by winningStore) */
  highlightWins: WinCombination[] | null
  /** DEPRECATED - Highlight animation state */
  highlightAnim: AnimationState

  /** DEPRECATED - Disappear animation (now managed by winningStore) */
  disappearPositions: Set<string>
  /** DEPRECATED - Disappear animation state */
  disappearAnim: AnimationState

  /** Timestamp of last cascade */
  lastCascadeTime: number
  /** Set of removed position keys from last cascade */
  lastRemovedPositions: Set<string>
  /** Set of positions where gold tiles transformed to wild (should not cascade) */
  goldTransformedPositions: Set<string>
  /** Whether drop animation is currently running */
  isDropAnimating: boolean
  /** Snapshot of grid before cascade for drop animation */
  previousGridSnapshot: Grid | null
}

/**
 * Grid Store - manages game grid and spin animations
 */
export const useGridStore = defineStore('grid', {
  state: (): GridStoreState => ({
    // Main grid: 5 columns × 6 rows = 5 × 6
    // Backend sends all 6 rows (no buffer rows needed - backend handles cascades)
    // Initialized with empty grid, will be replaced by backend grid on load
    grid: createEmptyGrid(),
    gridLoading: false,
    gridError: null,

    // Reel strips for spinning animation (longer strip = smoother animation)
    reelStrips: createReelStrips(CONFIG.reels.count, CONFIG.reels.stripLength),
    reelTopIndex: Array(CONFIG.reels.count).fill(0),

    // Spin animation state
    spinOffsets: Array(CONFIG.reels.count).fill(0),
    spinVelocities: Array(CONFIG.reels.count).fill(0),
    columnsCompleted: Array(CONFIG.reels.count).fill(true), // Explicit completion flags (true = use grid, false = use reelStrips)
    convergenceMode: Array(CONFIG.reels.count).fill(false), // Per-column convergence flag

    // Win highlight animation (DEPRECATED - now managed by winningStore)
    highlightWins: null,
    highlightAnim: { start: 0, duration: 0 },

    // Disappear animation (DEPRECATED - now managed by winningStore)
    disappearPositions: new Set(),
    disappearAnim: { start: 0, duration: 0 },

    // Cascade/drop animation state
    lastCascadeTime: 0,
    lastRemovedPositions: new Set(),
    goldTransformedPositions: new Set(),
    isDropAnimating: false,
    previousGridSnapshot: null
  }),

  actions: {
    /**
     * Fetch initial grid from backend
     * Ensures zero frontend RNG - all symbol generation is backend-controlled
     *
     * Backend handles ALL game logic including cascades and drops.
     * Frontend only displays what backend provides (6 rows per column).
     */
    async fetchInitialGrid(): Promise<void> {
      this.gridLoading = true
      this.gridError = null

      try {
        const response = await gameApi.getInitialGrid()

        console.log('📥 Initial grid from server:', JSON.stringify(response.grid))

        // Convert backend grid to frontend format
        // Backend provides 6 rows per column, frontend displays them as-is
        this.grid = mapBackendGridToFrontend(response.grid)

        console.log('📤 Frontend grid after mapping:', JSON.stringify(this.grid))
      } catch (error) {
        this.gridError = error as Error

        // Fallback to local generation (for offline development)
        this.grid = createEmptyGrid()
      } finally {
        this.gridLoading = false
      }
    },

    /**
     * Reset grid to empty state
     */
    resetGrid(): void {
      this.grid = createEmptyGrid()
    },

    /**
     * Reset spin animation state
     */
    resetSpinState(): void {
      this.spinOffsets = Array(CONFIG.reels.count).fill(0)
      this.spinVelocities = Array(CONFIG.reels.count).fill(0)
      this.convergenceMode = Array(CONFIG.reels.count).fill(false)
    },

    /**
     * Clear highlight animation state
     */
    clearHighlights(): void {
      this.highlightWins = null
      this.highlightAnim = { start: 0, duration: 0 }
    },

    /**
     * Clear disappear animation state
     */
    clearDisappear(): void {
      this.disappearPositions = new Set()
      this.disappearAnim = { start: 0, duration: 0 }
    },

    /**
     * Regenerate reel strips with new random symbols
     */
    regenerateReelStrips(): void {
      this.reelStrips = createReelStrips(CONFIG.reels.count, CONFIG.reels.stripLength)
      this.reelTopIndex = Array(CONFIG.reels.count).fill(0)
    }
  }
})
