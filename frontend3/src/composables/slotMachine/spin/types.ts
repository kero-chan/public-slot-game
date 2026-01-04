import type { GridState, SpinResponse } from '@/types/global'

/**
 * Animation object for tracking column position during spin
 */
export interface AnimObject {
  col: number
  position: number
  targetIndex: number
  isSlowingDown: boolean
  hasSlowedDown: boolean
}

/**
 * Spin animation configuration
 */
export interface SpinConfig {
  baseDuration: number
  stagger: number
  totalRows: number
  cols: number
  stripLength: number
  anticipationSlowdownPerColumn: number
  maxBonusPerColumn: number
}

/**
 * Spin animation callbacks
 */
export interface SpinCallbacks {
  onColumnStop: (col: number) => void
  onAnticipationActivate: (col: number) => void
  onComplete: () => void
  playEffect: (effectName: string) => void
}

/**
 * Reel strip builder options
 */
export interface StripBuilderOptions {
  gridState: GridState
  cols: number
  totalRows: number
  stripLength: number
  bufferOffset: number
  winCheckStartRow: number
  winCheckEndRow: number
  maxBonusPerColumn: number
}

/**
 * Backend grid injection options
 */
export interface BackendInjectionOptions {
  backendGrid: string[][] | null
  targetIndexes: number[]
  gridState: GridState
  cols: number
  stripLength: number
}
