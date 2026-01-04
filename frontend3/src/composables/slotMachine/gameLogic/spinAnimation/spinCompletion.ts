import type { GridState } from '@/types/global'

export interface SpinCompletionHandler {
  complete: () => void
  isCompleted: () => boolean
}

export function createSpinCompletionHandler(
  gridState: GridState,
  gameStore: any,
  resolve: () => void
): SpinCompletionHandler {
  let hasResolved = false

  const complete = (): void => {
    if (hasResolved) return
    hasResolved = true

    if (gameStore.anticipationMode) {
      gameStore.deactivateAnticipationMode()
    }

    gridState.activeSlowdownColumn = -1
    gridState.grid = [...gridState.grid.map(col => [...col])]
    resolve()
  }

  const isCompleted = (): boolean => hasResolved

  return { complete, isCompleted }
}
