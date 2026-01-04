/**
 * Spin Feature Module
 * Public API for spin-related functionality
 */

// Type exports
export type {
  Grid,
  GridVerification,
  GridMismatch,
  CascadeData,
  WinCombination,
  WinPosition,
  SpinResponse,
  UseSpinState,
} from './types'

// Service exports
export { validateGridIntegrity, verifyGridMatch, logVerification } from './services/spinService'

// Composable exports
export { useSpinState } from './composables/useSpinState'
