import { useTimingStore } from '@/stores'
import type { GridState, WinCombination } from '@/types/global'

export interface HighlightAnimationConfig {
  gridState: GridState
}

export interface UseHighlightAnimation {
  highlightWinsAnimation: (wins: WinCombination[]) => Promise<void>
  stopHighlightAnimation: () => void
  isAnimating: () => boolean
}

export function useHighlightAnimation(config: HighlightAnimationConfig): UseHighlightAnimation {
  const timingStore = useTimingStore()
  const { gridState } = config

  let highlightAnimationActive = false
  let stopHighlightRequested = false

  function isAnimating(): boolean {
    return highlightAnimationActive
  }

  function highlightWinsAnimation(wins: WinCombination[]): Promise<void> {
    const duration = timingStore.HIGHLIGHT_ANIMATION_DURATION
    const startTime = Date.now()
    gridState.highlightAnim = { start: startTime, duration }
    highlightAnimationActive = true
    stopHighlightRequested = false

    return new Promise(resolve => {
      const animate = (): void => {
        const elapsed = Date.now() - startTime

        if (stopHighlightRequested || elapsed >= duration) {
          gridState.highlightWins = null
          gridState.highlightAnim = { start: 0, duration: 0 }
          highlightAnimationActive = false
          stopHighlightRequested = false
          resolve()
          return
        }

        gridState.highlightWins = wins
        requestAnimationFrame(animate)
      }
      animate()
    })
  }

  function stopHighlightAnimation(): void {
    if (highlightAnimationActive) {
      stopHighlightRequested = true
    }
  }

  return {
    highlightWinsAnimation,
    stopHighlightAnimation,
    isAnimating
  }
}
