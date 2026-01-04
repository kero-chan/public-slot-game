import { Container } from 'pixi.js'
import { gsap } from 'gsap'
import { getTextureForSymbol } from '../textures'
import { useTimingStore } from '@/stores'
import type { Sprite } from 'pixi.js'

/**
 * Gold to wild animation state
 */
interface GoldToWildAnim {
  cellKey: string
  col: number
  row: number
  sprite: Sprite
  tween: gsap.core.Timeline | null
  originalScaleX: number
}

/**
 * Callback to update grid when texture swaps
 */
export type OnTextureSwapCallback = (col: number, row: number) => void

/**
 * Gold to wild animation manager interface
 */
export interface GoldToWildAnimationManager {
  container: Container
  startTransform: (cellKey: string, col: number, row: number, sprite: Sprite, x: number, y: number, onTextureSwap?: OnTextureSwapCallback) => void
  update: (deltaTime?: number) => void
  clear: () => void
  isAnimating: (cellKey: string) => boolean
  hasActiveAnimations: () => boolean
  getAnimationProgress: (cellKey: string) => number
}

/**
 * Creates and manages gold-to-wild transformation animations
 * Simple snap effect: quick scale punch when texture swaps
 */
export function createGoldToWildAnimationManager(): GoldToWildAnimationManager {
  const animations = new Map<string, GoldToWildAnim>()
  const container = new Container()
  const timingStore = useTimingStore()

  /**
   * Start a gold-to-wild transformation animation
   * Simple snap effect: scale up briefly then back to normal when texture swaps
   * @param onTextureSwap - Optional callback called when texture swaps (use to update grid)
   */
  function startTransform(
    cellKey: string,
    col: number,
    row: number,
    sprite: Sprite,
    x: number,
    y: number,
    onTextureSwap?: OnTextureSwapCallback
  ): void {
    if (animations.has(cellKey)) return

    const originalScaleX = sprite.scale.x > 0 ? sprite.scale.x : 1
    const originalScaleY = sprite.scale.y > 0 ? sprite.scale.y : 1

    const anim: GoldToWildAnim = {
      cellKey,
      col,
      row,
      sprite,
      tween: null,
      originalScaleX
    }

    // Use timing store for duration (but make it snappier)
    const duration = Math.min(timingStore.GOLD_TRANSFORM_DURATION / 1000, 0.25)

    const tl = gsap.timeline({
      onComplete: () => {
        if (sprite && !sprite.destroyed) {
          sprite.scale.x = originalScaleX
          sprite.scale.y = originalScaleY
          sprite.alpha = 1
          sprite.tint = 0xffffff
        }
        animations.delete(cellKey)
      }
    })

    // Quick scale up punch
    tl.to(sprite.scale, {
      x: originalScaleX * 1.15,
      y: originalScaleY * 1.15,
      duration: duration * 0.4,
      ease: 'power2.out'
    })

    // Swap texture to wild at peak and update grid
    tl.call(() => {
      if (sprite && !sprite.destroyed) {
        const wildTex = getTextureForSymbol('wild')
        if (wildTex) {
          sprite.texture = wildTex
        }
        sprite.alpha = 1
        // Update grid
        if (onTextureSwap) {
          onTextureSwap(col, row)
        }
      }
    })

    // Snap back to original scale
    tl.to(sprite.scale, {
      x: originalScaleX,
      y: originalScaleY,
      duration: duration * 0.6,
      ease: 'back.out(2)'
    })

    anim.tween = tl
    animations.set(cellKey, anim)
  }

  /**
   * Update - GSAP handles animations automatically
   */
  function update(deltaTime: number = 1): void {
    // GSAP handles all animation updates automatically
  }

  /**
   * Clear all animations
   */
  function clear(): void {
    for (const [, anim] of animations) {
      if (anim.tween) {
        gsap.killTweensOf(anim.sprite?.scale)
      }
      if (anim.sprite && !anim.sprite.destroyed) {
        anim.sprite.scale.x = anim.originalScaleX
        anim.sprite.alpha = 1
        anim.sprite.tint = 0xffffff
      }
    }
    animations.clear()
  }

  /**
   * Check if a cell is currently animating
   */
  function isAnimating(cellKey: string): boolean {
    return animations.has(cellKey)
  }

  /**
   * Check if there are any active animations
   */
  function hasActiveAnimations(): boolean {
    return animations.size > 0
  }

  /**
   * Get animation progress for a cell (0-1)
   */
  function getAnimationProgress(cellKey: string): number {
    const anim = animations.get(cellKey)
    if (!anim || !anim.tween) return 1
    return anim.tween.progress()
  }

  return {
    container,
    startTransform,
    update,
    clear,
    isAnimating,
    hasActiveAnimations,
    getAnimationProgress
  }
}
