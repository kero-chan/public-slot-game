import { CONFIG } from '@/config/constants'
import { getRandomSymbol } from '@/utils/gameHelpers'
import { isBonusTile } from '@/utils/tileHelpers'
import type { GridState } from '@/types/global'
import type { StripBuilderOptions } from './types'

/**
 * Build reel strips from current grid state
 * Places current grid symbols at positions the renderer will read at reelTop=0
 */
export function buildReelStrips(options: StripBuilderOptions): void {
  const {
    gridState,
    cols,
    totalRows,
    stripLength,
    winCheckStartRow,
    winCheckEndRow,
    maxBonusPerColumn
  } = options

  // Build protected positions set - these are the current grid symbols that must not change
  const protectedPositions = new Map<number, Set<number>>()

  for (let col = 0; col < cols; col++) {
    // Initialize strip with random symbols
    const strip: string[] = Array(stripLength).fill(null).map(() =>
      getRandomSymbol({ col, allowGold: true, allowBonus: true })
    )

    // Place current grid symbols at positions renderer will read when reelTop=0
    const reelTopAtStart = 0
    const protectedSet = new Set<number>()

    for (let row = 0; row < totalRows; row++) {
      const stripIdx = ((reelTopAtStart - row) % stripLength + stripLength) % stripLength
      const currentSymbol = gridState.grid?.[col]?.[row]

      if (currentSymbol && typeof currentSymbol === 'string') {
        strip[stripIdx] = currentSymbol
      }
      protectedSet.add(stripIdx)
    }

    protectedPositions.set(col, protectedSet)
    gridState.reelStrips[col] = strip
  }

  // Enforce bonus limits in the landing area of each strip
  enforceBonusLimits(gridState, cols, stripLength, protectedPositions, winCheckStartRow, winCheckEndRow, maxBonusPerColumn)

  // Trigger reactivity
  gridState.reelStrips = [...gridState.reelStrips]
}

/**
 * Enforce maximum bonus tiles per column in landing positions
 */
function enforceBonusLimits(
  gridState: GridState,
  cols: number,
  stripLength: number,
  protectedPositions: Map<number, Set<number>>,
  winCheckStartRow: number,
  winCheckEndRow: number,
  maxBonusPerColumn: number
): void {
  for (let col = 0; col < cols; col++) {
    const strip = gridState.reelStrips[col]
    const protectedSet = protectedPositions.get(col)!

    for (let reelTop = 0; reelTop < stripLength; reelTop++) {
      const bonusPositions: Array<{ idx: number; gridRow: number }> = []

      for (let gridRow = winCheckStartRow; gridRow <= winCheckEndRow; gridRow++) {
        const stripIdx = ((reelTop - gridRow) % strip.length + strip.length) % strip.length
        if (isBonusTile(strip[stripIdx])) {
          bonusPositions.push({ idx: stripIdx, gridRow })
        }
      }

      if (bonusPositions.length > maxBonusPerColumn) {
        let replaced = 0
        const requiredReplacements = bonusPositions.length - maxBonusPerColumn

        for (let i = bonusPositions.length - 1; i >= 0 && replaced < requiredReplacements; i--) {
          const { idx, gridRow } = bonusPositions[i]
          if (protectedSet.has(idx)) continue

          const visualRow = gridRow - CONFIG.reels.bufferRows
          strip[idx] = getRandomSymbol({ col, visualRow, allowGold: true, allowBonus: false })
          replaced++
        }
      }
    }
  }
}

/**
 * Generate random landing target indexes
 */
export function generateTargetIndexes(cols: number, totalRows: number, stripLength: number): number[] {
  const minLanding = totalRows + 10
  const maxLanding = stripLength - totalRows

  return Array(cols).fill(0).map(() =>
    Math.floor(minLanding + Math.random() * (maxLanding - minLanding))
  )
}

/**
 * Inject backend grid data at target landing positions
 */
export function injectBackendGrid(
  backendGrid: string[][] | null,
  targetIndexes: number[],
  gridState: GridState,
  cols: number,
  stripLength: number
): boolean {
  if (!backendGrid || backendGrid.length !== cols) {
    return false
  }

  const backendRows = CONFIG.reels.rows

  for (let col = 0; col < cols; col++) {
    const targetReelTop = targetIndexes[col]
    const targetColumn = backendGrid[col]
    const strip = gridState.reelStrips[col]

    if (targetColumn && targetColumn.length >= backendRows) {
      // Fill intermediate positions with random symbols
      for (let stripPos = 1; stripPos < targetReelTop; stripPos++) {
        if (strip[stripPos] === null) {
          strip[stripPos] = getRandomSymbol({ col, visualRow: stripPos % 4, allowGold: true, allowBonus: false })
        }
      }

      // Place backend target grid at landing position
      for (let row = 0; row < backendRows; row++) {
        const stripIdx = ((targetReelTop - row) % stripLength + stripLength) % stripLength
        strip[stripIdx] = targetColumn[row]
      }
    }
  }

  return true
}

/**
 * Inject backend grid for a single column at a specific position
 */
export function injectBackendGridForColumn(
  backendGrid: string[][] | null,
  col: number,
  targetPosition: number,
  gridState: GridState,
  cols: number,
  stripLength: number
): boolean {
  if (!backendGrid || backendGrid.length !== cols) {
    return false
  }

  const targetColumn = backendGrid[col]
  const strip = gridState.reelStrips[col]
  const backendRows = CONFIG.reels.rows

  if (!targetColumn || targetColumn.length < backendRows) {
    return false
  }

  for (let row = 0; row < backendRows; row++) {
    const stripIdx = ((targetPosition - row) % stripLength + stripLength) % stripLength
    strip[stripIdx] = targetColumn[row]
  }

  return true
}

/**
 * Validate strips have no null/empty values
 */
export function validateStrips(gridState: GridState, cols: number, stripLength: number): void {
  for (let col = 0; col < cols; col++) {
    const strip = gridState.reelStrips[col]

    for (let i = 0; i < stripLength; i++) {
      if (!strip[i] || strip[i] === null || strip[i] === undefined || strip[i] === '') {
        strip[i] = getRandomSymbol({ col, visualRow: i % 4, allowGold: true, allowBonus: false })
      }
    }
  }
}

/**
 * Sync a column from strip to grid
 */
export function syncColumnToGrid(gridState: GridState, col: number, bufferOffset: number): void {
  const totalRows = CONFIG.reels.rows + bufferOffset
  const reelTop = gridState.reelTopIndex[col]
  const strip = gridState.reelStrips[col]

  for (let row = 0; row < totalRows; row++) {
    const stripIdx = ((reelTop - row) % strip.length + strip.length) % strip.length
    gridState.grid[col][row] = strip[stripIdx]
  }
}
