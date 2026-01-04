import { Container } from 'pixi.js'
import { useAudioEffects } from '@/composables/useAudioEffects'
import { useTimingStore } from '@/stores'
import { getBufferOffset } from '@/utils/gameHelpers'
import { CONFIG } from '@/config/constants'
import { createTimerManager, type BaseOverlay } from './base'

/**
 * Grid state interface
 */
interface GridState {
  grid: string[][]
}

/**
 * Reels interface with triggerPop method
 */
interface Reels {
  triggerPop?: (col: number, row: number) => void
}

/**
 * Bonus tile pop animation interface
 */
export interface BonusTilePopAnimation extends BaseOverlay {
  show: (canvasWidth: number, canvasHeight: number, mainRect: any, tileSize: any, onComplete: () => void) => void
}

/**
 * Creates a bonus tile pop animation overlay
 * Pops each bonus tile one-by-one with sound effect before the jackpot video
 */
export function createBonusTilePopAnimation(gridState: GridState, reels: Reels): BonusTilePopAnimation {
  const audioEffects = useAudioEffects()
  const timingStore = useTimingStore()
  const timers = createTimerManager()

  const container = new Container()
  container.visible = false
  container.zIndex = 1100

  const BUFFER_OFFSET = getBufferOffset()
  const WIN_CHECK_START_ROW = CONFIG.reels.winCheckStartRow
  const WIN_CHECK_END_ROW = CONFIG.reels.winCheckEndRow

  let isPlaying = false
  let onCompleteCallback: (() => void) | null = null
  let currentTileIndex = 0
  let bonusTilePositions: Array<{ col: number; row: number }> = []

  /**
   * Pop tiles one by one recursively
   */
  function popNextTile(sessionId: number): void {
    if (sessionId !== timers.getSessionId()) return

    if (currentTileIndex >= bonusTilePositions.length) {
      // All tiles popped, complete after short delay
      timers.setTimeout(() => {
        hide()
      }, 500, sessionId)
      return
    }

    const tilePos = bonusTilePositions[currentTileIndex]
    const visualRow = tilePos.row - BUFFER_OFFSET

    // Play pop sound
    audioEffects.playEffect('lot')

    // Trigger pop animation on the actual tile sprite
    if (reels?.triggerPop) {
      reels.triggerPop(tilePos.col, visualRow)
    }

    // Move to next tile after delay
    currentTileIndex++
    timers.setTimeout(() => {
      popNextTile(sessionId)
    }, 300, sessionId)
  }

  /**
   * Start the pop animation for all bonus tiles
   */
  function show(
    _canvasWidth: number,
    _canvasHeight: number,
    _mainRect: any,
    _tileSize: any,
    onComplete: () => void
  ): void {
    const sessionId = timers.newSession()

    container.visible = true
    isPlaying = true
    onCompleteCallback = onComplete

    // Find all bonus tile positions in the winning check rows
    bonusTilePositions = []
    for (let col = 0; col < CONFIG.reels.count; col++) {
      for (let row = WIN_CHECK_START_ROW; row <= WIN_CHECK_END_ROW; row++) {
        const cell = gridState.grid[col][row]
        if (cell === 'bonus') {
          bonusTilePositions.push({ col, row })
        }
      }
    }

    if (bonusTilePositions.length === 0) {
      // No tiles to pop, complete immediately
      hide()
      return
    }

    // Reset state
    currentTileIndex = 0

    // Delay before starting pop animation
    const pauseBeforePop = timingStore.JACKPOT_PAUSE_BEFORE_POP

    timers.setTimeout(() => {
      popNextTile(sessionId)
    }, pauseBeforePop, sessionId)
  }

  /**
   * Hide the animation
   */
  function hide(): void {
    container.visible = false
    isPlaying = false
    timers.clearAll()

    // Trigger completion callback
    if (onCompleteCallback) {
      onCompleteCallback()
      onCompleteCallback = null
    }
  }

  /**
   * Update (called each frame) - no-op for this overlay
   */
  function update(_timestamp: number): void {
    // Animation is handled by setTimeout sequences
  }

  /**
   * Build/rebuild for canvas resize - no-op for this overlay
   */
  function build(_canvasWidth: number, _canvasHeight: number): void {
    // Nothing to rebuild, positions are calculated on show
  }

  return {
    container,
    show,
    hide,
    update,
    build,
    isShowing: () => isPlaying || container.visible
  }
}
