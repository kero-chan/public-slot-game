import { gsap } from 'gsap'
import type { Sprite } from 'pixi.js'

/**
 * Animation data interface
 */
interface BumpAnimation {
  tween: gsap.core.Timeline | gsap.core.Tween
  sprite: Sprite
  baseScaleX: number
  baseScaleY: number
  isPulsing?: boolean  // Track if this is a continuous pulse
}

/**
 * Bump animation manager interface
 */
export interface BumpAnimationManager {
  startBump: (key: string, sprite: Sprite) => void
  update: () => void
  clear: () => void
  reset: (key: string) => void
  isAnimating: (key: string) => boolean
  isAlreadyBumped: (key: string) => boolean
}

/**
 * Manages bump animations for bonus tiles when they appear
 * NOW USING GSAP for smoother animations
 * After initial bump, continues with a subtle pulse animation
 */
export function createBumpAnimationManager(): BumpAnimationManager {
  const animations = new Map<string, BumpAnimation>() // key -> { tween, sprite, baseScaleX, baseScaleY }
  const bumpedTiles = new Set<string>() // Track tiles that have already been bumped
  const pulsingTiles = new Map<string, BumpAnimation>() // Track tiles with continuous pulse

  /**
   * Start a continuous subtle pulse animation after initial bump
   */
  function startPulse(key: string, sprite: Sprite, baseScaleX: number, baseScaleY: number): void {
    if (!sprite || sprite.destroyed || pulsingTiles.has(key)) return

    // Subtle continuous pulse - smaller scale change than initial bump
    const pulseTween = gsap.to(sprite.scale, {
      x: baseScaleX * 1.05,  // Only 5% scale increase
      y: baseScaleY * 1.05,
      duration: 0.8,
      ease: 'sine.inOut',
      repeat: -1,  // Infinite repeat
      yoyo: true,  // Bounce back and forth
    })

    pulsingTiles.set(key, {
      tween: pulseTween,
      sprite,
      baseScaleX,
      baseScaleY,
      isPulsing: true
    })
  }

  /**
   * Start a bump animation for a tile using GSAP
   * @param key - Tile key (e.g., "0:1")
   * @param sprite - The sprite to animate
   */
  function startBump(key: string, sprite: Sprite): void {
    if (!sprite || animations.has(key) || bumpedTiles.has(key)) return

    const baseScaleX = sprite.scale.x
    const baseScaleY = sprite.scale.y

    // Create a timeline for the bump animation
    const timeline = gsap.timeline({
      onComplete: () => {
        // Don't manually reset scale - let render loop handle it
        // Just mark as bumped and remove from animations
        bumpedTiles.add(key)
        animations.delete(key)
        // Note: Disabled continuous pulse - it causes scale conflicts
      }
    })

    // Bump animation: scale up with bounce, then yoyo back
    timeline.to(sprite.scale, {
      x: baseScaleX * 1.15,
      y: baseScaleY * 1.15,
      duration: 0.2,
      ease: 'back.out(2)',
      yoyo: true,
      repeat: 1
    })

    animations.set(key, {
      tween: timeline,
      sprite: sprite,
      baseScaleX,
      baseScaleY
    })
  }

  /**
   * Update - GSAP handles updates automatically, this is a no-op now
   */
  function update(): void {
    // GSAP handles all updates automatically
  }

  /**
   * Clear all animations and tracking - kill GSAP tweens
   */
  function clear(): void {
    // Clear initial bump animations - don't reset scale, let render loop handle it
    for (const [, anim] of animations.entries()) {
      if (anim.tween) {
        anim.tween.kill()
      }
    }
    animations.clear()

    // Clear pulsing animations (legacy, no longer started)
    for (const [, anim] of pulsingTiles.entries()) {
      if (anim.tween) {
        anim.tween.kill()
      }
    }
    pulsingTiles.clear()

    bumpedTiles.clear()
  }

  /**
   * Reset a specific tile (when it's reused or removed) - kill GSAP tween
   */
  function reset(key: string): void {
    // Reset initial bump animation - don't reset scale, let render loop handle it
    const anim = animations.get(key)
    if (anim?.tween) {
      anim.tween.kill()
    }
    animations.delete(key)

    // Reset pulsing animation (legacy)
    const pulseAnim = pulsingTiles.get(key)
    if (pulseAnim?.tween) {
      pulseAnim.tween.kill()
    }
    pulsingTiles.delete(key)

    bumpedTiles.delete(key)
  }

  /**
   * Check if a tile is currently doing initial bump animation
   * Note: Does NOT include continuous pulse - pulse should not block scale updates
   */
  function isAnimating(key: string): boolean {
    return animations.has(key)
  }

  /**
   * Check if a tile is pulsing (continuous animation after initial bump)
   */
  function isPulsing(key: string): boolean {
    return pulsingTiles.has(key)
  }

  /**
   * Check if a tile has already been bumped
   */
  function isAlreadyBumped(key: string): boolean {
    return bumpedTiles.has(key)
  }

  return {
    startBump,
    update,
    clear,
    reset,
    isAnimating,
    isAlreadyBumped
  }
}
