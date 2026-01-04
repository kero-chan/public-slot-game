import { useTimingStore } from '@/stores'
import { gsap } from 'gsap'
import type { Sprite, Texture } from 'pixi.js'

/**
 * Drop state interface
 */
interface DropState {
  fromY: number
  toY: number
  currentY: number
  tween: gsap.core.Tween
  symbol: string
}

/**
 * Completed state interface
 */
interface CompletedState {
  symbol: string
  completedAt: number
}

/**
 * Drop animation manager interface
 */
export interface DropAnimationManager {
  startDrop: (
    key: string,
    sprite: Sprite,
    fromY: number,
    toY: number,
    symbol: string,
    getTexture: ((symbol: string) => Texture | null) | null
  ) => void
  update: () => void
  getDropY: (key: string, baseY: number) => number
  isDropping: (key: string) => boolean
  hasActiveDrops: () => boolean
  getAnimatingSymbol: (key: string) => string | null
  getCompletedSymbol: (key: string) => string | null
  isRecentlyCompleted: (key: string) => boolean
  clear: () => void
  clearCompleted: () => void
}

/**
 * Manages drop animations for tiles cascading down
 * NOW USING GSAP for smoother, more performant animations
 */
export function createDropAnimationManager(): DropAnimationManager {
  const timingStore = useTimingStore()
  const dropStates = new Map<string, DropState>()
  const completedStates = new Map<string, CompletedState>()

  /**
   * Start a drop animation for a tile using GSAP
   * OPTIMIZED: GSAP animates sprite.y directly for butter-smooth 60fps animation
   * @param key - The cellKey (col:visualRow)
   * @param sprite - The sprite to animate
   * @param fromY - Starting Y position
   * @param toY - Target Y position
   * @param symbol - The symbol this sprite should display during animation
   * @param getTexture - Function to get texture for a symbol
   */
  function startDrop(
    key: string,
    sprite: Sprite,
    fromY: number,
    toY: number,
    symbol: string,
    getTexture: ((symbol: string) => Texture | null) | null
  ): void {
    if (!sprite) return

    // CRITICAL: Set the sprite's texture AND position IMMEDIATELY to prevent visible flashing
    // The sprite must show the correct symbol at the correct starting position
    if (getTexture) {
      const tex = getTexture(symbol)
      if (tex && sprite.texture !== tex) {
        sprite.texture = tex
      }
    }

    // Set sprite to starting Y position immediately
    sprite.y = fromY

    // Kill any existing tween for this sprite
    const existingDrop = dropStates.get(key)
    if (existingDrop && existingDrop.tween) {
      existingDrop.tween.kill()
    }

    // GSAP OPTIMIZATION: Animate sprite.y directly (no intermediate object)
    // This allows GSAP to use optimized transforms and provides smoothest animation
    const tween = gsap.to(sprite, {
      y: toY,
      duration: timingStore.DROP_DURATION / 1000, // Convert to seconds (default 650ms = 0.65s)
      ease: 'power1.out', // Smooth deceleration - starts fast, slows down for soft landing
      overwrite: 'auto', // Automatically kill conflicting tweens
      onComplete: () => {
        // Animation complete - move to completed states to preserve symbol
        completedStates.set(key, {
          symbol: symbol,
          completedAt: Date.now()
        })
        dropStates.delete(key)
      }
    })

    dropStates.set(key, {
      fromY,
      toY,
      currentY: fromY, // Updated by getDropY
      tween,
      symbol // Store the symbol for this animation
    })

    // Clear from completed states if it was there
    completedStates.delete(key)
  }

  /**
   * Update - GSAP handles animation updates automatically
   * This function only cleans up completed states after grace period
   */
  function update(): void {
    // Fast-path: skip iteration if no completed states to clean up
    if (completedStates.size === 0) return

    const now = Date.now()

    // Auto-clear completed states after grace period
    // This prevents delays when waiting for drops to finish before showing win announcements
    for (const [key, completed] of completedStates.entries()) {
      if (now - completed.completedAt > timingStore.DROP_GRACE_PERIOD) {
        completedStates.delete(key)
      }
    }
  }

  /**
   * Get the current Y position for a dropping tile
   * GSAP animates sprite.y directly, so we just return baseY (sprite position is handled by GSAP)
   */
  function getDropY(key: string, baseY: number): number {
    const drop = dropStates.get(key)
    if (!drop) return baseY

    // Since GSAP animates sprite.y directly, return baseY
    // The sprite's actual Y position is managed by GSAP's tween
    return baseY
  }

  /**
   * Check if a tile is currently dropping
   */
  function isDropping(key: string): boolean {
    return dropStates.has(key)
  }

  /**
   * Check if ANY tiles are currently dropping
   * Only checks active animations, not completed ones in grace period
   */
  function hasActiveDrops(): boolean {
    return dropStates.size > 0
  }

  /**
   * Get the symbol for an animating sprite
   * Returns null if not animating
   */
  function getAnimatingSymbol(key: string): string | null {
    const drop = dropStates.get(key)
    return drop ? drop.symbol : null
  }

  /**
   * Get the symbol for a sprite that just completed its animation
   * Returns null if not recently completed
   */
  function getCompletedSymbol(key: string): string | null {
    const completed = completedStates.get(key)
    return completed ? completed.symbol : null
  }

  /**
   * Check if a sprite recently completed its drop animation
   */
  function isRecentlyCompleted(key: string): boolean {
    return completedStates.has(key)
  }

  /**
   * Clear all animations - kill GSAP tweens
   */
  function clear(): void {
    // Kill all active GSAP tweens
    for (const [key, drop] of dropStates.entries()) {
      if (drop.tween) {
        drop.tween.kill()
      }
    }
    dropStates.clear()
    completedStates.clear()
  }

  /**
   * Clear only completed states (used when new cascade starts)
   */
  function clearCompleted(): void {
    completedStates.clear()
  }

  return {
    startDrop,
    update,
    getDropY,
    isDropping,
    hasActiveDrops,
    getAnimatingSymbol,
    getCompletedSymbol,
    isRecentlyCompleted,
    clear,
    clearCompleted
  }
}
