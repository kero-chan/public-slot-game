/**
 * Gold Transform Service
 * Pure business logic for gold tile to wild transformation
 * No Vue dependencies, fully testable
 */

import { istileWilden, isTileWildcard } from '@/utils/tileHelpers'
import type { Grid, WinCombination } from '../../spin/types'
import type { GoldTransformResult } from '../types'

/**
 * Transform gold tiles to wild in winning positions
 * Modifies the grid in place and filters win positions
 *
 * @param grid - 2D grid array (modified in place)
 * @param wins - Array of win objects with positions (modified in place)
 * @returns Transform result with positions and count
 *
 * @example
 * ```ts
 * const result = transformGoldTilesToWild(grid, wins)
 * console.log(`Transformed ${result.transformCount} gold tiles`)
 * ```
 */
export function transformGoldTilesToWild(
  grid: Grid,
  wins: WinCombination[]
): GoldTransformResult {
  const transformedPositions = new Set<string>()
  let transformCount = 0

  wins.forEach(win => {
    // Filter and transform positions
    win.positions = win.positions.filter(pos => {
      // Backend sends positions as objects: {reel, row}
      const col = pos.reel
      const row = pos.row
      const currentTile = grid[col][row]
      const isGolden = istileWilden(currentTile)
      const isWild = isTileWildcard(currentTile)

      if (isGolden && !isWild) {
        grid[col][row] = 'wild'
        transformedPositions.add(`${col},${row}`)
        transformCount++
        return false // Remove from positions
      }
      return true // Keep in positions
    })
  })

  return { transformedPositions, transformCount }
}
