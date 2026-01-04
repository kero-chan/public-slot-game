import { gsap } from 'gsap'
import type { Sprite } from 'pixi.js'

/**
 * Pop animation data interface
 */
interface PopAnimation {
  tween: gsap.core.Tween
  sprite: Sprite
  baseScaleX: number
  baseScaleY: number
}

/**
 * Pop animation manager interface
 */
export interface PopAnimationManager {
  startPop: (cellKey: string, sprite: Sprite) => void
  update: () => boolean
  isAnimating: (cellKey: string) => boolean
  hasActiveAnimations: () => boolean
  clear: () => void
}

/**
 * Pop Animation Manager
 * Handles tile pop animations (scale bounce effect) for bonus tiles during jackpot
 * NOW USING GSAP for smoother animations with built-in back easing
 */
export function createPopAnimationManager(): PopAnimationManager {
  const activeAnimations = new Map<string, PopAnimation>() // cellKey -> { tween, sprite, baseScaleX, baseScaleY }

  /**
   * Start a pop animation for a tile using GSAP
   * @param cellKey - The cell key (e.g., "0:1")
   * @param sprite - The tile sprite to animate
   */
  function startPop(cellKey: string, sprite: Sprite): void {
    if (!sprite) return

    const baseScaleX = sprite.scale.x
    const baseScaleY = sprite.scale.y
    const maxScale = 1.7 // Pop to 170% size - bigger burst!

    // Kill any existing animation for this tile
    const existing = activeAnimations.get(cellKey)
    if (existing && existing.tween) {
      existing.tween.kill()
    }

    // Create GSAP tween with back easing for overshoot effect
    // Use yoyo to animate back to original scale smoothly
    const tween = gsap.to(sprite.scale, {
      x: baseScaleX * maxScale,
      y: baseScaleY * maxScale,
      duration: 0.25, // 250ms up
      ease: 'back.out(1.7)', // Built-in back easing with overshoot
      yoyo: true,
      repeat: 1, // Go up then back down
      onComplete: () => {
        // Don't manually reset scale - let render loop handle correct scale
        // Just remove from active animations so isAnimating returns false
        activeAnimations.delete(cellKey)
      }
    })

    activeAnimations.set(cellKey, {
      tween,
      sprite,
      baseScaleX,
      baseScaleY
    })
  }

  /**
   * Update - GSAP handles updates automatically, this just checks for active animations
   * @returns True if any animations are still active
   */
  function update(): boolean {
    // GSAP handles all updates automatically
    // Just return whether we have active animations
    return activeAnimations.size > 0
  }

  /**
   * Check if a tile is currently animating
   */
  function isAnimating(cellKey: string): boolean {
    return activeAnimations.has(cellKey)
  }

  /**
   * Check if any animations are active
   */
  function hasActiveAnimations(): boolean {
    return activeAnimations.size > 0
  }

  /**
   * Clear all animations - kill GSAP tweens
   */
  function clear(): void {
    // Kill all tweens - don't reset scale, let render loop handle it
    for (const [_cellKey, anim] of activeAnimations.entries()) {
      if (anim.tween) {
        anim.tween.kill()
      }
    }
    activeAnimations.clear()
  }

  return {
    startPop,
    update,
    isAnimating,
    hasActiveAnimations,
    clear,
  }
}
