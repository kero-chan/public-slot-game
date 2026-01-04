import type { GridState, WinCombination } from '@/types/global'

/**
 * Highlight animation state
 */
interface HighlightState {
  active: boolean
  stopRequested: boolean
}

/**
 * Create highlight animation controller
 */
export function createHighlightAnimation(gridState: GridState) {
  const state: HighlightState = {
    active: false,
    stopRequested: false
  }

  /**
   * Run highlight animation for wins
   */
  function animate(wins: WinCombination[], duration: number): Promise<void> {
    const startTime = Date.now()
    gridState.highlightAnim = { start: startTime, duration }
    state.active = true
    state.stopRequested = false

    return new Promise(resolve => {
      const tick = (): void => {
        const elapsed = Date.now() - startTime

        if (state.stopRequested || elapsed >= duration) {
          gridState.highlightWins = null
          gridState.highlightAnim = { start: 0, duration: 0 }
          state.active = false
          state.stopRequested = false
          resolve()
          return
        }

        gridState.highlightWins = wins
        requestAnimationFrame(tick)
      }

      tick()
    })
  }

  /**
   * Stop the animation early
   */
  function stop(): void {
    if (state.active) {
      state.stopRequested = true
    }
  }

  /**
   * Check if animation is active
   */
  function isActive(): boolean {
    return state.active
  }

  return {
    animate,
    stop,
    isActive
  }
}

export type HighlightAnimation = ReturnType<typeof createHighlightAnimation>
