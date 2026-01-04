/**
 * Win Animation Manager
 * Routes win animations to the correct handler based on intensity level
 */

import { Container } from 'pixi.js'
import type { WinIntensity } from '@/types/global'
import { createSmallWinAnimation, type SmallWinAnimation } from './smallWinAnimation'
import { createMediumWinAnimation, type MediumWinAnimation } from './mediumWinAnimation'
import { createBigWinAnimation, type BigWinAnimation } from './bigWinAnimation'
import { createMegaWinAnimation, type MegaWinAnimation, type TilePosition } from './megaWinAnimation'
import type { BaseOverlay } from './base'

export type { TilePosition }

/**
 * Win animation manager interface
 */
export interface WinAnimationManager extends BaseOverlay {
  show: (
    canvasWidth: number,
    canvasHeight: number,
    tilePositions: TilePosition[],
    intensity: WinIntensity,
    amount: number,
    onComplete?: () => void
  ) => void
  getContainer: () => Container
}

/**
 * Creates a win animation manager that routes to appropriate animation based on intensity
 */
export function createWinAnimationManager(reelsRef?: any): WinAnimationManager {
  // Create all animation instances
  const smallWin: SmallWinAnimation = createSmallWinAnimation()
  const mediumWin: MediumWinAnimation = createMediumWinAnimation()
  const bigWin: BigWinAnimation = createBigWinAnimation(reelsRef)
  const megaWin: MegaWinAnimation = createMegaWinAnimation(reelsRef)

  // Main container holds all animation containers
  const container = new Container()
  container.zIndex = 2000

  container.addChild(smallWin.container)
  container.addChild(mediumWin.container)
  container.addChild(bigWin.container)
  container.addChild(megaWin.container)

  // Track which animation is currently active
  let activeAnimation: BaseOverlay | null = null

  /**
   * Get the appropriate animation handler for intensity
   */
  function getAnimationForIntensity(intensity: WinIntensity): {
    animation: SmallWinAnimation | MediumWinAnimation | BigWinAnimation | MegaWinAnimation
    needsTiles: boolean
  } {
    switch (intensity) {
      case 'small':
        return { animation: smallWin, needsTiles: false }
      case 'medium':
        return { animation: mediumWin, needsTiles: false }
      case 'big':
        return { animation: bigWin, needsTiles: true }
      case 'mega':
        return { animation: megaWin, needsTiles: true }
      default:
        return { animation: smallWin, needsTiles: false }
    }
  }

  /**
   * Show win animation based on intensity
   */
  function show(
    canvasWidth: number,
    canvasHeight: number,
    tilePositions: TilePosition[],
    intensity: WinIntensity,
    amount: number,
    onComplete?: () => void
  ): void {
    // Hide any currently active animation
    if (activeAnimation) {
      activeAnimation.hide()
    }

    const { animation } = getAnimationForIntensity(intensity)
    activeAnimation = animation

    // Call the appropriate show method
    animation.show(canvasWidth, canvasHeight, tilePositions, amount, () => {
      activeAnimation = null
      // Wrap callback in try/catch to prevent errors from breaking the animation flow
      if (onComplete) {
        try {
          onComplete()
        } catch (error) {
          console.error('Error in win animation onComplete callback:', error)
        }
      }
    })
  }

  /**
   * Hide current animation
   */
  function hide(): void {
    if (activeAnimation) {
      activeAnimation.hide()
      activeAnimation = null
    }

    // Hide all just in case
    smallWin.hide()
    mediumWin.hide()
    bigWin.hide()
    megaWin.hide()
  }

  /**
   * Update all animations
   */
  function update(deltaTime = 1): void {
    smallWin.update?.(deltaTime)
    mediumWin.update?.(deltaTime)
    bigWin.update?.(deltaTime)
    megaWin.update?.(deltaTime)
  }

  /**
   * Build/resize all animations
   */
  function build(width: number, height: number): void {
    smallWin.build?.(width, height)
    mediumWin.build?.(width, height)
    bigWin.build?.(width, height)
    megaWin.build?.(width, height)
  }

  /**
   * Check if any animation is showing
   */
  function isShowing(): boolean {
    return (
      smallWin.isShowing() ||
      mediumWin.isShowing() ||
      bigWin.isShowing() ||
      megaWin.isShowing()
    )
  }

  /**
   * Get the main container
   */
  function getContainer(): Container {
    return container
  }

  return {
    container,
    show,
    hide,
    update,
    build,
    isShowing,
    getContainer
  }
}
