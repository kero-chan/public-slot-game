// @ts-nocheck
/**
 * Jackpot Trigger Animation
 *
 * THE MOST EXCITING effect in the game!
 * Plays when 3+ bonus tiles trigger free spins.
 *
 * Uses the same epic animation style as the old massiveWinAnimation:
 * - Dark overlay with frame glow
 * - Intense particle effects (shockwaves, fireworks)
 * - Tiles orbit around center then spiral inward
 * - Central explosion when tiles converge
 * - JACKPOT text reveal
 */

import { Container, Graphics, Sprite, Text, Texture } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import gsap from 'gsap'
import { useAudioEffects } from '@/composables/useAudioEffects'
import { audioManager } from '@/composables/audioManager'
import {
  createTimerManager,
  createParticleSystem,
  createScreenShake,
  createDarkOverlay,
  clearContainer,
  type BaseOverlay
} from './base'
import { createWinImageSprite } from './winImageUtils'
import {
  WIN_ANIMATION_LAYOUT,
  WIN_TEXT_HEIGHT_PERCENT,
  WIN_ANIMATION_DURATION,
  WIN_IMAGE_KEYS,
  WIN_TEXT_FALLBACK,
  WIN_COLORS,
  isMobileViewport
} from './winAnimationConstants'

interface AnimatedSprite extends Sprite {
  _baseScaleX?: number
  _baseScaleY?: number
}

export interface BonusTilePosition {
  x: number
  y: number
  width?: number
  height?: number
  col?: number
  row?: number
}

export interface JackpotTriggerAnimation extends BaseOverlay {
  show: (canvasWidth: number, canvasHeight: number, bonusTilePositions: BonusTilePosition[], onComplete: () => void) => void
}

const HOLD_DURATION = WIN_ANIMATION_DURATION.JACKPOT
const FIXED_ORBIT_TILE_COUNT = 8 // Fixed number of tiles for consistent animation

// Mobile detection for performance optimization
const IS_MOBILE = typeof window !== 'undefined' && isMobileViewport(window.innerWidth)

// Particle counts - significantly reduced on mobile for better performance
const PARTICLE_COUNTS = {
  explosionBurst: IS_MOBILE ? 40 : 300,
  fireworkInterval: IS_MOBILE ? 600 : 300,
  fireworkParticles: IS_MOBILE ? 15 : 40,
  energyParticles: IS_MOBILE ? 15 : 100,
  orbitParticles: IS_MOBILE ? 10 : 40,
  celebrationParticles: IS_MOBILE ? 15 : 50
}

/**
 * Jackpot win configuration - responsive sizing based on canvas dimensions
 */
const JACKPOT_CONFIG = {
  imageKey: WIN_IMAGE_KEYS.JACKPOT,
  targetHeightPercent: WIN_TEXT_HEIGHT_PERCENT.JACKPOT,
  maxWidthPercent: WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT,
  textFallback: WIN_TEXT_FALLBACK.JACKPOT
}

export function createJackpotTriggerAnimation(reelsRef?: any, footerRef?: any): JackpotTriggerAnimation {
  let reels = reelsRef
  let footer = footerRef
  const { playEffect, stopEffect } = useAudioEffects()

  const container = new Container()
  container.visible = false
  container.zIndex = 1200

  // Sub-containers (same structure as massiveWinAnimation)
  const backgroundOverlayContainer = new Container()
  const backgroundEffectsContainer = new Container()
  const frameGlowContainer = new Container()
  const tileCollapseContainer = new Container()
  const crystalContainer = new Container()
  const textContainer = new Container()

  container.addChild(backgroundOverlayContainer)
  container.addChild(backgroundEffectsContainer)
  container.addChild(frameGlowContainer)
  container.addChild(tileCollapseContainer)
  container.addChild(crystalContainer)
  container.addChild(textContainer)

  // Managers
  const timers = createTimerManager()
  const crystalParticles = createParticleSystem(crystalContainer)
  const energyParticles = createParticleSystem(backgroundEffectsContainer)
  const screenShake = createScreenShake(container)

  // State
  let isAnimating = false
  let animationTimeline: gsap.core.Timeline | null = null
  let onCompleteCallback: (() => void) | null = null
  let canvasW = 0
  let canvasH = 0
  const tileSprites: (AnimatedSprite | Graphics)[] = []

  /**
   * Create frame glow effect
   */
  function createFrameGlow(width: number, height: number): void {
    frameGlowContainer.removeChildren()

    const layers = 5
    const baseThickness = Math.max(10, width * 0.015)
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

    // Fallback: golden bonus tile sized to viewport
    if (!tile!) {
      const graphicsTile = new Graphics() as Graphics & { _baseScaleX?: number; _baseScaleY?: number }
      const fallbackSize = targetTileSize
      const cornerRadius = fallbackSize * 0.15
      graphicsTile.roundRect(-fallbackSize / 2, -fallbackSize / 2, fallbackSize, fallbackSize, cornerRadius)
      graphicsTile.fill({ color: 0xFFD700, alpha: 0.9 })
      graphicsTile.roundRect(-fallbackSize / 2, -fallbackSize / 2, fallbackSize, fallbackSize, cornerRadius)
      graphicsTile.stroke({ color: 0xFFAA00, width: 3, alpha: 0.8 })
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
        const progress = this.progress()
        const currentAngle = angle + (1.5 * Math.PI * 2) + (progress * Math.PI * 4)
        const currentRadius = orbitRadius * (1 - progress)
        tile.x = centerX + Math.cos(currentAngle) * currentRadius
        tile.y = centerY + Math.sin(currentAngle) * currentRadius
        tile.rotation += 0.3
        tile.alpha = 1 - progress * 0.5
      }
    })
    tl.to(tile.scale, { x: baseScaleX * 0.2, y: baseScaleY * 0.2, duration: 0.8, ease: 'power2.in' }, '<')

    // Phase 5: Disappear
    tl.to(tile, { alpha: 0, duration: 0.1 })

    return tile
  }

  /**
   * Create JACKPOT text/image with bounds checking
   * Sized to be as large as possible while fitting the viewport
   */
  function createJackpotText(width: number, height: number): void {
    textContainer.removeChildren()

    const centerX = width / 2
    const centerY = height / 2
    // Use generous sizing for jackpot - fill up to 95% width and 50% height
    const maxWidth = width * 0.95
    const maxHeight = height * 0.50

    const result = createWinImageSprite(JACKPOT_CONFIG, width, height)

    if (result) {
      const { sprite } = result

      // Calculate scale to fit within bounds while being as large as possible
      const scaleByWidth = maxWidth / sprite.width
      const scaleByHeight = maxHeight / sprite.height
      let finalScale = Math.min(scaleByWidth, scaleByHeight)

      // Account for rotation animation (max rotation is 0.1 radians)
      const rotationAngle = 0.1
      const cosAngle = Math.cos(rotationAngle)
      const sinAngle = Math.sin(rotationAngle)
      const rotatedWidth = Math.abs(sprite.width * finalScale * cosAngle) + Math.abs(sprite.height * finalScale * sinAngle)
      const rotatedHeight = Math.abs(sprite.width * finalScale * sinAngle) + Math.abs(sprite.height * finalScale * cosAngle)

      if (rotatedWidth > maxWidth || rotatedHeight > maxHeight) {
        const scaleByRotatedWidth = maxWidth / rotatedWidth
        const scaleByRotatedHeight = maxHeight / rotatedHeight
        finalScale = Math.min(finalScale, scaleByRotatedWidth, scaleByRotatedHeight)
      }

      sprite.x = centerX
      sprite.y = centerY
      sprite.scale.set(0)

      textContainer.addChild(sprite)

      gsap.to(sprite.scale, { x: finalScale, y: finalScale, duration: 0.8, ease: 'elastic.out(1, 0.5)' })
      gsap.to(sprite, { rotation: 0.1, duration: 0.3, yoyo: true, repeat: 3, ease: 'sine.inOut' })
    } else {
      // Fallback text with bounds checking
      const isMobile = isMobileViewport(width)
      let fontSize = isMobile 
        ? Math.min(height * 0.1, width * 0.14)
        : Math.min(height * 0.12, width * 0.17)

      // Test text to calculate proper size
      const testText = new Text({
        text: JACKPOT_CONFIG.textFallback,
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
        text: JACKPOT_CONFIG.textFallback,
        style: {
          fontFamily: 'Arial Black, sans-serif',
          fontSize: fontSize,
          fontWeight: 'bold',
          fill: WIN_COLORS.YELLOW,
          stroke: { color: WIN_COLORS.ORANGE, width: Math.max(4, width * 0.008) },
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
   * Spawn jackpot-tier effects (most intense, reduced on mobile)
   */
  function spawnJackpotEffects(centerX: number, centerY: number, width: number): void {
    const sessionId = timers.getSessionId()

    // Strong screen shake (reduced on mobile)
    screenShake.start(IS_MOBILE ? 12 : 20, IS_MOBILE ? 500 : 800)

    // Shockwave burst (fewer on mobile)
    const shockwaveScale = Math.min(width, canvasH) * (IS_MOBILE ? 0.5 : 0.8)
    crystalParticles.spawnShockwave(centerX, centerY, 0xFFD700, shockwaveScale)
    if (!IS_MOBILE) {
      timers.setTimeout(() => crystalParticles.spawnShockwave(centerX, centerY, 0xFF00FF, shockwaveScale * 0.85), 100, sessionId)
      timers.setTimeout(() => crystalParticles.spawnShockwave(centerX, centerY, 0x00FFFF, shockwaveScale * 0.7), 200, sessionId)
    }

    // Particle explosion (reduced on mobile)
    crystalParticles.spawnParticles({
      count: PARTICLE_COUNTS.explosionBurst,
      x: centerX,
      y: centerY,
      colors: [0xFFD700, 0xFF00FF, 0x00FFFF],
      speedRange: [8, 15]
    })

    // Continuous fireworks (slower interval on mobile)
    timers.setInterval(() => {
      const x = canvasW * 0.2 + Math.random() * canvasW * 0.6
      const y = canvasH * 0.2 + Math.random() * canvasH * 0.4
      crystalParticles.spawnFirework(x, y, PARTICLE_COUNTS.fireworkParticles)
    }, PARTICLE_COUNTS.fireworkInterval, sessionId)

    // Extra staggered fireworks (fewer on mobile)
    const staggeredCount = IS_MOBILE ? 2 : 5
    for (let i = 0; i < staggeredCount; i++) {
      timers.setTimeout(() => {
        const x = canvasW * 0.2 + Math.random() * canvasW * 0.6
        const y = canvasH * 0.2 + Math.random() * canvasH * 0.4
        crystalParticles.spawnFirework(x, y, PARTICLE_COUNTS.fireworkParticles)
      }, 500 + i * 200, sessionId)
    }
  }

  /**
   * Show the animation
   */
  function show(
    width: number,
    height: number,
    bonusTilePositions: BonusTilePosition[],
    onComplete: () => void
  ): void {
    const sessionId = timers.newSession()

    canvasW = width
    canvasH = height
    container.visible = true
    container.alpha = 1
    isAnimating = true
    onCompleteCallback = onComplete

    const centerX = width / 2
    const centerY = height / 2

    // Hide spin button during animation
    if (footer && footer.setSpinButtonVisible) {
      footer.setSpinButtonVisible(false)
    }

    // Audio
    audioManager.pause()
    playEffect('jackpot_start')

    // Delay reach_bonus slightly to avoid conflict
    timers.setTimeout(() => {
      playEffect('reach_bonus')
    }, 100, sessionId)

    // Kill existing timeline
    if (animationTimeline) {
      animationTimeline.kill()
    }
    gsap.killTweensOf(container)

    // Clear previous content
    backgroundOverlayContainer.removeChildren()
    backgroundEffectsContainer.removeChildren()
    frameGlowContainer.removeChildren()
    tileCollapseContainer.removeChildren()
    crystalContainer.removeChildren()
    textContainer.removeChildren()
    crystalParticles.clear()
    energyParticles.clear()
    tileSprites.length = 0

    // Create animation timeline
    animationTimeline = gsap.timeline({
      onComplete: () => {
        gsap.to(container, {
          alpha: 0,
          duration: 0.8,
          ease: 'power2.in',
          onComplete: () => {
            hide()
          }
        })
      }
    })

    // Step 0: Background overlay
    createDarkOverlay(backgroundOverlayContainer, width, height, 0x1a0a2a, 0.85, true)

    // Step 1: Frame glow
    createFrameGlow(width, height)

    // Step 1.5: Jackpot-tier effects
    spawnJackpotEffects(centerX, centerY, width)

    // Step 2: Energy particles from edges (reduced on mobile)
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return

      const energyCount = PARTICLE_COUNTS.energyParticles
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

    // Step 3: Tiles orbit and spiral (fixed 8 tiles for consistent animation)
    if (bonusTilePositions && bonusTilePositions.length > 0) {
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return

        // Hide actual tiles
        if (reels && reels.tilesContainer) {
          reels.tilesContainer.visible = false
        }

        // Create fixed number of orbiting tiles
        for (let i = 0; i < FIXED_ORBIT_TILE_COUNT; i++) {
          // Cycle through actual bonus tile positions for texture/position reference
          const pos = bonusTilePositions[i % bonusTilePositions.length]
          createOrbitingTile(
            pos.x, pos.y,
            pos.width || 100, pos.height || 100,
            i, FIXED_ORBIT_TILE_COUNT,
            centerX, centerY,
            pos.col || 0, pos.row || 0
          )
        }
      }, null, 0.5)

      // Step 3.5: Central explosion when tiles converge (reduced on mobile)
      animationTimeline.call(() => {
        if (sessionId !== timers.getSessionId()) return

        crystalParticles.spawnParticles({
          count: IS_MOBILE ? 25 : 150,
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
      }, null, 1.8)
    }

    // Step 4: JACKPOT text - show much earlier
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return
      createJackpotText(width, height)

      // Extra crystals around text (reduced on mobile)
      const crystalCount = PARTICLE_COUNTS.celebrationParticles
      for (let i = 0; i < crystalCount; i++) {
        const angle = (i / crystalCount) * Math.PI * 2
        const radius = Math.min(width, height) * 0.15 + Math.random() * Math.min(width, height) * 0.07
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
    }, null, 1.0)

    // Step 5: Hold
    animationTimeline.to({}, { duration: HOLD_DURATION })
  }

  /**
   * Hide the animation
   */
  function hide(): void {
    container.visible = false
    isAnimating = false

    // Stop jackpot start audio
    stopEffect('jackpot_start')

    // Restore tiles
    if (reels && reels.tilesContainer) {
      reels.tilesContainer.visible = true
    }

    // Restore spin button visibility
    if (footer && footer.setSpinButtonVisible) {
      footer.setSpinButtonVisible(true)
    }

    // Cleanup
    timers.clearAll()
    screenShake.stop()
    crystalParticles.clear()
    energyParticles.clear()

    if (animationTimeline) {
      animationTimeline.kill()
      animationTimeline = null
    }

    // Properly destroy all children and kill GSAP tweens to prevent memory leaks
    clearContainer(backgroundOverlayContainer)
    clearContainer(backgroundEffectsContainer)
    clearContainer(frameGlowContainer)
    clearContainer(tileCollapseContainer)
    clearContainer(crystalContainer)
    clearContainer(textContainer)

    tileSprites.length = 0

    container.removeAllListeners()

    if (onCompleteCallback) {
      onCompleteCallback()
      onCompleteCallback = null
    }
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
