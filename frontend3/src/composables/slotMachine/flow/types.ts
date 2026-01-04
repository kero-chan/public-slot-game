import type { WinCombination } from '@/types/global'
import type { TilePosition } from '../reels/winning/imageLightningLinks'
import type { HighValueSymbol } from '../effects/symbolWinAnimation'

// Re-export TilePosition for use in winHandler
export type { TilePosition }

/**
 * Game logic API interface (subset used by flow controller)
 */
export interface GameLogicAPI {
  getBackendWins: () => WinCombination[] | null
  findWinningCombinations: () => WinCombination[]
  getCurrentWinningTileKind: () => string | undefined
  highlightWinsAnimation: (wins: WinCombination[]) => Promise<void>
  stopHighlightAnimation: () => void
  computeGoldTransformedPositions: (wins: WinCombination[]) => void
  transformGoldTilesToWild: (wins: WinCombination[]) => Promise<void>
  startGoldTransformVisuals?: (wins: WinCombination[]) => void
  animateDisappear: (wins: WinCombination[]) => Promise<void>
  cascadeSymbols: (wins: WinCombination[]) => Promise<void>
  getWinIntensity: (wins: WinCombination[]) => 'small' | 'medium' | 'big' | 'mega'
  playConsecutiveWinSound: (consecutiveWins: number, isFreeSpinMode: boolean) => void
  playWinSound: (wins: WinCombination[]) => void
  playLineWinSound: () => void
  showWinOverlay?: (intensity: 'small' | 'medium' | 'big' | 'mega', amount: number) => void
  showFinalJackpotResult?: (amount: number) => void
  clearCompletedShatterAnimations?: () => void
  startLightningLinksAnimation?: (winsBySymbol: Map<string, TilePosition[]>) => Promise<void>
  showSymbolWinAnimation?: (symbol: HighValueSymbol) => Promise<void>
  hideSymbolWinAnimation?: () => void
}

/**
 * Timer manager for flow controller
 */
export interface FlowTimerManager {
  activeTimer: ReturnType<typeof setTimeout> | null
  clear: () => void
  set: (callback: () => void, delay: number) => void
}

/**
 * Create a timer manager
 */
export function createFlowTimerManager(): FlowTimerManager {
  let activeTimer: ReturnType<typeof setTimeout> | null = null

  return {
    get activeTimer() {
      return activeTimer
    },
    clear() {
      if (activeTimer) {
        clearTimeout(activeTimer)
        activeTimer = null
      }
    },
    set(callback: () => void, delay: number) {
      this.clear()
      activeTimer = setTimeout(callback, delay)
    }
  }
}
