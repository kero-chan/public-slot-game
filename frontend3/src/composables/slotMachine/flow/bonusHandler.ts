import { CONFIG } from '@/config/constants'
import type { GridState } from '@/types/global'
import type { GameLogicAPI } from './types'

/**
 * Handle bonus tile checking
 *
 * SECURITY: All bonus trigger decisions come from backend via response.free_spins_triggered
 */
export async function handleCheckBonus(
  gameLogic: GameLogicAPI,
  gridState: GridState,
  gameState: any,
  winningStore: any,
  gameStore: any
): Promise<void> {
  let freeSpinsTriggered = false
  let freeSpinsAwarded: number | undefined

  if (gameState?.backendSpinResponse) {
    const response = gameState.backendSpinResponse
    freeSpinsTriggered = response.free_spins_triggered || false

    if (freeSpinsTriggered && response.free_spins_remaining_spins) {
      freeSpinsAwarded = response.free_spins_remaining_spins

      const bonusCellKeys = findBonusCellKeys(gridState)
      if (bonusCellKeys.length > 0) {
        winningStore.setHighlighted(bonusCellKeys)
      }
      winningStore.clearWinningState()
    }
  }

  gameStore.setBonusResults(freeSpinsTriggered, freeSpinsAwarded)
}

/**
 * Find all bonus tile cell keys in visible rows
 */
function findBonusCellKeys(gridState: GridState): string[] {
  const bonusCellKeys: string[] = []
  const bufferRows = CONFIG.reels.bufferRows
  const fullyVisibleRows = CONFIG.reels.fullyVisibleRows
  // Include partial rows above and below the fully visible area
  const VISIBLE_START_ROW = Math.max(0, bufferRows - 1)
  const VISIBLE_END_ROW = Math.min(CONFIG.reels.rows - 1, bufferRows + fullyVisibleRows)

  for (let col = 0; col < CONFIG.reels.count; col++) {
    for (let row = VISIBLE_START_ROW; row <= VISIBLE_END_ROW; row++) {
      const cell = gridState.grid[col][row]
      if (cell === 'bonus') {
        const visualRow = row - bufferRows
        bonusCellKeys.push(`${col}:${visualRow}`)
      }
    }
  }

  return bonusCellKeys
}
