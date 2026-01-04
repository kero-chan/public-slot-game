import type { GameLogicAPI, FlowTimerManager } from './types'

/**
 * Handle gold tile transformation
 * First computes which positions should transform (from backend data),
 * then runs the transformation animation
 */
export async function handleGoldTransformation(
  gameLogic: GameLogicAPI,
  gameStore: any,
  render: () => void
): Promise<void> {
  const wins = gameStore.currentWins

  if (!wins || wins.length === 0) {
    gameStore.completeGoldTransformation()
    return
  }

  // Compute gold transformed positions from backend cascade data BEFORE running animation
  // This populates gridState.goldTransformedPositions which is used by both:
  // 1. transformGoldTilesToWild - to know which tiles to animate
  // 2. cascadeSymbols - to exclude these positions from cascade/drop
  gameLogic.computeGoldTransformedPositions(wins)

  await gameLogic.transformGoldTilesToWild(wins)
  render()

  gameStore.completeGoldTransformation()
}

/**
 * Handle wait after gold transformation
 */
export function handleGoldWait(
  timerManager: FlowTimerManager,
  gameStore: any,
  goldWaitDuration: number
): void {
  timerManager.set(() => {
    gameStore.completeGoldWait()
  }, goldWaitDuration)
}

/**
 * Handle tile disappearing animation
 */
export async function handleDisappearingTiles(
  gameLogic: GameLogicAPI,
  gameStore: any,
  winningStore: any
): Promise<void> {
  const wins = gameStore.currentWins

  if (!wins || wins.length === 0) {
    gameStore.completeDisappearing()
    return
  }

  // Transition to DISAPPEARING state
  winningStore.setDisappearing()

  // Run disappear animation
  await gameLogic.animateDisappear(wins)

  gameStore.completeDisappearing()
}

/**
 * Handle cascade animation
 */
export async function handleCascade(
  gameLogic: GameLogicAPI,
  gameStore: any,
  winningStore: any
): Promise<void> {
  const wins = gameStore.currentWins

  if (!wins || wins.length === 0) {
    gameStore.completeCascade()
    return
  }

  // Clear winning state since positions are changing
  winningStore.clearWinningState()

  // Run cascade
  await gameLogic.cascadeSymbols(wins)

  gameStore.completeCascade()
}

/**
 * Handle wait after cascade
 */
export function handleCascadeWait(
  timerManager: FlowTimerManager,
  gameStore: any,
  cascadeWaitDuration: number
): void {
  timerManager.set(() => {
    gameStore.completeCascadeWait()
  }, cascadeWaitDuration)
}

/**
 * Handle no wins state - all cascades are done
 */
export function handleNoWins(gameStore: any): void {
  gameStore.completeNoWins()
}
