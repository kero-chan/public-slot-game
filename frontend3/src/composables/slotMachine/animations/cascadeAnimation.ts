import type { GridState } from '@/types/global'

/**
 * Run cascade drop animation
 * Waits for the drop animation to complete
 */
export function animateCascade(
  gridState: GridState,
  render: () => void,
  maxWait: number
): Promise<void> {
  const startTime = Date.now()

  return new Promise(resolve => {
    const tick = (): void => {
      // Render to update isDropAnimating flag
      render()

      const elapsed = Date.now() - startTime
      const shouldWait = gridState.isDropAnimating

      if (shouldWait && elapsed < maxWait) {
        requestAnimationFrame(tick)
      } else {
        resolve()
      }
    }

    tick()
  })
}

/**
 * Clear cascade animation state
 */
export function clearCascadeState(gridState: GridState): void {
  gridState.previousGridSnapshot = null
  gridState.lastRemovedPositions = new Set()
  gridState.bufferRows = []
  gridState.isDropAnimating = false
}
