import { CONFIG } from '@/config/constants'
import type { GridState } from '@/types/global'
import type { SpinConfig } from './types'

export function syncColumnToGrid(
  gridState: GridState,
  col: number,
  totalRows: number
): void {
  const reelTop = gridState.reelTopIndex[col]
  const strip = gridState.reelStrips[col]

  for (let row = 0; row < totalRows; row++) {
    const stripIdx = ((reelTop - row) % strip.length + strip.length) % strip.length
    gridState.grid[col][row] = strip[stripIdx]
  }
}

export function injectBackendGrid(
  gridState: GridState,
  backendTargetGrid: string[][] | null,
  targetIndexes: number[],
  config: SpinConfig,
  getRandomSymbol: (opts: any) => string
): boolean {
  const { cols, stripLength } = config

  if (!backendTargetGrid || backendTargetGrid.length !== cols) {
    return false
  }

  for (let col = 0; col < cols; col++) {
    const targetReelTop = targetIndexes[col]
    const targetColumn = backendTargetGrid[col]
    const strip = gridState.reelStrips[col]
    const backendRows = CONFIG.reels.rows

    if (targetColumn && targetColumn.length >= backendRows) {
      for (let stripPos = 1; stripPos < targetReelTop; stripPos++) {
        if (strip[stripPos] === null) {
          strip[stripPos] = getRandomSymbol({ col, visualRow: stripPos % 4, allowGold: true, allowBonus: false })
        }
      }

      for (let row = 0; row < backendRows; row++) {
        const stripIdx = ((targetReelTop - row) % stripLength + stripLength) % stripLength
        strip[stripIdx] = targetColumn[row]
      }
    }
  }

  return true
}

export function injectBackendGridForColumn(
  gridState: GridState,
  backendTargetGrid: string[][] | null,
  col: number,
  newTargetPosition: number,
  config: SpinConfig
): boolean {
  const { cols, stripLength } = config

  if (!backendTargetGrid || backendTargetGrid.length !== cols) {
    return false
  }

  const targetColumn = backendTargetGrid[col]
  const strip = gridState.reelStrips[col]
  const backendRows = CONFIG.reels.rows

  if (!targetColumn || targetColumn.length < backendRows) {
    return false
  }

  for (let row = 0; row < backendRows; row++) {
    const stripIdx = ((newTargetPosition - row) % stripLength + stripLength) % stripLength
    strip[stripIdx] = targetColumn[row]
  }

  return true
}
