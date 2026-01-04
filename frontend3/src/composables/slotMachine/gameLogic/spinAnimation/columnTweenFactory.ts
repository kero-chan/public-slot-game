import type { GridState } from '@/types/global'
import type { AnimObject, SpinConfig, SpinAnimationState } from './types'
import { syncColumnToGrid } from './gridSync'
import { checkAndActivateAnticipation, type AnticipationContext } from './anticipationHandler'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'

export interface TweenFactoryConfig {
  baseDuration: number
  stagger: number
}

export function createColumnTween(
  col: number,
  animObj: AnimObject,
  gridState: GridState,
  config: SpinConfig,
  tweenConfig: TweenFactoryConfig,
  state: SpinAnimationState,
  gsap: any,
  anticipationCtx: AnticipationContext
): any {
  const { totalRows } = config
  const { baseDuration, stagger } = tweenConfig
  const sequentialDuration = baseDuration + (col * stagger * 2)

  return gsap.to(animObj, {
    position: animObj.targetIndex,
    duration: sequentialDuration,
    ease: 'power2.out',
    delay: col * stagger,
    onUpdate: function() {
      updateColumnPosition(gridState, animObj, col, totalRows)
    },
    onComplete: () => {
      onColumnComplete(col, animObj, gridState, config, state, gsap, anticipationCtx)
    }
  })
}

function updateColumnPosition(
  gridState: GridState,
  animObj: AnimObject,
  col: number,
  totalRows: number
): void {
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

  syncColumnToGrid(gridState, col, totalRows)
}

function onColumnComplete(
  col: number,
  animObj: AnimObject,
  gridState: GridState,
  config: SpinConfig,
  state: SpinAnimationState,
  gsap: any,
  anticipationCtx: AnticipationContext
): void {
  const { totalRows, cols } = config

  gridState.reelTopIndex[col] = animObj.targetIndex % gridState.reelStrips[col].length
  syncColumnToGrid(gridState, col, totalRows)

  // Simple snap settle - immediately set velocity to 0 and animate offset
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

  // Play reel_spin_stop sound when each column locks in
  audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'reel_spin_stop', volume: 0.9 })

  // When the last column (index 4) stops, fade out the reel_spin sound
  if (col === cols - 1) {
    // Fade reel_spin from current volume to 0 over 150ms
    audioEvents.emit(AUDIO_EVENTS.EFFECT_FADE, { audioKey: 'reel_spin', from: 0.9, to: 0, duration: 150 })
  }

  checkAndActivateAnticipation(anticipationCtx, col)

  if (state.currentSlowdownColumn === col) {
    state.currentSlowdownColumn = -1
    gridState.activeSlowdownColumn = -1
  }
}
