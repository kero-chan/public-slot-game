// @ts-nocheck
import { Container, Graphics, Sprite, type Application } from 'pixi.js'
import type { UseGameState } from '@/composables/slotMachine/useGameState'
import type { HeaderRect, MultiplierItem } from './modules/types'
import { buildBackgrounds, updateMultiplierBackground, type FrameTheme } from './modules/backgroundBuilder'
import {
  buildMultipliers,
  updateMultiplierSprites,
  transitionMultipliers,
  type MultiplierDisplayState
} from './modules/multiplierDisplay'

export type { HeaderRect, FrameTheme }

export interface UseHeader {
  container: Container
  build: (rect: HeaderRect) => void
  updateValues: () => void
  initializeComposedTextures: (pixiApp: Application) => void
  /** Set the frame theme (for high-value symbol wins) */
  setFrameTheme: (theme: FrameTheme) => void
}

export function useHeader(gameState: UseGameState): UseHeader {
  const container = new Container()

  let multiplierState: MultiplierDisplayState = { sprites: [] }
  let app: Application | null = null
  let isTransitioning = false
  let lastAutoSpinState = false
  let headerRect: HeaderRect = { x: 0, y: 0, w: 0, h: 0 }
  let bg02Y = 0
  let bg02Height = 0
  let multiplierBgSprite: Sprite | null = null
  let currentFrameTheme: FrameTheme = 'default'

  function initializeComposedTextures(pixiApp: Application): void {
    app = pixiApp
    // Force rebuild header after assets are loaded to ensure textures have proper dimensions
    if (headerRect.w > 0 && headerRect.h > 0) {
      build(headerRect)
    }
  }

  function getDisplayedMultiplier(): number {
    return gameState.currentMultiplier?.value || 1
  }

  function build(rect: HeaderRect): void {
    container.removeChildren()
    multiplierState = { sprites: [] }
    headerRect = { ...rect }
    lastAutoSpinState = gameState.inFreeSpinMode?.value || false

    // Add mask to clip header content to game area
    const headerMask = new Graphics()
    headerMask.rect(rect.x, rect.y, rect.w, rect.h)
    headerMask.fill(0xffffff)
    container.mask = headerMask
    container.addChild(headerMask)

    // Build backgrounds
    const bgResult = buildBackgrounds(container, rect, currentFrameTheme)
    bg02Y = bgResult.bg02Y
    bg02Height = bgResult.bg02Height
    multiplierBgSprite = bgResult.multiplierBgSprite

    // Build multipliers
    const config = {
      inFreeSpinMode: gameState.inFreeSpinMode?.value || false,
      currentMultiplier: getDisplayedMultiplier()
    }
    multiplierState = buildMultipliers(container, rect, bg02Y, bg02Height, config)
  }

  function updateValues(): void {
    const currentAutoSpinState = gameState.inFreeSpinMode?.value || false

    // Check if mode changed - trigger transition
    if (currentAutoSpinState !== lastAutoSpinState) {
      lastAutoSpinState = currentAutoSpinState

      if (!isTransitioning) {
        isTransitioning = true
        const config = {
          inFreeSpinMode: currentAutoSpinState,
          currentMultiplier: getDisplayedMultiplier()
        }
        transitionMultipliers(
          container,
          multiplierState,
          headerRect,
          config,
          bg02Y,
          bg02Height,
          (newState) => {
            multiplierState = newState
            isTransitioning = false
          }
        )
      }
      return
    }

    // Normal update - just update textures and backgrounds
    const config = {
      inFreeSpinMode: currentAutoSpinState,
      currentMultiplier: getDisplayedMultiplier()
    }
    multiplierState = updateMultiplierSprites(
      container,
      multiplierState,
      config,
      headerRect.h
    )
  }

  /**
   * Set the frame theme (for high-value symbol wins)
   */
  function setFrameTheme(theme: FrameTheme): void {
    if (theme === currentFrameTheme) return
    currentFrameTheme = theme
    updateMultiplierBackground(multiplierBgSprite, theme)
  }

  return { container, build, updateValues, initializeComposedTextures, setFrameTheme }
}
