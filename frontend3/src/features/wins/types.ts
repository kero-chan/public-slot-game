/**
 * Wins Feature Type Definitions
 */

import type { UseSpinState, WinCombination } from '../spin/types'

/**
 * Win intensity levels
 */
export type WinIntensity = 'small' | 'medium' | 'big' | 'mega'

/**
 * Win intensity level mapping
 */
export interface WinIntensityLevels {
  small: number
  medium: number
  big: number
  mega: number
}

/**
 * Win detection functions type
 */
export interface WinDetection {
  getBackendWins: (spinState: UseSpinState) => WinCombination[]
  findWinningCombinations: (spinState: UseSpinState) => WinCombination[]
}
