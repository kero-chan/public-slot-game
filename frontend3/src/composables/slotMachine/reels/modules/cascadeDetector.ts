import type { Sprite } from 'pixi.js'
import type { GridState } from '@/types/global'
import { CONFIG } from '@/config/constants'
import type { DropAnimationManager } from '../dropAnimation'

export interface CascadeDetectorParams {
  gridState: GridState
  dropAnimations: DropAnimationManager
  spriteCache: Map<string, Sprite>
  COLS: number
  ROWS_FULL: number
  BUFFER_OFFSET: number
  startY: number
  stepY: number
  originX: number
  stepX: number
  scaledTileH: number
  BLEED: number
  getTextureForSymbol: (symbol: string) => any
}

export function detectCascadeAndStartDrops(params: CascadeDetectorParams): void {
  const {
    gridState,
    dropAnimations,
    spriteCache,
    COLS,
    ROWS_FULL,
    BUFFER_OFFSET,
    startY,
    stepY,
    originX,
    stepX,
    scaledTileH,
    BLEED,
    getTextureForSymbol
  } = params

  const removedPositions = gridState.lastRemovedPositions || new Set()
  if (!gridState.previousGridSnapshot || removedPositions.size === 0) return

  // Clear completed animation states from previous cascade
  dropAnimations.clearCompleted()

  for (let col = 0; col < COLS; col++) {
    const oldCol = gridState.previousGridSnapshot[col] || []
    const newCol = gridState.grid[col] || []
    const totalRows = oldCol.length

    // Count removed tiles from GAME area
    let removedCount = 0
    for (let gridRow = BUFFER_OFFSET; gridRow < totalRows; gridRow++) {
      if (removedPositions.has(`${col},${gridRow}`)) {
        removedCount++
      }
    }

    if (removedCount === 0) continue

    // Collect kept GAME tiles
    const keptGameTiles: Array<{ oldGridRow: number; symbol: string }> = []
    for (let row = totalRows - 1; row >= BUFFER_OFFSET; row--) {
      if (!removedPositions.has(`${col},${row}`)) {
        keptGameTiles.unshift({ oldGridRow: row, symbol: oldCol[row] })
      }
    }

    // Tiles from bottom of buffer
    const takeFromBuffer: Array<{ oldGridRow: number; symbol: string }> = []
    for (let i = BUFFER_OFFSET - removedCount; i < BUFFER_OFFSET; i++) {
      if (i >= 0) {
        takeFromBuffer.push({ oldGridRow: i, symbol: oldCol[i] })
      }
    }

    // Remaining buffer tiles
    const remainingBuffer: Array<{ oldGridRow: number; symbol: string }> = []
    for (let i = 0; i < BUFFER_OFFSET - removedCount; i++) {
      remainingBuffer.push({ oldGridRow: i, symbol: oldCol[i] })
    }

    let newGridRow = removedCount // Skip new random tiles

    // Process remaining buffer tiles
    remainingBuffer.forEach((tile) => {
      const oldGridRow = tile.oldGridRow
      if (oldGridRow !== newGridRow) {
        const oldVisualRow = oldGridRow - BUFFER_OFFSET
        const newVisualRow = newGridRow - BUFFER_OFFSET

        if (newVisualRow >= 0 && newVisualRow < ROWS_FULL) {
          startDropAnimation(col, newVisualRow, oldVisualRow, newGridRow, newCol)
        }
      }
      newGridRow++
    })

    // Process buffer tiles moving into game
    takeFromBuffer.forEach((tile) => {
      const oldGridRow = tile.oldGridRow
      const oldVisualRow = oldGridRow - BUFFER_OFFSET
      const newVisualRow = newGridRow - BUFFER_OFFSET

      if (newVisualRow >= 0 && newVisualRow < ROWS_FULL) {
        startDropAnimation(col, newVisualRow, oldVisualRow, newGridRow, newCol)
      }
      newGridRow++
    })

    // Process kept game tiles
    keptGameTiles.forEach((tile) => {
      const oldGridRow = tile.oldGridRow
      if (oldGridRow !== newGridRow) {
        const oldVisualRow = oldGridRow - BUFFER_OFFSET
        const newVisualRow = newGridRow - BUFFER_OFFSET

        if (newVisualRow >= 0 && newVisualRow < ROWS_FULL) {
          const gridSymbol = newCol[newGridRow]
          const isFullyVisible =
            newVisualRow >= CONFIG.reels.visualWinStartRow && newVisualRow <= CONFIG.reels.visualWinEndRow

          if (isFullyVisible) {
            startDropAnimation(col, newVisualRow, oldVisualRow, newGridRow, newCol)
          } else {
            // Partial rows: no drop animation
            const cellKey = `${col}:${newVisualRow}`
            const sprite = spriteCache.get(cellKey)
            if (sprite) {
              const tex = getTextureForSymbol(gridSymbol)
              if (tex && sprite.texture !== tex) {
                sprite.texture = tex
              }
              const newY = startY + newVisualRow * stepY - BLEED + (scaledTileH + BLEED * 2) / 2
              sprite.y = newY
            }
          }
        }
      }
      newGridRow++
    })
  }

  function startDropAnimation(
    col: number,
    newVisualRow: number,
    oldVisualRow: number,
    newGridRow: number,
    newCol: string[]
  ): void {
    const cellKey = `${col}:${newVisualRow}`
    const sprite = spriteCache.get(cellKey)
    if (sprite) {
      const oldY = startY + oldVisualRow * stepY - BLEED + (scaledTileH + BLEED * 2) / 2
      const newY = startY + newVisualRow * stepY - BLEED + (scaledTileH + BLEED * 2) / 2
      const gridSymbol = newCol[newGridRow]
      dropAnimations.startDrop(cellKey, sprite, oldY, newY, gridSymbol, getTextureForSymbol)
    }
  }
}
