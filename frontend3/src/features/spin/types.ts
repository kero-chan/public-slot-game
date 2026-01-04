/**
 * Spin Feature Type Definitions
 */

import type { Ref } from 'vue'

/**
 * Grid verification result
 */
export interface GridVerification {
  allMatch: boolean
  mismatches: GridMismatch[]
  totalCells: number
}

/**
 * Grid mismatch details
 */
export interface GridMismatch {
  col: number
  row: number
  backend: string
  displayed: string
}

/**
 * Grid type - 2D array of symbol strings
 * Format: [column][row] where column is 0-4, row is 0-9
 */
export type Grid = string[][]

/**
 * Cascade data from backend
 */
export interface CascadeData {
  wins: WinCombination[]
  win_amount: number
  grid_after: Grid
  cascade_number?: number
  multiplier?: number  // Current multiplier for this cascade (from backend)
}

/**
 * Win combination from backend
 */
export interface WinCombination {
  symbol: number  // Symbol ID (same as grid values)
  count: number
  positions: WinPosition[]
  payout?: number
  win_intensity: 'small' | 'medium' | 'big' | 'mega'  // Visual intensity from backend
}

/**
 * Win position
 */
export interface WinPosition {
  reel: number  // column index (0-4)
  row: number   // row index in grid
}

/**
 * Backend spin response
 */
export interface SpinResponse {
  spin_id: string
  session_id: string
  bet_amount: number
  balance_before: number
  balance_after_bet: number
  new_balance: number
  grid: Grid
  cascades: CascadeData[]
  spin_total_win: number
  scatter_count: number
  is_free_spin: boolean
  free_spins_triggered: boolean
  free_spins_retriggered: boolean
  free_spins_additional?: number
  free_spins_session_id?: string
  free_spins_remaining_spins: number
  free_session_total_win: number
  timestamp: string
}

/**
 * Spin state composable return type
 */
export interface UseSpinState {
  // State
  backendTargetGrid: Ref<Grid | null>
  backendCascades: Ref<CascadeData[]>
  currentCascadeIndex: Ref<number>
  currentCascade: Ref<CascadeData | null>

  // Actions
  setBackendSpinResult: (response: SpinResponse) => void
  clearBackendSpinData: () => void
  hasMoreBackendCascades: () => boolean
  advanceToCascade: (index: number | null) => void
}
