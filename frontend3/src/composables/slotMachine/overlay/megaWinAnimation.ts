import { Container, Graphics, Sprite, Texture, Text } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import gsap from 'gsap'
import { useAudioEffects } from '@/composables/useAudioEffects'
import { audioManager } from '@/composables/audioManager'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'
import {
  createTimerManager,
  createParticleSystem,
  createScreenShake,
  createDarkOverlay,
  createNumberDisplay,
  clearContainer,
  type BaseOverlay,
  type Particle
} from './base'
import { createWinImageSprite } from './winImageUtils'
import { getGlyphSprite } from '@/config/spritesheet'
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

// Types
interface AnimatedSprite extends Sprite {
  _baseScaleX?: number
  _baseScaleY?: number
}

export interface TilePosition {
  x: number
  y: number
  width: number
  height: number
  col: number
  row: number
}

/**
 * Win configuration for mega intensity - responsive sizing based on canvas dimensions
 */
const WIN_CONFIG = {
  imageKey: WIN_IMAGE_KEYS.MEGA,
  targetHeightPercent: WIN_TEXT_HEIGHT_PERCENT.MEGA,
  maxWidthPercent: WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT,
  textFallback: WIN_TEXT_FALLBACK.MEGA
}
const HOLD_DURATION = WIN_ANIMATION_DURATION.MEGA

// Mobile detection for performance optimization
const IS_MOBILE = typeof window !== 'undefined' && isMobileViewport(window.innerWidth)

// Particle counts - significantly reduced on mobile for better performance
const PARTICLE_COUNTS = {
  explosionBurst: IS_MOBILE ? 30 : 200,
  fireworkInterval: IS_MOBILE ? 800 : 500,
  energyParticles: IS_MOBILE ? 3 : 8,
  orbitParticles: IS_MOBILE ? 15 : 40,
  celebrationParticles: IS_MOBILE ? 20 : 60
}

/**
 * Mega win animation interface
 */
export interface MegaWinAnimation extends BaseOverlay {
  show: (canvasWidth: number, canvasHeight: number, tilePositions: TilePosition[], amount: number, onComplete?: () => void) => void
}

/**
 * Creates an epic mega win animation system
 * Full cinematic experience: tiles vortex to center, gold transform, explosions, fireworks
 */
export function createMegaWinAnimation(reelsRef?: any): MegaWinAnimation {
  const { playWinningAnnouncement, stopWinningAnnouncement } = useAudioEffects()

  let reels = reelsRef

  const container = new Container()
  container.visible = false
  container.zIndex = 2000

  // Sub-containers
  const backgroundOverlayContainer = new Container()
  const backgroundEffectsContainer = new Container()
  const frameGlowContainer = new Container()
  const tileCollapseContainer = new Container()
  const crystalContainer = new Container()
  const textContainer = new Container()
  const skipButtonContainer = new Container()

  container.addChild(backgroundOverlayContainer)
  container.addChild(backgroundEffectsContainer)
  container.addChild(frameGlowContainer)
  container.addChild(tileCollapseContainer)
  container.addChild(crystalContainer)
  container.addChild(textContainer)
  container.addChild(skipButtonContainer)

  // Managers
  const timers = createTimerManager()
  const crystalParticles = createParticleSystem(crystalContainer)
  const energyParticles = createParticleSystem(backgroundEffectsContainer)
  const screenShake = createScreenShake(container)

  // State
  let isAnimating = false
  let animationTimeline: gsap.core.Timeline | null = null
  let counterTween: gsap.core.Tween | null = null
  const tileSprites: (AnimatedSprite | Graphics)[] = []
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
   * Create frame glow effect
   */
  function createFrameGlow(width: number, height: number): void {
    frameGlowContainer.removeChildren()

    const layers = 5
    const baseThickness = 20
    const colors = [0xffaa00, 0xffdd00, 0xffff00, 0xffffaa, 0xffffff]

    for (let i = 0; i < layers; i++) {
      const glow = new Graphics()
      const thickness = baseThickness * (layers - i) / layers
      const color = colors[i]
      const alpha = (layers - i) / layers * 0.3

      glow.rect(0, 0, width, height)
      glow.stroke({ color, width: thickness, alpha })
      glow.blendMode = BLEND_MODES.ADD as any
      glow.alpha = 0

      frameGlowContainer.addChild(glow)

      gsap.to(glow, { alpha, duration: 0.3, delay: i * 0.05, ease: 'power2.out' })
      gsap.to(glow, {
        alpha: alpha * 0.5,
        duration: 0.5,
        delay: 0.5 + i * 0.05,
        repeat: -1,
        yoyo: true,
        ease: 'sine.inOut'
      })
    }
  }

  /**
   * Create orbiting tile that spirals inward
   * Size is calculated based on viewport dimensions to ensure proper display
   */
  function createOrbitingTile(
    x: number, y: number, width: number, height: number,
    index: number, totalTiles: number,
    centerX: number, centerY: number,
    col: number, row: number
  ): AnimatedSprite | Graphics {
    let tile: AnimatedSprite | Graphics

    // Calculate proper tile size based on viewport
    // Use larger tiles for better visibility during win animations
    const viewportMin = Math.min(canvasW, canvasH)
    const targetTileSize = viewportMin * 0.3 // 30% of smaller dimension

    // Try to clone actual tile sprite
    if (reels && reels.getSpriteCache) {
      const spriteCache = reels.getSpriteCache()
      const visualRow = row - 4
      const cellKey = `${col}:${visualRow}`
      const originalSprite = spriteCache.get(cellKey)

      if (originalSprite && originalSprite.texture) {
        tile = new Sprite(originalSprite.texture) as AnimatedSprite
        tile.anchor.set(0.5, 0.5)
        // Calculate scale to match viewport-based target size
        const textureWidth = originalSprite.texture.width
        const textureHeight = originalSprite.texture.height
        let initialScale = 1
        if (textureWidth > 0 && textureHeight > 0) {
          const scaleX = targetTileSize / textureWidth
          const scaleY = targetTileSize / textureHeight
          initialScale = Math.min(scaleX, scaleY)
        }
        tile.scale.set(initialScale)
        tile._baseScaleX = initialScale
        tile._baseScaleY = initialScale
      }
    }

    // Fallback: white tile sized to viewport
    if (!tile!) {
      const graphicsTile = new Graphics() as Graphics & { _baseScaleX?: number; _baseScaleY?: number }
      const fallbackSize = targetTileSize
      const cornerRadius = fallbackSize * 0.15
      graphicsTile.roundRect(-fallbackSize / 2, -fallbackSize / 2, fallbackSize, fallbackSize, cornerRadius)
      graphicsTile.fill({ color: 0xffffff, alpha: 0.5 })
      graphicsTile.roundRect(-fallbackSize / 2, -fallbackSize / 2, fallbackSize, fallbackSize, cornerRadius)
      graphicsTile.stroke({ color: 0xdddddd, width: 3, alpha: 0.5 })
      graphicsTile._baseScaleX = 1
      graphicsTile._baseScaleY = 1
      tile = graphicsTile as unknown as AnimatedSprite
    }

    tile.x = x
    tile.y = y
    tile.rotation = 0

    tileCollapseContainer.addChild(tile)
    tileSprites.push(tile as AnimatedSprite)

    const angle = (index / totalTiles) * Math.PI * 2
    // Calculate orbit radius based on canvas size to ensure it stays within bounds
    // Use the smaller dimension and leave some padding
    const maxOrbitRadius = Math.min(canvasW, canvasH) * 0.35
    const orbitRadius = Math.min(300, maxOrbitRadius)

    const orbitX = centerX + Math.cos(angle) * orbitRadius
    const orbitY = centerY + Math.sin(angle) * orbitRadius

    const tl = gsap.timeline()
    const baseScaleX = (tile as AnimatedSprite)._baseScaleX || 1
    const baseScaleY = (tile as AnimatedSprite)._baseScaleY || 1

    // Phase 1: Lift
    tl.to(tile.scale, { x: baseScaleX * 1.2, y: baseScaleY * 1.2, duration: 0.3, ease: 'power2.out' })
    tl.to(tile, { y: y - 20, duration: 0.3, ease: 'power2.out' }, '<')

    // Phase 2: Move to orbit
    tl.to(tile, { x: orbitX, y: orbitY, duration: 1.0, ease: 'power2.inOut' })
    tl.to(tile, { rotation: angle + Math.PI * 2, duration: 1.0, ease: 'none' }, '<')

    // Phase 3: Orbit
    tl.to({}, {
      duration: 1.0,
      onUpdate: function() {
        if (!tile || tile.destroyed) return
        const progress = this.progress()
        const currentAngle = angle + (progress * 1.5 * Math.PI * 2)
        tile.x = centerX + Math.cos(currentAngle) * orbitRadius
        tile.y = centerY + Math.sin(currentAngle) * orbitRadius
        tile.rotation = currentAngle + Math.PI * 2 * progress
      }
    })

    // Phase 4: Spiral inward
    tl.to({}, {
      duration: 0.8,
      onUpdate: function() {
        if (!tile || tile.destroyed) return
        const progress = this.progress()
        const currentAngle = angle + (1.5 * Math.PI * 2) + (progress * Math.PI * 4)
        const currentRadius = orbitRadius * (1 - progress)
        tile.x = centerX + Math.cos(currentAngle) * currentRadius
        tile.y = centerY + Math.sin(currentAngle) * currentRadius
        tile.rotation += 0.3
        tile.alpha = 0.5 * (1 - progress)
      }
    })
    tl.to(tile.scale, { x: baseScaleX * 0.2, y: baseScaleY * 0.2, duration: 0.8, ease: 'power2.in' }, '<')

    // Phase 5: Disappear
    tl.to(tile, { alpha: 0, duration: 0.1 })

    return tile
  }

  /**
   * Create win text/image
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

      textContainer.addChild(sprite)

      gsap.to(sprite.scale, { x: finalScale, y: finalScale, duration: 0.8, ease: 'elastic.out(1, 0.5)' })
      gsap.to(sprite, { rotation: 0.1, duration: 0.3, yoyo: true, repeat: 3, ease: 'sine.inOut' })
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
          stroke: { color: WIN_COLORS.ORANGE, width: 12 },
          dropShadow: { color: WIN_COLORS.RED, blur: 30, distance: 0, alpha: 0.8 },
          align: 'center'
        }
      })
      text.anchor.set(0.5)
      text.x = centerX
      text.y = centerY
      text.scale.set(0)

      textContainer.addChild(text)

      gsap.to(text.scale, { x: 1, y: 1, duration: 0.8, ease: 'elastic.out(1, 0.5)' })
    }
  }

  /**
   * Create amount display with counting animation
   * Positioned below the win image, always smaller than the image
   */
  function createAmountDisplay(amount: number, width: number, height: number): void {
    const isMobile = isMobileViewport(width)
    const baseScale = getAmountDisplayScale('MEGA', isMobile)

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
      duration: 0.6,
      delay: 0.3,
      ease: 'back.out(1.5)'
    })

    // Counting animation - store tween so it can be killed on cleanup
    const counterDuration = getCounterDuration(amount)
    counterTween = gsap.to({ value: 0 }, {
      value: amount,
      duration: counterDuration,
      delay: 0.3,
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
   * Spawn mega win effects - intense shockwaves, particles, continuous fireworks
   */
  function spawnMegaEffects(centerX: number, centerY: number): void {
    const sessionId = timers.getSessionId()

    // Strong screen shake (reduced on mobile)
    screenShake.start(IS_MOBILE ? 10 : 15, IS_MOBILE ? 300 : 500)

    // Shockwave burst (fewer on mobile)
    crystalParticles.spawnShockwave(centerX, centerY, 0xFF6B00, IS_MOBILE ? 600 : 1000)
    if (!IS_MOBILE) {
      timers.setTimeout(() => crystalParticles.spawnShockwave(centerX, centerY, 0xFF0000, 1200), 100, sessionId)
      timers.setTimeout(() => crystalParticles.spawnShockwave(centerX, centerY, 0xFFD700, 800), 200, sessionId)
    }

    // Particle explosion (reduced on mobile)
    crystalParticles.spawnParticles({ count: PARTICLE_COUNTS.explosionBurst, x: centerX, y: centerY, colors: [0xFFD700, 0xFF6B00, 0xFF0000], speedRange: [5, 12] })

    // Continuous fireworks throughout animation (slower interval on mobile)
    timers.setInterval(() => {
      const x = canvasW * 0.2 + Math.random() * canvasW * 0.6
      const y = canvasH * 0.2 + Math.random() * canvasH * 0.4
      crystalParticles.spawnFirework(x, y, IS_MOBILE ? 15 : 40)
    }, PARTICLE_COUNTS.fireworkInterval, sessionId)
  }

  /**
   * Show the mega win animation
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

    // Audio
    audioManager.pause()
    playWinningAnnouncement()

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
    backgroundOverlayContainer.removeChildren()
    backgroundEffectsContainer.removeChildren()
    frameGlowContainer.removeChildren()
    tileCollapseContainer.removeChildren()
    crystalContainer.removeChildren()
    textContainer.removeChildren()
    skipButtonContainer.removeChildren()
    crystalParticles.clear()
    energyParticles.clear()
    tileSprites.length = 0

    // Create animation timeline
    animationTimeline = gsap.timeline({
      onComplete: () => {
        // Restore grid visibility
        if (reels && reels.tilesContainer) {
          reels.tilesContainer.visible = true
        }

        gsap.to(container, {
          alpha: 0,
          duration: 0.8,
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

    // Step 0: Background overlay
    createDarkOverlay(backgroundOverlayContainer, width, height, 0x1a0a2a, 0.85, true)

    // Create skip button
    createSkipButton(width, height)

    // Step 1: Frame glow
    createFrameGlow(width, height)

    // Step 1.5: Mega effects
    spawnMegaEffects(centerX, centerY)

    // Step 2: Energy particles (reduced on mobile)
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return

      const energyCount = IS_MOBILE ? 20 : 100
      for (let i = 0; i < energyCount; i++) {
        const side = Math.floor(Math.random() * 4)
        let x = 0, y = 0

        switch (side) {
          case 0: x = Math.random() * width; y = Math.random() * 100; break
          case 1: x = width - Math.random() * 100; y = Math.random() * height; break
          case 2: x = Math.random() * width; y = height - Math.random() * 100; break
          case 3: x = Math.random() * 100; y = Math.random() * height; break
        }

        energyParticles.spawnParticles({
          count: 1,
          x, y,
          colors: [0x00ff88],
          sizeRange: [3, 9],
          speedRange: [0.5, 2],
          gravity: 0,
          maxLife: 3 + Math.random() * 2
        })
      }
    }, null, 0.2)

    // Step 3: Tiles orbit and spiral
    if (tilePositions && tilePositions.length > 0) {
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return

        // Hide actual tiles
        if (reels && reels.tilesContainer) {
          reels.tilesContainer.visible = false
        }

        tilePositions.forEach((pos, index) => {
          createOrbitingTile(
            pos.x, pos.y, pos.width, pos.height,
            index, tilePositions.length,
            centerX, centerY,
            pos.col, pos.row
          )
        })
      }, null, 0.5)

      // Step 3.5: Central explosion when tiles converge (reduced on mobile)
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return

        crystalParticles.spawnParticles({
          count: IS_MOBILE ? 30 : 150,
          x: centerX,
          y: centerY,
          colors: [0x00ffff, 0xFFD700, 0xFF00FF, 0x00FF88],
          sizeRange: [10, 30],
          speedRange: [3, 10],
          shape: 'diamond',
          maxLife: 3
        })

        // Shockwave
        const shockwave = new Graphics()
        shockwave.circle(0, 0, 20)
        shockwave.fill({ color: 0xffff00, alpha: 0.8 })
        shockwave.x = centerX
        shockwave.y = centerY
        shockwave.blendMode = BLEND_MODES.ADD as any
        crystalContainer.addChild(shockwave)

        gsap.to(shockwave.scale, { x: 15, y: 15, duration: 0.6, ease: 'power2.out' })
        gsap.to(shockwave, { alpha: 0, duration: 0.6, ease: 'power2.out', onComplete: () => shockwave.destroy() })
      }, null, 3.6)
    }

    // Step 4: Win text
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return
      createWinText(width, height)

      // Extra crystals around text (reduced on mobile)
      const crystalCount = IS_MOBILE ? 15 : 50
      for (let i = 0; i < crystalCount; i++) {
        const angle = (i / crystalCount) * Math.PI * 2
        const radius = 200 + Math.random() * 100
        crystalParticles.spawnParticles({
          count: 1,
          x: centerX + Math.cos(angle) * radius,
          y: centerY + Math.sin(angle) * radius,
          colors: [0x00ffff, 0xFFD700],
          sizeRange: [10, 20],
          shape: 'diamond',
          maxLife: 3
        })
      }
    }, null, 4.2)

    // Step 5: Amount display
    if (amount > 0) {
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return
        createAmountDisplay(amount, width, height)
      }, null, 4.5)
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

    // Audio
    stopWinningAnnouncement()
    audioManager.resume()

    // Cleanup
    timers.clearAll()
    screenShake.stop()
    crystalParticles.clear()
    energyParticles.clear()

    if (animationTimeline) {
      animationTimeline.kill()
      animationTimeline = null
    }

    if (counterTween) {
      counterTween.kill()
      counterTween = null
    }

    // Properly destroy all children and kill GSAP tweens to prevent memory leaks
    clearContainer(backgroundOverlayContainer)
    clearContainer(backgroundEffectsContainer)
    clearContainer(frameGlowContainer)
    clearContainer(tileCollapseContainer)
    clearContainer(crystalContainer)
    clearContainer(textContainer)
    clearContainer(skipButtonContainer)

    tileSprites.length = 0
  }

  /**
   * Update animation
   */
  function update(_deltaTime = 1): void {
    if (!isAnimating || !container.visible) return

    screenShake.update()
    crystalParticles.update(canvasH)
    energyParticles.update(canvasH)
  }

  /**
   * Build/rebuild for canvas resize
   */
  function build(width: number, height: number): void {
    canvasW = width
    canvasH = height
    if (container.visible) {
      createDarkOverlay(backgroundOverlayContainer, width, height, 0x1a0a2a, 0.85, false)
      createFrameGlow(width, height)
    }
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
