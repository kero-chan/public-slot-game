/**
 * Bonus Detector Service
 * Pure business logic for bonus tile detection
 * No Vue dependencies, fully testable
 *
 * SECURITY NOTE: Bonus trigger decisions are made by the BACKEND.
 * Frontend functions here are for VISUAL PURPOSES ONLY.
 */

import { isBonusTile } from '@/utils/tileHelpers'
import { CONFIG } from '@/config/constants'
import type { Grid } from '../../spin/types'

/**
 * ⚠️ VISUAL-ONLY FUNCTION - NOT FOR GAME DECISIONS
 *
 * Count bonus tiles in visible rows for VISUAL PURPOSES ONLY:
 * - Highlighting bonus tiles during anticipation
 * - Visual feedback animations
 * - UI display of bonus count
 *
 * ❌ DO NOT USE THIS FOR GAME LOGIC:
 * - Backend determines if free spins are triggered via `response.free_spins_triggered`
 * - Backend determines free spins count via `response.free_spins_remaining_spins`
 *
 * @param grid - 2D grid array
 * @param bonusCheckStartRow - Start row for checking (inclusive)
 * @param bonusCheckEndRow - End row for checking (inclusive)
 * @returns Count of bonus tiles found (for visual display only)
 *
 * @example
 * ```ts
 * // CORRECT: Use for visual highlighting
 * const count = checkBonusTiles(grid, 2, 7)
 * highlightBonusTiles(count)
 *
 * // WRONG: Do NOT use for game decisions
 * // if (count >= 3) triggerFreeSpins() // ❌ FORBIDDEN
 * ```
 */
export function checkBonusTiles(
  grid: Grid,
  bonusCheckStartRow: number,
  bonusCheckEndRow: number
): number {
  let bonusCount = 0

  // Only check the fully visible rows (excluding partially visible top and bottom)
  for (let col = 0; col < CONFIG.reels.count; col++) {
    for (let row = bonusCheckStartRow; row <= bonusCheckEndRow; row++) {
      const cell = grid[col][row]
      if (isBonusTile(cell)) {
        bonusCount++
      }
    }
  }

  return bonusCount
}
