/**
 * Winning Store
 * Manages the current winning cycle state and tile animations
 */

import { defineStore } from 'pinia'
import { useTimingStore } from './timingStore'
import type { WinCombination } from '@/features/spin/types'
import type { Position } from '@/types/api'

/**
 * Winning tile states for visual control
 */
export const WINNING_STATES = {
  IDLE: 'idle',
  HIGHLIGHTED: 'highlighted',
  FLIPPING: 'flipping',
  FLIPPED: 'flipped',
  DISAPPEARING: 'disappearing'
} as const

/**
 * Type for winning states
 */
export type WinningState = typeof WINNING_STATES[keyof typeof WINNING_STATES]

/**
 * Winning store state interface
 */
export interface WinningStoreState {
  /** Current state of winning tiles (all tiles share the same state) */
  currentState: WinningState
  /** When the current state started (for animation timing) */
  stateStartTime: number
  /** Cell keys of tiles in the current win (e.g., ["0:1", "0:2", "1:1"]) */
  winningCellKeys: string[]
}

/**
 * Timing constants interface
 */
export interface WinningTimings {
  HIGHLIGHT_BEFORE_FLIP: number
  FLIP_DURATION: number
  FLIP_TO_DISAPPEAR: number
}

/**
 * Winning Store - Manages the current winning cycle state
 * Since the game processes one win at a time, this tracks the shared state
 * of all tiles in the current winning combination
 */
export const useWinningStore = defineStore('winning', {
  state: (): WinningStoreState => ({
    // Current state of winning tiles (all tiles share the same state)
    currentState: WINNING_STATES.IDLE,

    // When the current state started (for animation timing)
    stateStartTime: 0,

    // Cell keys of tiles in the current win (e.g., ["0:1", "0:2", "1:1"])
    winningCellKeys: []
  }),

  getters: {
    /**
     * Get timing constants from central timing store
     */
    TIMINGS(): WinningTimings {
      const timingStore = useTimingStore()
      return {
        HIGHLIGHT_BEFORE_FLIP: timingStore.HIGHLIGHT_BEFORE_FLIP,
        FLIP_DURATION: timingStore.FLIP_DURATION,
        FLIP_TO_DISAPPEAR: timingStore.DISAPPEAR_WAIT
      }
    },

    /**
     * Check if a specific cell is part of the current win
     * @returns Function that takes cellKey and returns boolean
     */
    isCellWinning(): (cellKey: string) => boolean {
      return (cellKey: string) => {
        return this.winningCellKeys.includes(cellKey)
      }
    },

    /**
     * Get the current state for a specific cell (only if it's winning)
     * @returns Function that takes cellKey and returns state
     */
    getCellState(): (cellKey: string) => WinningState {
      return (cellKey: string) => {
        if (this.winningCellKeys.includes(cellKey)) {
          return this.currentState
        }
        return WINNING_STATES.IDLE
      }
    },

    /**
     * Get time elapsed in current state
     */
    timeInCurrentState(): number {
      if (this.stateStartTime === 0) return 0
      return Date.now() - this.stateStartTime
    },

    /**
     * Check if any tiles are currently winning
     */
    hasWinningTiles(): boolean {
      return this.winningCellKeys.length > 0
    }
  },

  actions: {
    /**
     * Set winning tiles to HIGHLIGHTED state
     * @param cellKeys - Array of cell keys (e.g., ["0:1", "0:2"])
     */
    setHighlighted(cellKeys: string[]): void {
      this.currentState = WINNING_STATES.HIGHLIGHTED
      this.winningCellKeys = [...cellKeys]
      this.stateStartTime = Date.now()
    },

    /**
     * Transition to FLIPPING state
     */
    setFlipping(): void {
      if (this.currentState === WINNING_STATES.HIGHLIGHTED) {
        this.currentState = WINNING_STATES.FLIPPING
        this.stateStartTime = Date.now()
      }
    },

    /**
     * Transition to FLIPPED state
     */
    setFlipped(): void {
      if (this.currentState === WINNING_STATES.FLIPPING) {
        this.currentState = WINNING_STATES.FLIPPED
        this.stateStartTime = Date.now()
      }
    },

    /**
     * Transition to DISAPPEARING state
     */
    setDisappearing(): void {
      if (this.currentState === WINNING_STATES.FLIPPED || this.currentState === WINNING_STATES.HIGHLIGHTED) {
        this.currentState = WINNING_STATES.DISAPPEARING
        this.stateStartTime = Date.now()
      }
    },

    /**
     * Clear all winning state (when cascade starts or spinning begins)
     */
    clearWinningState(): void {
      this.currentState = WINNING_STATES.IDLE
      this.winningCellKeys = []
      this.stateStartTime = 0
    },

    /**
     * Helper: Convert win positions to cell keys
     * @param wins - Array of win objects with positions
     * @param bufferOffset - Buffer row offset
     * @returns Array of cell keys
     */
    winsToCellKeys(wins: WinCombination[], bufferOffset: number): string[] {
      const cellKeys: string[] = []
      if (!wins || wins.length === 0) return cellKeys

      wins.forEach(win => {
        if (!win.positions || win.positions.length === 0) return

        win.positions.forEach(pos => {
          // Backend sends positions as objects: {reel: 0, row: 1}
          const col = pos.reel
          const gridRow = pos.row
          const visualRow = gridRow - bufferOffset
          cellKeys.push(`${col}:${visualRow}`)
        })
      })

      return cellKeys
    }
  }
})
