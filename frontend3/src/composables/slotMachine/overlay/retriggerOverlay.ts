import { Container, Sprite } from 'pixi.js'
import gsap from 'gsap'
import { getGlyphSprite, getBackgroundSprite } from '@/config/spritesheet'
import { WIN_ANIMATION_LAYOUT } from './winAnimationConstants'
import type { BaseOverlay } from './base'

/**
 * Retrigger overlay interface
 */
export interface RetriggerOverlay extends BaseOverlay {
  show: (additionalSpins: number, cWidth: number, cHeight: number, onComplete: () => void) => void
}

/**
 * Creates a retrigger overlay that displays when free spins are retriggered
 * Simple animation matching freeSpinCountdown behavior
 */
export function createRetriggerOverlay(): RetriggerOverlay {
  const container = new Container()
  container.visible = false
  container.zIndex = 1100

  let isAnimating = false
  let storedOnComplete: (() => void) | null = null
  let currentTimeline: gsap.core.Timeline | null = null

  /**
   * Create number sprites for a given number
   */
  function createNumberSprites(num: number, scale: number): { sprites: Sprite[], width: number } {
    const digits = String(num).split('')
    const sprites: Sprite[] = []
    let totalWidth = 0

    for (const d of digits) {
      const sprite = getGlyphSprite(`glyph_${d}.webp`)
      if (sprite) {
        sprite.scale.set(scale)
        sprite.anchor.set(0.5, 0.5)
        sprites.push(sprite)
        totalWidth += sprite.width
      }
    }

    return { sprites, width: totalWidth }
  }

  /**
   * Calculate background image dimensions ensuring it fits within game view
   */
  function calculateBackgroundDimensions(
    cWidth: number,
    cHeight: number,
    bgKey: string
  ): { width: number; height: number; scale: number } {
    const bgSprite = getBackgroundSprite(bgKey)
    if (!bgSprite) {
      return { width: 300, height: 150, scale: 1 }
    }

    const maxWidth = Math.min(
      cWidth * 0.85,
      cWidth * WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT
    )
    const maxHeight = cHeight * 0.25

    const scaleX = maxWidth / bgSprite.width
    const scaleY = maxHeight / bgSprite.height
    const scale = Math.min(scaleX, scaleY, 1.0)

    return {
      width: bgSprite.width * scale,
      height: bgSprite.height * scale,
      scale
    }
  }

  /**
   * Show the overlay - simple animation like freeSpinCountdown
   */
  function show(additionalSpins: number, cWidth: number, cHeight: number, onComplete: () => void): void {
    container.removeChildren()
    container.visible = true
    isAnimating = true
    storedOnComplete = onComplete

    const centerX = cWidth / 2
    const centerY = cHeight / 2

    // Use the retrigger background image
    const bgKey = 'free_spin_retrigger_background.webp'
    const { width: bgWidth, height: bgHeight, scale: bgScale } = calculateBackgroundDimensions(
      cWidth,
      cHeight,
      bgKey
    )

    // Create content container
    const contentContainer = new Container()

    // Create background sprite
    const bgSprite = getBackgroundSprite(bgKey)
    if (!bgSprite) {
      onComplete()
      return
    }

    bgSprite.anchor.set(0.5)
    bgSprite.scale.set(bgScale)
    bgSprite.x = centerX
    bgSprite.y = centerY
    contentContainer.addChild(bgSprite)

    // Calculate number scale - number height should be 25% of background height
    const targetNumberHeight = bgHeight * 0.25
    const maxNumberWidth = Math.min(
      bgWidth * 0.60,
      cWidth * WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT
    )

    const tempNumberData = createNumberSprites(additionalSpins, 1.0)
    const tempNumberHeight = tempNumberData.sprites[0]?.height || 100
    const tempNumberWidth = tempNumberData.width

    const scaleByHeight = targetNumberHeight / tempNumberHeight
    const scaleByWidth = maxNumberWidth / tempNumberWidth
    const numberScale = Math.min(scaleByHeight, scaleByWidth)

    // Create number container
    const numberContainer = new Container()
    const numData = createNumberSprites(additionalSpins, numberScale)
    let offsetX = -numData.width / 2
    for (const sprite of numData.sprites) {
      sprite.x = offsetX + sprite.width / 2
      sprite.y = 0
      numberContainer.addChild(sprite)
      offsetX += sprite.width
    }

    // Position number in the "hole" of the background (right side offset like freeSpinCountdown)
    const numberOffsetX = bgWidth * 0.20
    numberContainer.x = centerX + numberOffsetX
    numberContainer.y = centerY
    contentContainer.addChild(numberContainer)

    container.addChild(contentContainer)

    // Animate: scale in, hold, scale out (matching freeSpinCountdown)
    contentContainer.scale.set(0)
    contentContainer.alpha = 0

    currentTimeline = gsap.timeline({
      onComplete: () => {
        const callback = storedOnComplete
        storedOnComplete = null
        currentTimeline = null
        container.visible = false
        isAnimating = false
        container.removeChildren()
        if (callback) callback()
      }
    })
    const tl = currentTimeline

    // Scale in
    tl.to(contentContainer, {
      alpha: 1,
      duration: 0.2,
      ease: 'power2.out'
    }, 0)
    tl.to(contentContainer.scale, {
      x: 1,
      y: 1,
      duration: 0.3,
      ease: 'back.out(1.5)'
    }, 0)

    // Pulse effect on number
    tl.to(numberContainer.scale, {
      x: 1.2,
      y: 1.2,
      duration: 0.15,
      ease: 'power2.out'
    }, 0.4)
    tl.to(numberContainer.scale, {
      x: 1,
      y: 1,
      duration: 0.15,
      ease: 'power2.in'
    }, 0.55)

    // Hold
    tl.to({}, { duration: 0.8 })

    // Scale out
    tl.to(contentContainer, {
      alpha: 0,
      duration: 0.2,
      ease: 'power2.in'
    })
    tl.to(contentContainer.scale, {
      x: 0.8,
      y: 0.8,
      duration: 0.2,
      ease: 'power2.in'
    }, '-=0.2')
  }

  /**
   * Hide the overlay immediately
   */
  function hide(): void {
    if (currentTimeline) {
      currentTimeline.kill()
      currentTimeline = null
    }
    gsap.killTweensOf(container)
    container.visible = false
    isAnimating = false
    storedOnComplete = null
    container.removeChildren()
  }

  /**
   * Update animation (no-op for simple overlay)
   */
  function update(_timestamp: number): void {
    // No continuous updates needed
  }

  /**
   * Build/rebuild for canvas resize
   */
  function build(_cWidth: number, _cHeight: number): void {
    // No rebuild needed
  }

  return {
    container,
    show,
    hide,
    update,
    build,
    isShowing: () => isAnimating || container.visible
  }
}
