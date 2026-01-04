import { Container, Sprite } from 'pixi.js'
import gsap from 'gsap'
import { getGlyphSprite, getBackgroundSprite } from '@/config/spritesheet'
import { WIN_ANIMATION_LAYOUT } from './winAnimationConstants'

export interface FreeSpinCountdownOverlay {
  container: Container
  /**
   * Show the countdown overlay with N → N-1 reduction animation
   * @param spinsRemaining - The number of spins remaining (N)
   * @param canvasWidth - Canvas width
   * @param canvasHeight - Canvas height
   */
  show: (spinsRemaining: number, canvasWidth: number, canvasHeight: number) => Promise<void>
  hide: () => void
  isShowing: () => boolean
}

/**
 * Creates a free spin countdown overlay that briefly shows remaining spins
 * before each free spin starts, with number reduction animation
 */
export function createFreeSpinCountdown(): FreeSpinCountdownOverlay {
  const container = new Container()
  container.visible = false
  container.zIndex = 1050 // Above game but below bonus overlay

  let isAnimating = false
  let resolvePromise: (() => void) | null = null
  let currentTimeline: gsap.core.Timeline | null = null

  /**
   * Create number sprites for a given number
   */
  function createNumberSprites(num: number, scale: number): { sprites: Sprite[], width: number } {
    const digits = String(num).split('')
    const sprites: Sprite[] = []
    let totalWidth = 0

    for (const d of digits) {
      // Use the gold number glyphs available in frontend3
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
    canvasWidth: number,
    canvasHeight: number,
    bgKey: string
  ): { width: number; height: number; scale: number } {
    const bgSprite = getBackgroundSprite(bgKey)
    if (!bgSprite) {
      // Fallback if image not found
      return { width: 300, height: 150, scale: 1 }
    }

    // Get max dimensions from constants
    const maxWidth = Math.min(
      canvasWidth * 0.85,
      canvasWidth * WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT
    )
    const maxHeight = canvasHeight * 0.25 // Max 25% of canvas height

    // Calculate scale to fit within max dimensions while maintaining aspect ratio
    const scaleX = maxWidth / bgSprite.width
    const scaleY = maxHeight / bgSprite.height
    const scale = Math.min(scaleX, scaleY, 1.0) // Don't scale up, only down

    return {
      width: bgSprite.width * scale,
      height: bgSprite.height * scale,
      scale
    }
  }

  /**
   * Show the countdown overlay with N → N-1 reduction animation
   * @param spinsRemaining - The number of spins remaining (N)
   * @param canvasWidth - Canvas width
   * @param canvasHeight - Canvas height
   * Returns a promise that resolves when the animation is complete
   */
  async function show(spinsRemaining: number, canvasWidth: number, canvasHeight: number): Promise<void> {
    // Kill any existing animation before starting a new one
    if (currentTimeline) {
      currentTimeline.kill()
      currentTimeline = null
    }

    // Resolve any pending promise from previous animation
    if (resolvePromise) {
      resolvePromise()
      resolvePromise = null
    }

    return new Promise((resolve) => {
      resolvePromise = resolve
      container.removeChildren()
      container.visible = true
      isAnimating = true

      const centerX = canvasWidth / 2
      const centerY = canvasHeight / 2

      const bgKey = 'free_spin_background.webp'

      // Calculate background dimensions
      const { width: bgWidth, height: bgHeight, scale: bgScale } = calculateBackgroundDimensions(
        canvasWidth,
        canvasHeight,
        bgKey
      )

      // Create background sprite
      const bgSprite = getBackgroundSprite(bgKey)
      if (!bgSprite) {
        resolve()
        return
      }

      bgSprite.anchor.set(0.5)
      bgSprite.scale.set(bgScale)
      bgSprite.x = centerX
      bgSprite.y = centerY

      // Create content container
      const contentContainer = new Container()
      contentContainer.addChild(bgSprite)

      // Check if this is the last spin (spinsRemaining - 1 === 0)
      const isLastSpin = spinsRemaining - 1 === 0

      // Only create number display if not last spin
      let oldNumberContainer: Container | null = null
      let newNumberContainer: Container | null = null

      if (!isLastSpin) {
        // Calculate number scale - number height should be 25% of background height
        // Also ensure it doesn't exceed background width or game view bounds
        const targetNumberHeight = bgHeight * 0.25
        const maxNumberWidth = Math.min(
          bgWidth * 0.60, // Max 60% of background width
          canvasWidth * WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT // Max 60% of game view width
        )
        
        const tempNumberData = createNumberSprites(spinsRemaining, 1.0)
        const tempNumberHeight = tempNumberData.sprites[0]?.height || 100
        const tempNumberWidth = tempNumberData.width
        
        const scaleByHeight = targetNumberHeight / tempNumberHeight
        const scaleByWidth = maxNumberWidth / tempNumberWidth
        const numberScale = Math.min(scaleByHeight, scaleByWidth)

        // Create number containers
        const numberContainer = new Container()
        oldNumberContainer = new Container()
        newNumberContainer = new Container()

        // N → N-1 animation
        const oldNumber = spinsRemaining
        const newNumber = spinsRemaining - 1

        // Create old number sprites (N)
        const oldNumData = createNumberSprites(oldNumber, numberScale)
        let offsetX = -oldNumData.width / 2
        for (const sprite of oldNumData.sprites) {
          sprite.x = offsetX + sprite.width / 2
          sprite.y = 0
          oldNumberContainer.addChild(sprite)
          offsetX += sprite.width
        }

        // Create new number sprites (N-1)
        const newNumData = createNumberSprites(newNumber, numberScale)
        offsetX = -newNumData.width / 2
        for (const sprite of newNumData.sprites) {
          sprite.x = offsetX + sprite.width / 2
          sprite.y = 0
          newNumberContainer.addChild(sprite)
          offsetX += sprite.width
        }

        // Initially hide new number (below)
        newNumberContainer.alpha = 0
        newNumberContainer.y = 50

        numberContainer.addChild(oldNumberContainer)
        numberContainer.addChild(newNumberContainer)

        // Position number container in background - offset based on background width, not canvas width
        // This ensures consistent positioning across mobile and PC
        const numberOffsetX = bgWidth * 0.20 // 20% of background width (equivalent to centerX/5 but relative to bg)
        numberContainer.x = centerX + numberOffsetX
        numberContainer.y = centerY
        contentContainer.addChild(numberContainer)
      }

      container.addChild(contentContainer)

      // Animate: scale in, show old number (if not last spin), animate number reduction (if not last spin), scale out
      contentContainer.scale.set(0)
      contentContainer.alpha = 0

      currentTimeline = gsap.timeline({
        onComplete: () => {
          currentTimeline = null
          container.visible = false
          isAnimating = false
          container.removeChildren()
          if (resolvePromise) {
            resolvePromise()
            resolvePromise = null
          }
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

      if (!isLastSpin && oldNumberContainer && newNumberContainer) {
        // Hold showing old number
        tl.to({}, { duration: 0.4 })

        // Animate number reduction: old number slides up and fades out
        tl.to(oldNumberContainer, {
          y: -50,
          alpha: 0,
          duration: 0.3,
          ease: 'power2.in'
        })

        // New number slides up and fades in
        tl.to(newNumberContainer, {
          y: 0,
          alpha: 1,
          duration: 0.3,
          ease: 'power2.out'
        }, '-=0.2')

        // Pulse effect on new number
        tl.to(newNumberContainer.scale, {
          x: 1.2,
          y: 1.2,
          duration: 0.15,
          ease: 'power2.out'
        })
        tl.to(newNumberContainer.scale, {
          x: 1,
          y: 1,
          duration: 0.15,
          ease: 'power2.in'
        })

        // Hold showing new number
        tl.to({}, { duration: 0.3 })
      } else {
        // Last spin: just hold showing background
        tl.to({}, { duration: 0.8 })
      }

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
    })
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
    container.removeChildren()
    if (resolvePromise) {
      resolvePromise()
      resolvePromise = null
    }
  }

  return {
    container,
    show,
    hide,
    isShowing: () => isAnimating || container.visible
  }
}
