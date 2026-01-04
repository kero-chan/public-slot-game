/**
 * Wins Feature Module
 * Public API for win detection and calculation
 */

// Type exports
export type { WinIntensity, WinIntensityLevels, WinDetection } from './types'

// Service exports
export { getWinIntensity, validateBackendWins } from './services/winValidator'

// Composable exports
export {
  getBackendWins,
  findWinningCombinations,
  findWinningCombinations_DEPRECATED_FORBIDDEN,
} from './composables/useWinDetection'
