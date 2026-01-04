import type { GridState } from '@/types/global'
import type { SpinConfig, SpinAnimationState } from './types'
import type { AnticipationContext } from './anticipationHandler'
import { injectBackendGrid } from './gridSync'
import { createColumnTween } from './columnTweenFactory'
import { createSpinCompletionHandler } from './spinCompletion'

export interface SpinTimelineParams {
  gridState: GridState
  backendTargetGrid: string[][] | null
  state: SpinAnimationState
  config: SpinConfig
  targetIndexes: number[]
  baseDuration: number
  stagger: number
  anticipationSlowdownPerColumn: number
  gameStore: any
  gsap: any
  playEffect: (name: string) => void
  getRandomSymbol: (opts: any) => string
}

export function createSpinTimeline(params: SpinTimelineParams): Promise<void> {
  const {
    gridState,
    backendTargetGrid,
    state,
    config,
    targetIndexes,
    baseDuration,
    stagger,
    anticipationSlowdownPerColumn,
    gameStore,
    gsap,
    playEffect,
    getRandomSymbol
  } = params

  const { cols } = config

  return new Promise(resolve => {
    const completion = createSpinCompletionHandler(gridState, gameStore, resolve)

    const anticipationCtx: AnticipationContext = {
      gridState,
      backendTargetGrid,
      state,
      config,
      gameStore,
      gsap,
      playEffect,
      anticipationSlowdownPerColumn,
      completeSpin: completion.complete
    }

    const timeline = gsap.timeline({
      onUpdate: () => {
        if (!state.backendInjected) {
          state.backendInjected = injectBackendGrid(gridState, backendTargetGrid, targetIndexes, config, getRandomSymbol)
        }
      },
      onComplete: () => {
        if (!gameStore.anticipationMode) {
          completion.complete()
        }
      }
    })

    for (let col = 0; col < cols; col++) {
      const tween = createColumnTween(
        col,
        state.animObjects[col],
        gridState,
        config,
        { baseDuration, stagger },
        state,
        gsap,
        anticipationCtx
      )
      state.columnTweens.set(col, tween)
      timeline.add(tween, 0)
    }
  })
}
