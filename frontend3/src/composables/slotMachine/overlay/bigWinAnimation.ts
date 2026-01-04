/**
 * Big Win Animation
 * Dramatic celebration for big wins
 * - Winning tiles zoom to center and explode
 * - Multiple shockwave bursts
 * - Fireworks display
 * - Screen shake
 * - Duration (~2.5s)
 */

import { Container, Graphics, Sprite, Texture, Text } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import gsap from 'gsap'
import {
  createTimerManager,
  createParticleSystem,
  createScreenShake,
  createDarkOverlay,
  createNumberDisplay,
  clearContainer,
  type BaseOverlay
} from './base'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'
import { createWinImageSprite } from './winImageUtils'
import { getGlyphSprite } from '@/config/spritesheet'
import type { TilePosition } from './megaWinAnimation'
import {
  WIN_ANIMATION_LAYOUT,
  WIN_TEXT_HEIGHT_PERCENT,
  WIN_ANIMATION_DURATION,
  WIN_IMAGE_KEYS,
  WIN_TEXT_FALLBACK,
  WIN_COLORS,
  WIN_IMAGE_SIZING,
  isMobileViewport,
  getAmountDisplayScale,
  calculateWinImageScale
} from './winAnimationConstants'
import { getCounterDuration } from '@/utils/gameHelpers'

// Configuration - responsive sizing based on canvas dimensions
const WIN_CONFIG = {
  imageKey: WIN_IMAGE_KEYS.BIG,
  targetHeightPercent: WIN_TEXT_HEIGHT_PERCENT.BIG,
  maxWidthPercent: WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT,
  textFallback: WIN_TEXT_FALLBACK.BIG
}
const HOLD_DURATION = WIN_ANIMATION_DURATION.BIG

// Mobile detection for performance optimization
const IS_MOBILE = typeof window !== 'undefined' && isMobileViewport(window.innerWidth)

// Particle counts - reduced on mobile for better performance
const PARTICLE_COUNTS = {
  convergeSparks: IS_MOBILE ? 2 : 5,
  explosionBurst: IS_MOBILE ? 40 : 120,
  fireworkCount: IS_MOBILE ? 20 : 40
}

/**
 * Big win animation interface
 */
export interface BigWinAnimation extends BaseOverlay {
  show: (canvasWidth: number, canvasHeight: number, tilePositions: TilePosition[], amount: number, onComplete?: () => void) => void
}

/**
 * Creates a dramatic big win animation
 * Tiles converge to center, explode into particles, fireworks burst
 */
export function createBigWinAnimation(reelsRef?: any): BigWinAnimation {
  let reels = reelsRef

  const container = new Container()
  container.visible = false
  container.zIndex = 2000

  // Sub-containers
  const backgroundContainer = new Container()
  const tileContainer = new Container()
  const effectsContainer = new Container()
  const particleContainer = new Container()
  const textContainer = new Container()
  const skipButtonContainer = new Container()

  container.addChild(backgroundContainer)
  container.addChild(effectsContainer)
  container.addChild(tileContainer)
  container.addChild(particleContainer)
  container.addChild(textContainer)
  container.addChild(skipButtonContainer)

  // Managers
  const timers = createTimerManager()
  const particles = createParticleSystem(particleContainer)
  const screenShake = createScreenShake(container)

  // State
  let isAnimating = false
  let animationTimeline: gsap.core.Timeline | null = null
  let counterTween: gsap.core.Tween | null = null
  let canvasW = 0
  let canvasH = 0
  let storedOnComplete: (() => void) | null = null

  /**
   * Play generic UI sound with pitch randomization (0.6-1.4)
   */
  function playGenericUISound(): void {
    const howl = howlerAudio.getHowl('generic_ui')
    if (!howl) return

    const randomPitch = 0.6 + Math.random() * 0.8
    howl.rate(randomPitch)

    audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'generic_ui', volume: 0.6 })
  }

  /**
   * Create skip button
   */
  function createSkipButton(width: number, height: number): void {
    skipButtonContainer.removeChildren()

    const skipButton = getGlyphSprite('glyph_skip_button.webp')
    if (!skipButton) return

    const isMobile = width < 600
    const buttonScale = isMobile ? 0.084 : 0.105
    skipButton.scale.set(buttonScale)
    skipButton.anchor.set(0.5, 1)
    skipButton.x = width / 2
    skipButton.y = height - 20

    skipButton.eventMode = 'static'
    skipButton.cursor = 'pointer'
    ;(skipButton as any).on('pointerdown', () => {
      if (isAnimating && storedOnComplete) {
        playGenericUISound()
        const callback = storedOnComplete
        storedOnComplete = null
        hide()
        callback()
      }
    })

    ;(skipButton as any).on('pointerover', () => {
      gsap.to(skipButton.scale, { x: buttonScale * 1.1, y: buttonScale * 1.1, duration: 0.15 })
    })
    ;(skipButton as any).on('pointerout', () => {
      gsap.to(skipButton.scale, { x: buttonScale, y: buttonScale, duration: 0.15 })
    })

    skipButtonContainer.addChild(skipButton)
    skipButton.alpha = 0
    gsap.to(skipButton, { alpha: 1, duration: 0.3, delay: 0.5 })
  }

  /**
   * Create converging tile sprite
   * Size is calculated based on viewport dimensions to ensure proper display
   */
  function createConvergingTile(
    pos: TilePosition,
    index: number,
    total: number,
    centerX: number,
    centerY: number,
    viewportWidth: number,
    viewportHeight: number
  ): Sprite | Graphics {
    let tile: Sprite | Graphics
    let initialScale = 1

    // Calculate proper tile size based on viewport
    // Use larger tiles for better visibility during win animations
    const viewportMin = Math.min(viewportWidth, viewportHeight)
    const targetTileSize = viewportMin * 0.3 // 30% of smaller dimension

    // Try to clone actual tile sprite
    if (reels && reels.getSpriteCache) {
      const spriteCache = reels.getSpriteCache()
      const visualRow = pos.row - 4
      const cellKey = `${pos.col}:${visualRow}`
      const originalSprite = spriteCache.get(cellKey)

      if (originalSprite && originalSprite.texture) {
        tile = new Sprite(originalSprite.texture)
        tile.anchor.set(0.5, 0.5)
        // Calculate scale to match viewport-based target size
        const textureWidth = originalSprite.texture.width
        const textureHeight = originalSprite.texture.height
        if (textureWidth > 0 && textureHeight > 0) {
          const scaleX = targetTileSize / textureWidth
          const scaleY = targetTileSize / textureHeight
          initialScale = Math.min(scaleX, scaleY)
        }
        tile.scale.set(initialScale)
      }
    }

    // Fallback: golden rectangle sized to viewport
    if (!tile!) {
      const gfx = new Graphics()
      const fallbackSize = targetTileSize
      gfx.roundRect(-fallbackSize / 2, -fallbackSize / 2, fallbackSize, fallbackSize, fallbackSize * 0.15)
      gfx.fill({ color: 0xFFD700, alpha: 0.8 })
      gfx.stroke({ color: 0xFFA500, width: 3 })
      tile = gfx
    }

    tile.x = pos.x + pos.width / 2
    tile.y = pos.y + pos.height / 2
    tileContainer.addChild(tile)

    // Animate tile to center with rotation
    const delay = index * 0.05
    const angle = (index / total) * Math.PI * 2

    // Calculate target scale (shrink to half size at center)
    const targetScale = initialScale * 0.5

    gsap.to(tile, {
      x: centerX,
      y: centerY,
      rotation: angle + Math.PI,
      duration: 0.6,
      delay: delay,
      ease: 'power2.in',
      onComplete: () => {
        // Spawn sparks as tile arrives
        particles.spawnParticles({
          count: PARTICLE_COUNTS.convergeSparks,
          x: centerX,
          y: centerY,
          colors: [0xFFD700, 0xFF6B00],
          sizeRange: [3, 8],
          speedRange: [2, 6]
        })
      }
    })

    gsap.to(tile.scale, {
      x: targetScale,
      y: targetScale,
      duration: 0.6,
      delay: delay,
      ease: 'power2.in'
    })

    gsap.to(tile, {
      alpha: 0.7,
      duration: 0.3,
      delay: delay + 0.3,
      ease: 'power2.in'
    })

    return tile
  }

  /**
   * Create central explosion when tiles converge
   */
  function createExplosion(centerX: number, centerY: number): void {
    // Clear tiles
    tileContainer.removeChildren()

    // Screen shake
    screenShake.start(12, 400)

    // Multiple shockwaves
    const colors = [0xFFD700, 0xFF6B00, 0xFFFFFF]
    colors.forEach((color, i) => {
      const shockwave = new Graphics()
      shockwave.circle(0, 0, 30)
      shockwave.fill({ color, alpha: 0.8 })
      shockwave.x = centerX
      shockwave.y = centerY
      shockwave.blendMode = BLEND_MODES.ADD as any
      effectsContainer.addChild(shockwave)

      gsap.to(shockwave.scale, {
        x: 8 + i * 2,
        y: 8 + i * 2,
        duration: 0.5 + i * 0.1,
        delay: i * 0.1,
        ease: 'power2.out'
      })
      gsap.to(shockwave, {
        alpha: 0,
        duration: 0.5 + i * 0.1,
        delay: i * 0.1,
        ease: 'power2.out',
        onComplete: () => shockwave.destroy()
      })
    })

    // Massive particle burst
    particles.spawnParticles({
      count: PARTICLE_COUNTS.explosionBurst,
      x: centerX,
      y: centerY,
      colors: [0xFFD700, 0xFF6B00, 0xFFFFFF, 0xFF0000],
      sizeRange: [5, 15],
      speedRange: [8, 18],
      gravity: 0.3,
      maxLife: 2.5,
      shape: 'diamond'
    })
  }

  /**
   * Spawn fireworks
   */
  function spawnFireworks(width: number, height: number): void {
    const sessionId = timers.getSessionId()

    // Fewer positions on mobile
    const positions = IS_MOBILE ? [
      { x: width * 0.3, y: height * 0.3 },
      { x: width * 0.7, y: height * 0.3 },
      { x: width * 0.5, y: height * 0.4 }
    ] : [
      { x: width * 0.25, y: height * 0.3 },
      { x: width * 0.75, y: height * 0.3 },
      { x: width * 0.5, y: height * 0.2 },
      { x: width * 0.3, y: height * 0.5 },
      { x: width * 0.7, y: height * 0.5 }
    ]

    positions.forEach((pos, i) => {
      timers.setTimeout(() => {
        particles.spawnFirework(pos.x, pos.y, PARTICLE_COUNTS.fireworkCount)
      }, 300 + i * 200, sessionId)
    })

    // Additional random fireworks - less frequent on mobile
    timers.setInterval(() => {
      const x = width * 0.2 + Math.random() * width * 0.6
      const y = height * 0.15 + Math.random() * height * 0.4
      particles.spawnFirework(x, y, PARTICLE_COUNTS.fireworkCount)
    }, IS_MOBILE ? 600 : 400, sessionId)
  }

  /**
   * Create win text with dramatic zoom
   * Uses unified sizing - all win images display at same size
   */
  function createWinText(width: number, height: number): void {
    textContainer.removeChildren()

    const centerX = width / 2
    const centerY = height * 0.35

    const result = createWinImageSprite(WIN_CONFIG, width, height)
    if (result) {
      const { sprite } = result

      // Use unified sizing calculation
      const finalScale = calculateWinImageScale(sprite.width, sprite.height, width, height)

      sprite.x = centerX
      sprite.y = centerY
      const initialScale = finalScale * 3
      sprite.scale.set(initialScale)
      sprite.alpha = 0

      textContainer.addChild(sprite)

      // Zoom in with bounce
      gsap.to(sprite, { alpha: 1, duration: 0.2, ease: 'power2.out' })
      gsap.to(sprite.scale, {
        x: finalScale,
        y: finalScale,
        duration: 0.5,
        ease: 'back.out(1.5)'
      })

      // Subtle pulse
      const pulseTargetScale = finalScale * 1.03
      gsap.to(sprite.scale, {
        x: pulseTargetScale,
        y: pulseTargetScale,
        duration: 0.3,
        delay: 0.6,
        yoyo: true,
        repeat: -1,
        ease: 'sine.inOut'
      })
    } else {
      // Fallback text - use unified sizing
      const maxWidth = Math.min(width * WIN_IMAGE_SIZING.MAX_WIDTH_PERCENT, WIN_IMAGE_SIZING.MAX_WIDTH_PX)
      const maxHeight = height * WIN_IMAGE_SIZING.MAX_HEIGHT_PERCENT

      const isMobile = isMobileViewport(width)
      let fontSize = isMobile
        ? Math.min(height * 0.08, width * 0.12)
        : Math.min(height * 0.1, width * 0.14)

      const testText = new Text({
        text: WIN_CONFIG.textFallback,
        style: { fontSize: fontSize, fontFamily: 'Arial Black, sans-serif' }
      })
      if (testText.width > maxWidth) {
        fontSize = fontSize * (maxWidth / testText.width)
      }
      if (testText.height > maxHeight) {
        fontSize = fontSize * (maxHeight / testText.height)
      }
      testText.destroy()

      const text = new Text({
        text: WIN_CONFIG.textFallback,
        style: {
          fontFamily: 'Arial Black, sans-serif',
          fontSize: fontSize,
          fontWeight: 'bold',
          fill: WIN_COLORS.YELLOW,
          stroke: { color: WIN_COLORS.ORANGE, width: 10 },
          dropShadow: { color: WIN_COLORS.RED, blur: 20, distance: 0, alpha: 0.7 },
          align: 'center'
        }
      })
      text.anchor.set(0.5)
      text.x = centerX
      text.y = centerY
      text.scale.set(3)
      text.alpha = 0

      textContainer.addChild(text)

      gsap.to(text, { alpha: 1, duration: 0.2, ease: 'power2.out' })
      gsap.to(text.scale, {
        x: 1,
        y: 1,
        duration: 0.5,
        ease: 'back.out(1.5)'
      })
    }
  }

  /**
   * Create amount display with counting animation
   * Positioned below the win image, always smaller than the image
   */
  function createAmountDisplay(amount: number, width: number, height: number): void {
    const isMobile = isMobileViewport(width)
    const baseScale = getAmountDisplayScale('BIG', isMobile)

    // Create container for amount display
    const amountContainer = new Container()

    // Calculate max dimensions - number is prominent but still smaller than image
    const maxWidth = Math.min(width * WIN_IMAGE_SIZING.MAX_WIDTH_PERCENT, WIN_IMAGE_SIZING.MAX_WIDTH_PX) * 0.85
    const maxHeight = height * 0.2

    // Create initial number display with 0
    let currentAmount = 0
    let numContainer = createNumberDisplay(0, { scale: baseScale })

    // Calculate final scale
    let finalScale = numContainer.scale.x
    let bounds = numContainer.getBounds()

    if (bounds.width > maxWidth) {
      const scaleByWidth = maxWidth / bounds.width
      finalScale = numContainer.scale.x * scaleByWidth
    }

    if (bounds.height > maxHeight) {
      const scaleByHeight = maxHeight / bounds.height
      finalScale = Math.min(finalScale, numContainer.scale.x * scaleByHeight)
    }

    // Helper function to rebuild number display
    const rebuildNumberDisplay = (newAmount: number) => {
      // Properly destroy old children to free resources while preserving textures
      const destroyChild = (obj: any) => {
        if (obj.children && obj.children.length > 0) {
          while (obj.children.length > 0) {
            const c = obj.children[0]
            obj.removeChild(c)
            destroyChild(c)
          }
        }
        if (obj.destroy) obj.destroy({ children: false, texture: false, textureSource: false })
      }
      while (amountContainer.children.length > 0) {
        const child = amountContainer.children[0]
        amountContainer.removeChild(child)
        destroyChild(child)
      }
      numContainer = createNumberDisplay(newAmount, { scale: baseScale })

      // Recalculate scale for new amount
      let newFinalScale = numContainer.scale.x
      bounds = numContainer.getBounds()

      if (bounds.width > maxWidth) {
        const scaleByWidth = maxWidth / bounds.width
        newFinalScale = numContainer.scale.x * scaleByWidth
      }

      if (bounds.height > maxHeight) {
        const scaleByHeight = maxHeight / bounds.height
        newFinalScale = Math.min(newFinalScale, numContainer.scale.x * scaleByHeight)
      }

      numContainer.scale.set(newFinalScale, newFinalScale)

      // Calculate final bounds after scaling for proper centering
      const finalBounds = numContainer.getBounds()
      const finalWidth = finalBounds.width

      // Center the number container within amount container
      numContainer.x = -finalWidth / 2
      numContainer.y = 0

      amountContainer.addChild(numContainer)
      return newFinalScale
    }

    // Initial display
    finalScale = rebuildNumberDisplay(0)

    // Set initial scale for animation (start from 0)
    amountContainer.scale.set(0)

    // Position amount below the image (image is at 35% height)
    const amountY = height * 0.55
    amountContainer.x = width / 2
    amountContainer.y = amountY

    textContainer.addChild(amountContainer)

    // Scale animation
    gsap.to(amountContainer.scale, {
      x: finalScale,
      y: finalScale,
      duration: 0.5,
      delay: 0.2,
      ease: 'elastic.out(1, 0.6)'
    })

    // Counting animation - store tween so it can be killed on cleanup
    const counterDuration = getCounterDuration(amount)
    counterTween = gsap.to({ value: 0 }, {
      value: amount,
      duration: counterDuration,
      delay: 0.2,
      ease: 'power2.out',
      onUpdate: function() {
        // Preserve decimals for amounts like 0.5 - truncate to 1 decimal place
        const rawValue = this.targets()[0].value
        const newAmount = amount % 1 !== 0 ? Math.floor(rawValue * 10) / 10 : Math.floor(rawValue)
        if (newAmount !== currentAmount) {
          currentAmount = newAmount
          rebuildNumberDisplay(currentAmount)
        }
      }
    })
  }

  /**
   * Show the big win animation
   */
  function show(
    width: number,
    height: number,
    tilePositions: TilePosition[],
    amount = 0,
    onComplete?: () => void
  ): void {
    const sessionId = timers.newSession()

    canvasW = width
    canvasH = height
    container.visible = true
    container.alpha = 1
    isAnimating = true
    storedOnComplete = onComplete || null

    const centerX = width / 2
    const centerY = height / 2

    // Kill existing animations
    if (animationTimeline) {
      animationTimeline.kill()
    }
    if (counterTween) {
      counterTween.kill()
      counterTween = null
    }
    gsap.killTweensOf(container)

    // Clear previous content
    backgroundContainer.removeChildren()
    tileContainer.removeChildren()
    effectsContainer.removeChildren()
    particleContainer.removeChildren()
    textContainer.removeChildren()
    skipButtonContainer.removeChildren()
    particles.clear()

    // Create animation timeline
    animationTimeline = gsap.timeline({
      onComplete: () => {
        // Restore tiles
        if (reels && reels.tilesContainer) {
          reels.tilesContainer.visible = true
        }

        gsap.to(container, {
          alpha: 0,
          duration: 0.6,
          ease: 'power2.in',
          onComplete: () => {
            const callback = storedOnComplete
            storedOnComplete = null
            hide()
            if (callback) callback()
          }
        })
      }
    })

    // Step 0: Dark overlay
    createDarkOverlay(backgroundContainer, width, height, 0x1a0a1a, 0.75, true)

    // Create skip button
    createSkipButton(width, height)

    // Step 1: Tiles converge to center
    if (tilePositions && tilePositions.length > 0) {
      // Hide actual tiles
      if (reels && reels.tilesContainer) {
        reels.tilesContainer.visible = false
      }

      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return
        tilePositions.forEach((pos, i) => {
          createConvergingTile(pos, i, tilePositions.length, centerX, centerY, width, height)
        })
      }, null, 0.1)

      // Step 2: Explosion when tiles arrive
      const convergeTime = 0.1 + 0.6 + tilePositions.length * 0.05
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return
        createExplosion(centerX, centerY)
      }, null, convergeTime)

      // Step 3: Win text appears
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return
        createWinText(width, height)
      }, null, convergeTime + 0.3)

      // Step 4: Fireworks
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return
        spawnFireworks(width, height)
      }, null, convergeTime + 0.5)

      // Step 5: Amount
      if (amount > 0) {
        animationTimeline.call(() => {
          if (sessionId !== timers.getSessionId()) return
          createAmountDisplay(amount, width, height)
        }, null, convergeTime + 0.6)
      }
    } else {
      // No tiles - just show text and effects
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return
        screenShake.start(10, 300)
        particles.spawnShockwave(centerX, centerY, 0xFFD700, 600)
        createWinText(width, height)
        spawnFireworks(width, height)
      }, null, 0.2)

      if (amount > 0) {
        animationTimeline.call(() => {
          if (sessionId !== timers.getSessionId()) return
          createAmountDisplay(amount, width, height)
        }, null, 0.5)
      }
    }

    // Step 6: Hold
    animationTimeline.to({}, { duration: HOLD_DURATION })
  }

  /**
   * Hide the animation
   */
  function hide(): void {
    container.visible = false
    isAnimating = false
    storedOnComplete = null

    // Restore tiles
    if (reels && reels.tilesContainer) {
      reels.tilesContainer.visible = true
    }

    timers.clearAll()
    screenShake.stop()
    particles.clear()

    if (animationTimeline) {
      animationTimeline.kill()
      animationTimeline = null
    }

    if (counterTween) {
      counterTween.kill()
      counterTween = null
    }

    // Properly destroy all children and kill GSAP tweens to prevent memory leaks
    clearContainer(backgroundContainer)
    clearContainer(tileContainer)
    clearContainer(effectsContainer)
    clearContainer(particleContainer)
    clearContainer(textContainer)
    clearContainer(skipButtonContainer)
  }

  /**
   * Update animation
   */
  function update(_deltaTime = 1): void {
    if (!isAnimating || !container.visible) return
    screenShake.update()
    particles.update(canvasH)
  }

  /**
   * Build/rebuild for canvas resize
   */
  function build(width: number, height: number): void {
    canvasW = width
    canvasH = height
  }

  return {
    container,
    show,
    hide,
    update,
    build,
    isShowing: () => isAnimating
  }
}
