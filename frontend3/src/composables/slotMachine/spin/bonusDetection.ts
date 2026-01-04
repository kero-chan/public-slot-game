import { isBonusTile } from '@/utils/tileHelpers'
import type { GridState } from '@/types/global'

// Visual threshold for anticipation animation (purely cosmetic)
const ANTICIPATION_VISUAL_THRESHOLD = 2

/**
 * Count bonus tiles in a column's visible rows
 */
export function countBonusTilesInColumn(
  gridState: GridState,
  col: number,
  winCheckStartRow: number,
  winCheckEndRow: number
): number {
  let count = 0

  for (let row = winCheckStartRow; row <= winCheckEndRow; row++) {
    const cell = gridState.grid[col][row]
    if (isBonusTile(cell)) {
      count++
    }
  }

  return count
}

/**
 * Count total bonus tiles across specified columns
 */
export function countTotalBonusTiles(
  gridState: GridState,
  stoppedColumns: Set<number>,
  winCheckStartRow: number,
  winCheckEndRow: number
): number {
  let total = 0

  for (const col of stoppedColumns) {
    total += countBonusTilesInColumn(gridState, col, winCheckStartRow, winCheckEndRow)
  }

  return total
}

/**
 * Get anticipation visual threshold (for animation only)
 */
export function getAnticipationThreshold(): number {
  return ANTICIPATION_VISUAL_THRESHOLD
}

/**
 * Find bonus tile positions in all columns
 */
export function findBonusTilePositions(
  gridState: GridState,
  cols: number,
  winCheckStartRow: number,
  winCheckEndRow: number
): Array<{ col: number; row: number }> {
  const positions: Array<{ col: number; row: number }> = []

  for (let col = 0; col < cols; col++) {
    for (let row = winCheckStartRow; row <= winCheckEndRow; row++) {
      const cell = gridState.grid[col][row]
      if (isBonusTile(cell)) {
        positions.push({ col, row })
      }
    }
  }

  return positions
}
