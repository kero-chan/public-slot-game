import { getBufferOffset } from '@/utils/gameHelpers'
import { useTimingStore } from '@/stores'
import type { GridState, WinCombination } from '@/types/global'
import type { UseReels } from '@/composables/slotMachine/reels'

export interface GoldTransformConfig {
  gridState: GridState
  getReels: () => UseReels | null
}

export interface UseGoldTransform {
  transformGoldTilesToWild: (wins: WinCombination[]) => Promise<void>
  startGoldTransformVisuals: (wins: WinCombination[]) => void
}

/**
 * Gold to Wild transformation handler
 *
 * IMPORTANT: Frontend does NOT detect which tiles to transform.
 * It reads the positions from gridState.goldTransformedPositions which is
 * populated by comparing current grid with backend's grid_after data.
 * This ensures all transformation logic is driven by the backend.
 */
export function useGoldTransform(config: GoldTransformConfig): UseGoldTransform {
  const timingStore = useTimingStore()
  const BUFFER_OFFSET = getBufferOffset()
  const { gridState, getReels } = config

  /**
   * Update a single grid cell from gold to wild
   */
  function updateGridCell(col: number, row: number): void {
    // Transform gold tile to wild (e.g., "fa_gold" -> "wild")
    gridState.grid[col][row] = 'wild'
  }

  /**
   * Transform gold tiles to wild based on backend-determined positions
   * Positions are read from gridState.goldTransformedPositions (set by cascadeSymbols)
   */
  function transformGoldTilesToWild(wins: WinCombination[]): Promise<void> {
    const reels = getReels()

    // Use positions from gridState.goldTransformedPositions (determined by backend cascade data)
    const transformPositions = gridState.goldTransformedPositions
    if (!transformPositions || transformPositions.size === 0) {
      return Promise.resolve()
    }

    const goldPositions: Array<{ col: number; row: number; cellKey: string }> = []

    // Parse positions from the Set (format: "col,row")
    transformPositions.forEach(posKey => {
      const [colStr, rowStr] = posKey.split(',')
      const col = parseInt(colStr, 10)
      const row = parseInt(rowStr, 10)
      const visualRow = row - BUFFER_OFFSET
      goldPositions.push({ col, row, cellKey: `${col}:${visualRow}` })
    })

    if (goldPositions.length > 0 && reels) {
      const spriteCache = reels.getSpriteCache()
      const goldToWildAnimations = reels.goldToWildAnimations

      goldPositions.forEach(({ col, row, cellKey }) => {
        const sprite = spriteCache.get(cellKey)
        if (sprite) {
          // Pass callback to update grid at midpoint when texture swaps
          goldToWildAnimations.startTransform(cellKey, col, row, sprite, sprite.x, sprite.y, updateGridCell)
        }
      })

      // Wait for animation to complete
      return new Promise(resolve => {
        setTimeout(resolve, timingStore.GOLD_TRANSFORM_DURATION)
      })
    }

    return Promise.resolve()
  }

  /**
   * Start gold transform visuals based on backend-determined positions
   * Positions are read from gridState.goldTransformedPositions
   */
  function startGoldTransformVisuals(wins: WinCombination[]): void {
    const reels = getReels()
    if (!reels) return

    // Use positions from gridState.goldTransformedPositions (determined by backend cascade data)
    const transformPositions = gridState.goldTransformedPositions
    if (!transformPositions || transformPositions.size === 0) {
      return
    }

    const spriteCache = reels.getSpriteCache()
    const goldToWildAnimations = reels.goldToWildAnimations

    // Parse positions from the Set (format: "col,row")
    transformPositions.forEach(posKey => {
      const [colStr, rowStr] = posKey.split(',')
      const col = parseInt(colStr, 10)
      const row = parseInt(rowStr, 10)
      const visualRow = row - BUFFER_OFFSET
      const cellKey = `${col}:${visualRow}`
      const sprite = spriteCache.get(cellKey)
      if (sprite && !goldToWildAnimations.isAnimating(cellKey)) {
        // Pass callback to update grid at midpoint when texture swaps
        goldToWildAnimations.startTransform(cellKey, col, row, sprite, sprite.x, sprite.y, updateGridCell)
      }
    })
  }

  return { transformGoldTilesToWild, startGoldTransformVisuals }
}
