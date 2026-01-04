import { isBonusTile } from '@/utils/tileHelpers'
import type { GridState } from '@/types/global'

export function countBonusTilesInColumn(
  gridState: GridState,
  col: number,
  winCheckStartRow: number,
  winCheckEndRow: number
): number {
  let count = 0
  for (let row = winCheckStartRow; row <= winCheckEndRow; row++) {
    if (isBonusTile(gridState.grid[col][row])) {
      count++
    }
  }
  return count
}

export function countTotalBonusTiles(
  gridState: GridState,
  stoppedCols: Set<number>,
  winCheckStartRow: number,
  winCheckEndRow: number
): number {
  let total = 0
  for (const col of stoppedCols) {
    total += countBonusTilesInColumn(gridState, col, winCheckStartRow, winCheckEndRow)
  }
  return total
}

export function countBonusAtStripPosition(
  strip: string[],
  candidateTarget: number,
  winCheckStartRow: number,
  winCheckEndRow: number
): number {
  let count = 0
  for (let gridRow = winCheckStartRow; gridRow <= winCheckEndRow; gridRow++) {
    const stripIdx = ((candidateTarget - gridRow) % strip.length + strip.length) % strip.length
    if (isBonusTile(strip[stripIdx])) {
      count++
    }
  }
  return count
}
