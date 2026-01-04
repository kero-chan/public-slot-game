import type { GridState } from '@/types/global'
import type { AnimObject, SpinConfig, SpinAnimationState } from './types'
import { syncColumnToGrid, injectBackendGridForColumn } from './gridSync'
import { countTotalBonusTiles, countBonusAtStripPosition } from './bonusCounter'

// Visual thresholds for anticipation animation (purely cosmetic)
const ANTICIPATION_THRESHOLD = 2  // Show anticipation when 2 bonus tiles visible
const JACKPOT_VISUAL_THRESHOLD = 3  // Stop animation early when 3 bonus tiles visible

export interface AnticipationContext {
  gridState: GridState
  backendTargetGrid: string[][] | null
  state: SpinAnimationState
  config: SpinConfig
  gameStore: any
  gsap: any
  playEffect: (name: string) => void
  anticipationSlowdownPerColumn: number
  completeSpin: () => void
}

export function findValidSlowdownTarget(
  strip: string[],
  currentPosition: number,
  config: SpinConfig
): number {
  const { stripLength, maxBonusPerColumn, winCheckStartRow, winCheckEndRow } = config
  const minDistanceForSlowdown = 50
  const maxSearchDistance = currentPosition >= stripLength
    ? 200
    : Math.max(stripLength - currentPosition - 5, 200)

  for (let distance = minDistanceForSlowdown; distance <= maxSearchDistance; distance += 5) {
    const candidateTarget = Math.floor(currentPosition + distance)

    if (currentPosition < stripLength && candidateTarget >= stripLength - 5) {
      break
    }

    const bonusCount = countBonusAtStripPosition(strip, candidateTarget, winCheckStartRow, winCheckEndRow)
    if (bonusCount <= maxBonusPerColumn) {
      return candidateTarget
    }
  }

  let fallback = Math.floor(currentPosition + minDistanceForSlowdown)
  if (currentPosition < stripLength && fallback >= stripLength - 5) {
    fallback = stripLength - 5
  }
  return fallback
}

export function createSlowdownTween(ctx: AnticipationContext, targetCol: number): void {
  const { gridState, backendTargetGrid, state, config, gameStore, gsap, playEffect, anticipationSlowdownPerColumn, completeSpin } = ctx
  const { cols, winCheckStartRow, winCheckEndRow } = config

  if (targetCol >= cols) return

  const targetTween = state.columnTweens.get(targetCol)
  const targetAnimObj = state.animObjects[targetCol]

  if (!targetTween || !targetTween.isActive() || !targetAnimObj) return

  targetTween.kill()

  const strip = gridState.reelStrips[targetCol]
  const newTarget = findValidSlowdownTarget(strip, targetAnimObj.position, config)

  targetAnimObj.targetIndex = newTarget
  injectBackendGridForColumn(gridState, backendTargetGrid, targetCol, newTarget, config)

  const newTween = gsap.to(targetAnimObj, {
    position: newTarget,
    duration: anticipationSlowdownPerColumn,
    ease: 'power4.out',
    onUpdate: function() {
      updateColumnDuringAnimation(gridState, targetAnimObj, targetCol)
    },
    onComplete: () => {
      onSlowdownComplete(ctx, targetCol, targetAnimObj)
    }
  })

  state.columnTweens.set(targetCol, newTween)
  state.currentSlowdownColumn = targetCol
  gridState.activeSlowdownColumn = targetCol
  playEffect("reach_bonus")
}

function updateColumnDuringAnimation(gridState: GridState, animObj: AnimObject, col: number): void {
  const currentPosition = animObj.position
  const newTopIndex = Math.floor(currentPosition)
  const newOffset = currentPosition - newTopIndex

  const distanceRemaining = animObj.targetIndex - currentPosition
  const totalDistance = animObj.targetIndex
  const progress = 1 - (distanceRemaining / totalDistance)

  // Linear velocity decay for consistent slowdown without jitter
  const baseVelocity = 12
  const linearDecay = Math.max(0, 1 - progress)
  const calculatedVelocity = Math.max(1, baseVelocity * linearDecay)

  gridState.reelTopIndex[col] = newTopIndex % gridState.reelStrips[col].length
  gridState.spinOffsets[col] = newOffset
  gridState.spinVelocities[col] = calculatedVelocity
}

function onSlowdownComplete(ctx: AnticipationContext, targetCol: number, targetAnimObj: AnimObject): void {
  const { gridState, backendTargetGrid, state, config, gameStore, gsap, completeSpin } = ctx
  const { cols, totalRows, winCheckStartRow, winCheckEndRow } = config

  gridState.reelTopIndex[targetCol] = targetAnimObj.targetIndex % gridState.reelStrips[targetCol].length
  syncColumnToGrid(gridState, targetCol, totalRows)

  // Simple snap settle - immediately set velocity to 0 and animate offset
  gridState.spinVelocities[targetCol] = 0

  const settleObj = { offset: gridState.spinOffsets[targetCol] }
  gsap.to(settleObj, {
    offset: 0,
    duration: 0.15,
    ease: 'power1.out',
    onUpdate: () => {
      gridState.spinOffsets[targetCol] = settleObj.offset
    },
    onComplete: () => {
      gridState.spinOffsets[targetCol] = 0
    }
  })

  state.stoppedColumns.add(targetCol)

  // Anticipation column stop verification (debug logs removed for performance)

  if (state.currentSlowdownColumn === targetCol) {
    state.currentSlowdownColumn = -1
    gridState.activeSlowdownColumn = -1
  }

  if (!gameStore.inFreeSpinMode && gameStore.anticipationMode) {
    const totalBonusTiles = countTotalBonusTiles(gridState, state.stoppedColumns, winCheckStartRow, winCheckEndRow)

    if (totalBonusTiles >= JACKPOT_VISUAL_THRESHOLD) {
      forceStopRemainingColumns(ctx, targetCol)
      completeSpin()
      return
    }
  }

  if (gameStore.anticipationMode && targetCol >= state.firstBonusColumn) {
    const nextCol = targetCol + 1
    if (nextCol < cols) {
      createSlowdownTween(ctx, nextCol)
    } else {
      completeSpin()
    }
  }
}

function forceStopRemainingColumns(ctx: AnticipationContext, fromCol: number): void {
  const { gridState, backendTargetGrid, state, config, gsap } = ctx
  const { cols, totalRows } = config

  // Force stopping remaining columns (debug log removed for performance)

  for (let col = fromCol + 1; col < cols; col++) {
    const remainingTween = state.columnTweens.get(col)
    if (remainingTween && remainingTween.isActive()) {
      remainingTween.kill()

      const remainingAnimObj = state.animObjects[col]
      const currentPos = Math.floor(remainingAnimObj.position)

      injectBackendGridForColumn(gridState, backendTargetGrid, col, currentPos, config)

      gridState.reelTopIndex[col] = currentPos % gridState.reelStrips[col].length
      syncColumnToGrid(gridState, col, totalRows)

      // Simple snap settle
      gridState.spinVelocities[col] = 0

      const settleObj = { offset: gridState.spinOffsets[col] }
      gsap.to(settleObj, {
        offset: 0,
        duration: 0.15,
        ease: 'power1.out',
        onUpdate: () => {
          gridState.spinOffsets[col] = settleObj.offset
        },
        onComplete: () => {
          gridState.spinOffsets[col] = 0
        }
      })

      state.stoppedColumns.add(col)
    }
  }
}

export function checkAndActivateAnticipation(
  ctx: AnticipationContext,
  col: number
): void {
  const { gridState, state, config, gameStore, gsap, completeSpin } = ctx
  const { cols, winCheckStartRow, winCheckEndRow } = config

  if (gameStore.inFreeSpinMode) return

  const totalBonusTiles = countTotalBonusTiles(gridState, state.stoppedColumns, winCheckStartRow, winCheckEndRow)
  const anticipationThreshold = ANTICIPATION_THRESHOLD

  if (!gameStore.anticipationMode && state.firstBonusColumn === -1 && totalBonusTiles === anticipationThreshold) {
    state.firstBonusColumn = col
    gameStore.activateAnticipationMode()

    let firstSpinningColumn = -1
    for (let i = col + 1; i < cols; i++) {
      const tween = state.columnTweens.get(i)
      if (tween && tween.isActive() && !state.stoppedColumns.has(i)) {
        firstSpinningColumn = i
        break
      }
    }

    if (firstSpinningColumn === -1) {
      gameStore.deactivateAnticipationMode()
      completeSpin()
    } else {
      convertToPerpetualSpin(ctx, firstSpinningColumn)
      createSlowdownTween(ctx, firstSpinningColumn)
    }
  }
}

function convertToPerpetualSpin(ctx: AnticipationContext, fromCol: number): void {
  const { gridState, state, config, gsap } = ctx
  const { cols } = config

  for (let i = fromCol + 1; i < cols; i++) {
    const remainingTween = state.columnTweens.get(i)
    const remainingAnimObj = state.animObjects[i]

    if (remainingTween && remainingTween.isActive() && remainingAnimObj) {
      remainingTween.kill()

      const currentPos = remainingAnimObj.position
      const perpetualDistance = 5000
      const perpetualDuration = perpetualDistance / 35

      const perpetualTween = gsap.to(remainingAnimObj, {
        position: currentPos + perpetualDistance,
        duration: perpetualDuration,
        ease: 'none',
        onUpdate: function() {
          const currentPosition = remainingAnimObj.position
          const newTopIndex = Math.floor(currentPosition)
          const newOffset = currentPosition - newTopIndex

          gridState.reelTopIndex[i] = newTopIndex % gridState.reelStrips[i].length
          gridState.spinOffsets[i] = newOffset
          gridState.spinVelocities[i] = 18
        }
      })

      state.columnTweens.set(i, perpetualTween)
    }
  }
}
