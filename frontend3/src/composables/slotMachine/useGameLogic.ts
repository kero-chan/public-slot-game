// @ts-nocheck
import { CONFIG } from '@/config/constants'
import { getBufferOffset, getRandomSymbol } from '@/utils/gameHelpers'
import { useAudioEffects } from '@/composables/useAudioEffects'
import { useGameStore, useTimingStore, useBackendGameStore, useSettingsStore, useFreeSpinsStore } from '@/stores'
import { getCurrentSpinBaseDuration, getCurrentSpinReelStagger, getCurrentAnticipationSlowdown } from '@/utils/spinTiming'
import { gsap } from 'gsap'
import type { UseGameState } from '@/composables/slotMachine/useGameState'
import type { GridState, WinCombination, SpinResponse, WinIntensity } from '@/types/global'
import type { UseReels } from '@/composables/slotMachine/reels'
import type { TilePosition } from '@/composables/slotMachine/reels/winning/imageLightningLinks'
import type { HighValueSymbol } from '@/composables/slotMachine/effects/symbolWinAnimation'

import { useSpinState, verifyGridMatch, logVerification, type UseSpinState } from '@/features/spin'
import { getWinIntensity } from '@/features/wins'
import { getBackendWins as getBackendWinsService, findWinningCombinations as findWins } from '@/features/wins'

import {
  useBetControls,
  useHighlightAnimation,
  useCascadeAnimation,
  useGoldTransform
} from './gameLogic'

import {
  buildReelStrips,
  injectBackendGrid,
  createSpinTimeline,
  type SpinConfig,
  type SpinAnimationState,
  type AnimObject
} from './gameLogic/spinAnimation'

export type RenderFunction = () => void
export type ShowWinOverlayFunction = (intensity: WinIntensity, amount: number) => void
export type ShowSymbolWinAnimationFunction = (symbol: HighValueSymbol) => Promise<void>
export type HideSymbolWinAnimationFunction = () => void

export interface UseGameLogic {
  spin: (gameMode?: string) => Promise<void>
  increaseBet: () => void
  decreaseBet: () => void
  animateSpin: (response: SpinResponse) => Promise<void>
  findWinningCombinations: () => WinCombination[]
  highlightWinsAnimation: (wins: WinCombination[]) => Promise<void>
  stopHighlightAnimation: () => void
  computeGoldTransformedPositions: (wins: WinCombination[]) => void
  transformGoldTilesToWild: (wins: WinCombination[]) => Promise<void>
  animateDisappear: (wins: WinCombination[]) => Promise<void>
  cascadeSymbols: (wins: WinCombination[]) => Promise<void>
  getWinIntensity: (wins: WinCombination[]) => WinIntensity
  getBackendWins: () => WinCombination[] | null
  getCurrentWinningTileKind: () => string | undefined
  hasMoreBackendCascades: () => boolean
  playConsecutiveWinSound: (consecutiveWins: number, isFreeSpinMode: boolean) => void
  playWinSound: (wins: WinCombination[]) => void
  playLineWinSound: () => void
  playEffect: (effectName: string) => void
  showWinOverlay?: ShowWinOverlayFunction
  showFinalJackpotResult?: (amount: number) => void
  showSymbolWinAnimation?: ShowSymbolWinAnimationFunction
  hideSymbolWinAnimation?: HideSymbolWinAnimationFunction
  setReels: (reelsRef: UseReels) => void
  clearCompletedShatterAnimations: () => void
  startLightningLinksAnimation: (winsBySymbol: Map<string, TilePosition[]>) => Promise<void>
}

export function useGameLogic(
  gameState: UseGameState,
  gridState: GridState,
  render: RenderFunction,
  showWinOverlayFn?: ShowWinOverlayFunction,
  showFinalJackpotResultFn?: (amount: number) => void,
  showSymbolWinAnimationFn?: ShowSymbolWinAnimationFunction,
  hideSymbolWinAnimationFn?: HideSymbolWinAnimationFunction,
  reelsAPI: UseReels | null = null
): UseGameLogic {
  const gameStore = useGameStore()
  const timingStore = useTimingStore()
  const backendGame = useBackendGameStore()

  let reels: UseReels | null = reelsAPI
  const spinState: UseSpinState = useSpinState()
  const { playConsecutiveWinSound, playWinSound, playLineWinSound, playEffect } = useAudioEffects()

  const betControls = useBetControls()
  const highlightAnimation = useHighlightAnimation({ gridState })
  const goldTransform = useGoldTransform({ gridState, getReels: () => reels })
  const cascadeAnimation = useCascadeAnimation({
    gridState,
    render,
    getCurrentCascade: () => spinState.currentCascade.value,
    currentCascadeIndex: spinState.currentCascadeIndex.value
  })

  const BUFFER_OFFSET = getBufferOffset()
  const WIN_CHECK_START_ROW = CONFIG.reels.winCheckStartRow
  const WIN_CHECK_END_ROW = CONFIG.reels.winCheckEndRow

  const getBackendWins = (): WinCombination[] | null => getBackendWinsService(spinState)
  const findWinningCombinations = (): WinCombination[] => findWins(spinState)
  const getCurrentWinningTileKind = (): string | undefined => {
    const cascade = spinState.currentCascade.value
    return cascade?.winning_tile_kind
  }

  function createSpinConfig(): SpinConfig {
    return {
      cols: CONFIG.reels.count,
      totalRows: CONFIG.reels.rows + BUFFER_OFFSET,
      stripLength: CONFIG.reels.stripLength,
      maxBonusPerColumn: CONFIG.game.maxBonusPerColumn || 2,
      winCheckStartRow: WIN_CHECK_START_ROW,
      winCheckEndRow: WIN_CHECK_END_ROW,
      bufferOffset: BUFFER_OFFSET
    }
  }

  function createAnimationState(targetIndexes: number[], cols: number): SpinAnimationState {
    const animObjects: AnimObject[] = []
    for (let col = 0; col < cols; col++) {
      animObjects.push({
        col,
        position: 0,
        targetIndex: targetIndexes[col],
        isSlowingDown: false,
        hasSlowedDown: false
      })
    }

    return {
      stoppedColumns: new Set<number>(),
      columnTweens: new Map<number, any>(),
      firstBonusColumn: -1,
      currentSlowdownColumn: -1,
      backendInjected: false,
      animObjects,
      targetIndexes
    }
  }

  const animateSpin = (response: SpinResponse): Promise<void> => {
    const config = createSpinConfig()
    const { cols, totalRows, stripLength } = config

    const settingsStore = useSettingsStore()
    const freeSpinsStore = useFreeSpinsStore()
    
    const baseDuration = getCurrentSpinBaseDuration(settingsStore.fastSpin, freeSpinsStore.inFreeSpinMode) / 1000
    const stagger = getCurrentSpinReelStagger(settingsStore.fastSpin, freeSpinsStore.inFreeSpinMode) / 1000
    const anticipationSlowdownPerColumn = getCurrentAnticipationSlowdown(settingsStore.fastSpin, freeSpinsStore.inFreeSpinMode) / 1000

    const backendTargetGrid = spinState.backendTargetGrid.value

    buildReelStrips(gridState, config, getRandomSymbol)

    gridState.reelTopIndex = Array(cols).fill(0)
    gridState.spinOffsets = Array(cols).fill(0)
    gridState.spinVelocities = Array(cols).fill(18)
    gridState.activeSlowdownColumn = -1

    const minLanding = totalRows + 10
    const maxLanding = stripLength - totalRows
    const targetIndexes: number[] = Array(cols).fill(0).map(() =>
      Math.floor(minLanding + Math.random() * (maxLanding - minLanding))
    )

    const state = createAnimationState(targetIndexes, cols)
    state.backendInjected = injectBackendGrid(gridState, backendTargetGrid, targetIndexes, config, getRandomSymbol)

    return createSpinTimeline({
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
    })
  }

  const spin = async (gameMode?: string): Promise<void> => {
    const started = gameStore.startSpinCycle()
    if (!started) return

    try {
      const expectedRows = CONFIG.reels.rows
      for (let col = 0; col < 5; col++) {
        if (!gridState.grid[col] || gridState.grid[col].length === 0) {
          gridState.grid[col] = Array(expectedRows).fill(null).map(() => 'fa')
        } else {
          for (let row = 0; row < expectedRows; row++) {
            if (!gridState.grid[col][row] || typeof gridState.grid[col][row] !== 'string') {
              gridState.grid[col][row] = 'fa'
            }
          }
        }
      }

      playEffect("lot")
      playEffect("reel_spin")

      let response: SpinResponse | null
      if (gameStore.inFreeSpinMode) {
        response = await backendGame.executeFreeSpin()
      } else {
        response = await backendGame.executeSpin(gameStore.bet, gameMode)
      }

      if (!response) {
        console.error('Failed to get spin result from backend')
        gameStore.transitionTo('idle')
        return
      }

      gameStore.setSpinResponse(response)
      spinState.setBackendSpinResult(response)
      gameState.backendSpinResponse = response as any

      await animateSpin(response)

      const verification = verifyGridMatch(response.grid, gridState.grid)
      logVerification(verification, 'SPIN COMPLETE')
      gameStore.completeSpinAnimation()

    } catch (error) {
      console.error('Spin failed:', error)
      gameStore.transitionTo('idle')
      await backendGame.refreshBalance()
    }
  }

  function setReels(reelsRef: UseReels): void {
    reels = reelsRef
  }

  const hasMoreBackendCascades = (): boolean => spinState.hasMoreBackendCascades()

  function clearCompletedShatterAnimations(): void {
    if (reels) {
      reels.clearCompletedShatterAnimations()
    }
  }

  async function startLightningLinksAnimation(winsBySymbol: Map<string, TilePosition[]>): Promise<void> {
    if (reels && reels.lightningLinks) {
      await reels.lightningLinks.startSequentialAnimation(winsBySymbol)
    }
  }

  return {
    spin,
    increaseBet: betControls.increaseBet,
    decreaseBet: betControls.decreaseBet,
    animateSpin,
    findWinningCombinations,
    highlightWinsAnimation: highlightAnimation.highlightWinsAnimation,
    stopHighlightAnimation: highlightAnimation.stopHighlightAnimation,
    computeGoldTransformedPositions: cascadeAnimation.computeGoldTransformedPositions,
    transformGoldTilesToWild: goldTransform.transformGoldTilesToWild,
    startGoldTransformVisuals: goldTransform.startGoldTransformVisuals,
    animateDisappear: cascadeAnimation.animateDisappear,
    cascadeSymbols: cascadeAnimation.cascadeSymbols,
    getWinIntensity,
    getBackendWins,
    getCurrentWinningTileKind,
    hasMoreBackendCascades,
    playConsecutiveWinSound,
    playWinSound,
    playLineWinSound,
    playEffect,
    showWinOverlay: showWinOverlayFn,
    showFinalJackpotResult: showFinalJackpotResultFn,
    showSymbolWinAnimation: showSymbolWinAnimationFn,
    hideSymbolWinAnimation: hideSymbolWinAnimationFn,
    setReels,
    clearCompletedShatterAnimations,
    startLightningLinksAnimation
  }
}
