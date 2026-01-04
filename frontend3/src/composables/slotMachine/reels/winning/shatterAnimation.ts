import { Container, Sprite, Graphics } from 'pixi.js'
import { gsap } from 'gsap'
import { useTimingStore } from '@/stores'
import { getTextureForSymbol } from '../textures'

/**
 * Eat animation state for a tile
 */
interface EatState {
  completed: boolean
  sprite: Sprite
  mask: Graphics
  originalScale: { x: number; y: number }
  isGold: boolean
  timeline: gsap.core.Timeline | null
}

/**
 * Shatter animation manager interface
 */
export interface ShatterAnimationManager {
  container: Container
  startShatter: (key: string, sprite: Sprite, x: number, y: number, baseScaleX: number, baseScaleY: number, isGold?: boolean) => void
  update: () => void
  clear: () => void
  clearCompleted: () => void
  isAnimating: (key: string) => boolean
  hasCompleted: (key: string) => boolean
  hasAnyCompleted: () => boolean
  wasGoldAnimation: (key: string) => boolean
  isTransformedWild: (key: string) => boolean
}

/**
 * Creates an "eating" animation manager for winning tiles
 * Tiles are eaten from outside edges inward (shrinking rectangle mask)
 * Gold tiles: flip horizontally, swap texture to wild
 */
export function createShatterAnimationManager(): ShatterAnimationManager {
  const container = new Container()
  const timingStore = useTimingStore()

  const animations = new Map<string, EatState>()
  const completedAnimations = new Set<string>()
  const completedGoldAnimations = new Set<string>()
  const transformedWilds = new Set<string>()

  /**
   * Start an "eating" animation for a tile
   * The tile shrinks from edges inward like being eaten/consumed
   */
  function startShatter(
    key: string,
    sprite: Sprite,
    x: number,
    y: number,
    baseScaleX: number,
    baseScaleY: number,
    isGold: boolean = false
  ): void {
    if (animations.has(key)) return
    if (transformedWilds.has(key)) return

    completedAnimations.delete(key)

    const duration = (timingStore.FLIP_DURATION / 1000) * 0.8

    if (isGold) {
      // Gold tiles: use GSAP flip animation to transform to wild
      sprite.scale.x = baseScaleX
      sprite.scale.y = baseScaleY
      sprite.alpha = 1
      sprite.skew.set(0, 0)

      const state: EatState = {
        completed: false,
        sprite,
        mask: new Graphics(),
        originalScale: { x: baseScaleX, y: baseScaleY },
        isGold: true,
        timeline: null
      }

      const tl = gsap.timeline({
        onComplete: () => {
          state.completed = true
          completedAnimations.add(key)
          completedGoldAnimations.add(key)
          transformedWilds.add(key)
          animations.delete(key)

          if (sprite && !sprite.destroyed) {
            sprite.skew.set(0, 0)
          }
        }
      })

      tl.to(sprite.scale, {
        duration: duration / 2,
        x: 0,
        ease: 'power1.in'
      })

      tl.call(() => {
        if (sprite && !sprite.destroyed) {
          const wildTex = getTextureForSymbol('wild')
          if (wildTex) {
            sprite.texture = wildTex
          }
        }
      })

      tl.to(sprite.scale, {
        duration: duration / 2,
        x: baseScaleX,
        ease: 'power1.out'
      })

      state.timeline = tl
      animations.set(key, state)
    } else {
      // Normal tiles: eating animation from edges inward
      const width = sprite.width
      const height = sprite.height

      // Create a rectangular mask that will shrink
      const mask = new Graphics()
      mask.rect(-width / 2, -height / 2, width, height)
      mask.fill({ color: 0xffffff })
      mask.x = x
      mask.y = y

      // Apply mask to sprite
      sprite.mask = mask
      container.addChild(mask)

      const state: EatState = {
        completed: false,
        sprite,
        mask,
        originalScale: { x: baseScaleX, y: baseScaleY },
        isGold: false,
        timeline: null
      }

      // Animate the mask shrinking from edges (eating effect)
      // Use object to animate mask dimensions
      const maskSize = { w: width, h: height }

      const tl = gsap.timeline({
        onUpdate: () => {
          // Redraw mask with current size
          mask.clear()
          mask.rect(-maskSize.w / 2, -maskSize.h / 2, maskSize.w, maskSize.h)
          mask.fill({ color: 0xffffff })
        },
        onComplete: () => {
          // Clean up
          sprite.mask = null
          container.removeChild(mask)
          mask.destroy()

          state.completed = true
          completedAnimations.add(key)
          animations.delete(key)

          if (sprite && !sprite.destroyed) {
            sprite.alpha = 0
          }
        }
      })

      // Smooth chomping animation - multiple bites with fluid easing
      const biteCount = 5
      const biteDuration = duration / biteCount

      for (let i = 0; i < biteCount; i++) {
        const progress = (i + 1) / biteCount
        // Ease the target sizes for natural deceleration feel
        const easedProgress = 1 - Math.pow(1 - progress, 1.5)
        const targetW = width * (1 - easedProgress)
        const targetH = height * (1 - easedProgress)

        // Smooth sine easing for fluid bites, slight overlap for blending
        tl.to(maskSize, {
          duration: biteDuration,
          w: targetW,
          h: targetH,
          ease: 'sine.inOut'
        }, i === 0 ? 0 : `-=${biteDuration * 0.15}`) // 15% overlap between bites
      }

      state.timeline = tl
      animations.set(key, state)
    }
  }

  /**
   * Update - GSAP handles animations automatically
   */
  function update(): void {
    // GSAP handles all animation updates
  }

  /**
   * Clear all animations
   */
  function clear(): void {
    for (const [, state] of animations) {
      if (state.timeline) {
        state.timeline.kill()
      }

      // Clean up mask
      if (state.mask && !state.mask.destroyed) {
        if (state.sprite) {
          state.sprite.mask = null
        }
        container.removeChild(state.mask)
        state.mask.destroy()
      }

      // Reset sprite state
      if (state.sprite && !state.sprite.destroyed) {
        gsap.killTweensOf(state.sprite)
        gsap.killTweensOf(state.sprite.scale)
        state.sprite.alpha = 1
        state.sprite.scale.x = state.originalScale.x
        state.sprite.scale.y = state.originalScale.y
        state.sprite.skew.set(0, 0)
        state.sprite.mask = null
      }
    }
    animations.clear()
    completedAnimations.clear()
    completedGoldAnimations.clear()
    transformedWilds.clear()
  }

  function isAnimating(key: string): boolean {
    return animations.has(key)
  }

  function hasCompleted(key: string): boolean {
    return completedAnimations.has(key)
  }

  function hasAnyCompleted(): boolean {
    return completedAnimations.size > 0
  }

  function clearCompleted(): void {
    completedAnimations.clear()
    completedGoldAnimations.clear()
    transformedWilds.clear()
  }

  function wasGoldAnimation(key: string): boolean {
    return completedGoldAnimations.has(key)
  }

  function isTransformedWild(key: string): boolean {
    return transformedWilds.has(key)
  }

  return {
    container,
    startShatter,
    update,
    clear,
    clearCompleted,
    isAnimating,
    hasCompleted,
    hasAnyCompleted,
    wasGoldAnimation,
    isTransformedWild
  }
}
