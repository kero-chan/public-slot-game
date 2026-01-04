import { watch, type WatchStopHandle } from 'vue'
import { useGameStore, GAME_STATES, useWinningStore, useTimingStore } from '@/stores'
import { getBufferOffset } from '@/utils/gameHelpers'
import type { GridState } from '@/types/global'

// Flow handlers
import {
  type GameLogicAPI,
  createFlowTimerManager,
  handleCheckBonus,
  handleCheckWins,
  handleHighlightWins,
  handleShowWinOverlay,
  handleShowFinalJackpotResult,
  handleGoldTransformation,
  handleGoldWait,
  handleDisappearingTiles,
  handleCascade,
  handleCascadeWait,
  handleNoWins
} from './flow'

// Re-export types for backward compatibility
export type { GameLogicAPI }

/**
 * Render function type
 */
export type RenderFunction = () => void

/**
 * Game Flow Controller composable interface
 */
export interface UseGameFlowController {
  startWatching: () => WatchStopHandle
  clearActiveTimer: () => void
}

/**
 * Game Flow Controller - Orchestrates game state transitions
 * This controller watches the game state and triggers appropriate actions/animations
 */
export function useGameFlowController(
  gameLogic: GameLogicAPI,
  gridState: GridState,
  render: RenderFunction,
  gameState?: any
): UseGameFlowController {
  const gameStore = useGameStore()
  const winningStore = useWinningStore()
  const timingStore = useTimingStore()
  const BUFFER_OFFSET = getBufferOffset()

  // Timer manager for delayed transitions
  const timerManager = createFlowTimerManager()

  /**
   * Main state machine watcher
   */
  const startWatching = (): WatchStopHandle => {
    return watch(
      () => gameStore.gameFlowState,
      async (newState) => {
        switch (newState) {
          case GAME_STATES.SPINNING:
            winningStore.clearWinningState()
            break

          case GAME_STATES.SPIN_COMPLETE:
            if (gameStore.inFreeSpinMode) {
              gameStore.startCheckingWins()
            } else {
              gameStore.startCheckingBonus()
            }
            break

          case GAME_STATES.CHECKING_BONUS:
            await handleCheckBonus(gameLogic, gridState, gameState, winningStore, gameStore)
            break

          case GAME_STATES.POPPING_BONUS_TILES:
          case GAME_STATES.SHOWING_JACKPOT_ANIMATION:
            // Handled by animation components
            break

          case GAME_STATES.SHOWING_BONUS_OVERLAY:
            // Handled by overlay component
            break

          case GAME_STATES.FREE_SPINS_ACTIVE:
            setTimeout(() => {
              if (gameStore.gameFlowState === GAME_STATES.FREE_SPINS_ACTIVE) {
                gameStore.transitionTo(GAME_STATES.IDLE)
              }
            }, 100)
            break

          case GAME_STATES.CHECKING_WINS:
            await handleCheckWins(gameLogic, gameStore)
            break

          case GAME_STATES.HIGHLIGHTING_WINS:
            await handleHighlightWins(gameLogic, gameStore, winningStore, timingStore, BUFFER_OFFSET)
            break

          case GAME_STATES.TRANSFORMING_GOLD:
            await handleGoldTransformation(gameLogic, gameStore, render)
            break

          case GAME_STATES.WAITING_AFTER_GOLD:
            handleGoldWait(timerManager, gameStore, timingStore.GOLD_WAIT)
            break

          case GAME_STATES.DISAPPEARING_TILES:
            await handleDisappearingTiles(gameLogic, gameStore, winningStore)
            break

          case GAME_STATES.CASCADING:
            await handleCascade(gameLogic, gameStore, winningStore)
            break

          case GAME_STATES.WAITING_AFTER_CASCADE:
            handleCascadeWait(timerManager, gameStore, timingStore.CASCADE_WAIT)
            break

          case GAME_STATES.NO_WINS:
            handleNoWins(gameStore)
            break

          case GAME_STATES.SHOWING_WIN_OVERLAY:
            handleShowWinOverlay(gameLogic, gameStore)
            break

          case GAME_STATES.SHOWING_FINAL_JACKPOT_RESULT:
            handleShowFinalJackpotResult(gameLogic, gameStore)
            break

          case GAME_STATES.IDLE:
            timerManager.clear()
            break
        }
      },
      { immediate: false }
    )
  }

  return {
    startWatching,
    clearActiveTimer: timerManager.clear
  }
}
