/**
 * Spin Animation Module
 *
 * Handles reel strip building, spin animation, and bonus detection during spin.
 */

// Types
export * from './types'

// Strip building utilities
export {
  buildReelStrips,
  generateTargetIndexes,
  injectBackendGrid,
  injectBackendGridForColumn,
  validateStrips,
  syncColumnToGrid
} from './stripBuilder'

// Bonus detection during spin (visual only)
export {
  countBonusTilesInColumn,
  countTotalBonusTiles,
  getAnticipationThreshold,
  findBonusTilePositions
} from './bonusDetection'
