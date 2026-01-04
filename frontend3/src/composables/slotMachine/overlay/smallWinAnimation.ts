/**
 * Small Win Animation
 * Simple, quick celebration for small wins
 * - Fast fade-in text with gentle bounce
 * - Subtle gold sparkles
 * - Quick duration (~1.5s)
 */

import { Container, Graphics, Sprite, Texture, Text } from 'pixi.js'
import gsap from 'gsap'
import {
  createTimerManager,
  createParticleSystem,
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
  imageKey: WIN_IMAGE_KEYS.SMALL,
  targetHeightPercent: WIN_TEXT_HEIGHT_PERCENT.SMALL,
  maxWidthPercent: WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT,
  textFallback: WIN_TEXT_FALLBACK.SMALL
}
const HOLD_DURATION = WIN_ANIMATION_DURATION.SMALL

/**
 * Small win animation interface
 */
export interface SmallWinAnimation extends BaseOverlay {
  show: (canvasWidth: number, canvasHeight: number, tilePositions: TilePosition[], amount: number, onComplete?: () => void) => void
}

/**
 * Creates a simple small win animation
 * Quick and subtle - doesn't interrupt gameplay flow
 */
export function createSmallWinAnimation(): SmallWinAnimation {
  const container = new Container()
  container.visible = false
  container.zIndex = 2000

  // Sub-containers
  const backgroundContainer = new Container()
  const particleContainer = new Container()
  const textContainer = new Container()
  const skipButtonContainer = new Container()

  container.addChild(backgroundContainer)
  container.addChild(particleContainer)
  container.addChild(textContainer)
  container.addChild(skipButtonContainer)

  // Managers
  const timers = createTimerManager()
  const particles = createParticleSystem(particleContainer)

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

    // Scale and position at bottom center
    const isMobile = width < 600
    const buttonScale = isMobile ? 0.084 : 0.105
    skipButton.scale.set(buttonScale)
    skipButton.anchor.set(0.5, 1)
    skipButton.x = width / 2
    skipButton.y = height - 20

    // Make interactive
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

    // Hover effect
    ;(skipButton as any).on('pointerover', () => {
      gsap.to(skipButton.scale, { x: buttonScale * 1.1, y: buttonScale * 1.1, duration: 0.15 })
    })
    ;(skipButton as any).on('pointerout', () => {
      gsap.to(skipButton.scale, { x: buttonScale, y: buttonScale, duration: 0.15 })
    })

    skipButtonContainer.addChild(skipButton)

    // Fade in the button
    skipButton.alpha = 0
    gsap.to(skipButton, { alpha: 1, duration: 0.3, delay: 0.5 })
  }

  /**
   * Create win text/image with gentle bounce animation
   * Uses unified sizing - all win images display at same size
   */
  function createWinText(width: number, height: number): void {
    textContainer.removeChildren()

    // Always center in game view - position image in upper portion
    const centerX = width / 2
    const centerY = height * 0.35

    const result = createWinImageSprite(WIN_CONFIG, width, height)
    if (result) {
      const { sprite } = result

      // Use unified sizing calculation
      const finalScale = calculateWinImageScale(sprite.width, sprite.height, width, height)

      sprite.x = centerX
      sprite.y = centerY
      sprite.scale.set(0)
      sprite.alpha = 0

      textContainer.addChild(sprite)

      // Gentle bounce entrance
      gsap.to(sprite, { alpha: 1, duration: 0.3, ease: 'power2.out' })
      gsap.to(sprite.scale, { x: finalScale, y: finalScale, duration: 0.5, ease: 'back.out(1.2)' })
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
          fill: WIN_COLORS.GOLD,
          stroke: { color: WIN_COLORS.DARK_GOLD, width: 6 },
          dropShadow: { color: WIN_COLORS.BLACK, blur: 10, distance: 0, alpha: 0.5 },
          align: 'center'
        }
      })
      text.anchor.set(0.5)
      text.x = centerX
      text.y = centerY
      text.scale.set(0)
      text.alpha = 0

      textContainer.addChild(text)

      gsap.to(text, { alpha: 1, duration: 0.3, ease: 'power2.out' })
      gsap.to(text.scale, { x: 1, y: 1, duration: 0.5, ease: 'back.out(1.2)' })
    }
  }

  /**
   * Create amount display with counting animation
   * Positioned below the win image, always smaller than the image
   */
  function createAmountDisplay(amount: number, width: number, height: number): void {
    const isMobile = isMobileViewport(width)
    const baseScale = getAmountDisplayScale('SMALL', isMobile)

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

    // Set initial scale for animation (70% of final scale)
    amountContainer.scale.set(finalScale * 0.7, finalScale * 0.7)
    amountContainer.alpha = 0

    // Position amount below the image (image is at 35% height)
    const amountY = height * 0.55
    amountContainer.x = width / 2
    amountContainer.y = amountY

    textContainer.addChild(amountContainer)

    // Scale and alpha animation
    gsap.to(amountContainer, { alpha: 1, duration: 0.3, delay: 0.2, ease: 'power2.out' })
    gsap.to(amountContainer.scale, { x: finalScale, y: finalScale, duration: 0.4, delay: 0.2, ease: 'back.out(1.2)' })

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
   * Spawn subtle gold sparkles
   */
  function spawnSparkles(centerX: number, centerY: number): void {
    // Initial burst of sparkles around the text
    particles.spawnParticles({
      count: 20,
      x: centerX,
      y: centerY,
      colors: [0xFFD700, 0xFFC800, 0xFFE066],
      sizeRange: [3, 8],
      speedRange: [2, 5],
      gravity: 0.1,
      maxLife: 1.5
    })

    // Gentle continuous sparkles
    const sessionId = timers.getSessionId()
    timers.setInterval(() => {
      const angle = Math.random() * Math.PI * 2
      const radius = 80 + Math.random() * 60
      particles.spawnParticles({
        count: 3,
        x: centerX + Math.cos(angle) * radius,
        y: centerY + Math.sin(angle) * radius,
        colors: [0xFFD700],
        sizeRange: [2, 5],
        speedRange: [1, 3],
        gravity: 0.05,
        maxLife: 1
      })
    }, 150, sessionId)
  }

  /**
   * Show the small win animation
   */
  function show(
    width: number,
    height: number,
    _tilePositions: TilePosition[],
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
    particleContainer.removeChildren()
    textContainer.removeChildren()
    skipButtonContainer.removeChildren()
    particles.clear()

    // Create animation timeline
    animationTimeline = gsap.timeline({
      onComplete: () => {
        gsap.to(container, {
          alpha: 0,
          duration: 0.4,
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

    // Step 0: Light semi-transparent overlay
    createDarkOverlay(backgroundContainer, width, height, 0x000000, 0.4, true)

    // Create skip button
    createSkipButton(width, height)

    // Step 1: Win text appears
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return
      createWinText(width, height)
      spawnSparkles(centerX, centerY - 50)
    }, null, 0.1)

    // Step 2: Amount display
    if (amount > 0) {
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return
        createAmountDisplay(amount, width, height)
      }, null, 0.3)
    }

    // Step 3: Hold
    animationTimeline.to({}, { duration: HOLD_DURATION })
  }

  /**
   * Hide the animation
   */
  function hide(): void {
    container.visible = false
    isAnimating = false
    storedOnComplete = null

    timers.clearAll()
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
    clearContainer(particleContainer)
    clearContainer(textContainer)
    clearContainer(skipButtonContainer)
  }

  /**
   * Update animation
   */
  function update(_deltaTime = 1): void {
    if (!isAnimating || !container.visible) return
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
