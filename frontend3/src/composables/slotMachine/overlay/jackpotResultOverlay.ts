// @ts-nocheck
/**
 * Jackpot Result Overlay
 * THE MOST EPIC celebration - shown after free spins complete
 *
 * Features:
 * - Explosive initial entrance with multiple shockwaves
 * - Golden frame with pulsing glow
 * - Massive coin fountain from bottom
 * - Rainbow fireworks display
 * - Rotating light beams
 * - Intense screen shake
 * - Epic title with color cycling
 * - Number counter with pulse effects on milestones
 */

import { Container, Graphics, Sprite, Texture, Text } from 'pixi.js'
import { BLEND_MODES } from '@pixi/constants'
import gsap from 'gsap'
import { ASSETS } from '@/config/assets'
import { useGameStore } from '@/stores'
import { useAudioEffects } from '@/composables/useAudioEffects'
import { audioManager } from '@/composables/audioManager'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'
import { getCounterDuration } from '@/utils/gameHelpers'
import winTotalImage from '@/assets/images/japaneseOmakase/winAnnouncements/win_total.webp'
import {
  createTimerManager,
  createParticleSystem,
  createScreenShake,
  createDarkOverlay,
  createNumberDisplay,
  createAssetSprite,
  clearContainer,
  type BaseOverlay
} from './base'
import { getGlyphSprite } from '@/config/spritesheet'

/**
 * Helper to recursively destroy display objects while preserving textures
 */
function destroyPreservingTextures(obj: any): void {
  if (obj.children && obj.children.length > 0) {
    while (obj.children.length > 0) {
      const c = obj.children[0]
      obj.removeChild(c)
      destroyPreservingTextures(c)
    }
  }
  if (obj.destroy) obj.destroy({ children: false, texture: false, textureSource: false })
}

// Constants
const SKIP_DELAY_MS = 2500
const DISPLAY_TIME_AFTER_COUNTER_MS = 3500
const FADE_DURATION_MS = 800
const INTRO_DURATION_S = 1.0

// Mobile detection for performance optimization
const IS_MOBILE = typeof window !== 'undefined' && window.innerWidth < 600

// Particle counts - drastically reduced on mobile for smooth performance
const PARTICLE_COUNTS = {
  // Coin fountain - minimal on mobile
  initialCoinBurst: IS_MOBILE ? 3 : 40,
  continuousCoinCount: IS_MOBILE ? 1 : 15,
  sideCoinCount: IS_MOBILE ? 0 : 8, // Skip side fountains on mobile
  // Fireworks - very few on mobile
  fireworkParticles: IS_MOBILE ? 5 : 40,
  fireworkExtraParticles: IS_MOBILE ? 0 : 50, // Skip extra particles on mobile
  sparkleParticles: IS_MOBILE ? 0 : 30, // Skip sparkles on mobile
  // Explosion entrance - minimal on mobile
  explosionParticles: IS_MOBILE ? 15 : 300,
  // Confetti - skip continuous on mobile
  initialConfetti: IS_MOBILE ? 10 : 150,
  continuousConfetti: IS_MOBILE ? 0 : 50, // Skip continuous confetti on mobile
  // Milestone celebration - minimal
  milestoneParticles: IS_MOBILE ? 5 : 50,
  finalCelebration: IS_MOBILE ? 8 : 100
}

// Timing intervals - much slower on mobile, some effects disabled
const TIMING = {
  coinFountainInterval: IS_MOBILE ? 1500 : 200, // Very slow on mobile
  sideFountainInterval: IS_MOBILE ? 99999 : 350, // Effectively disabled on mobile
  fireworkInterval: IS_MOBILE ? 2000 : 250, // Very slow on mobile
  sparkleInterval: IS_MOBILE ? 99999 : 400, // Disabled on mobile
  confettiInterval: IS_MOBILE ? 99999 : 1000 // Disabled on mobile
}

/**
 * Jackpot result overlay interface
 */
export interface JackpotResultOverlay extends BaseOverlay {
  show: (totalAmount: number, cWidth: number, cHeight: number) => void
}

/**
 * Creates the ultimate jackpot celebration overlay
 */
export function createJackpotResultOverlay(): JackpotResultOverlay {
  const gameStore = useGameStore()
  const { playWinningAnnouncement, stopWinningAnnouncement } = useAudioEffects()

  // Container setup - layered for proper z-ordering
  const container = new Container()
  container.visible = false
  container.zIndex = 1000

  const backgroundContainer = new Container()
  const lightBeamsContainer = new Container()
  const frameContainer = new Container()
  const effectsContainer = new Container()
  const coinContainer = new Container()
  const textContainer = new Container()
  const amountContainer = new Container()
  const frontEffectsContainer = new Container()
  const skipButtonContainer = new Container()

  container.addChild(backgroundContainer)
  container.addChild(lightBeamsContainer)
  container.addChild(frameContainer)
  container.addChild(effectsContainer)
  container.addChild(coinContainer)
  container.addChild(textContainer)
  container.addChild(amountContainer)
  container.addChild(frontEffectsContainer)
  container.addChild(skipButtonContainer)

  // Managers
  const timers = createTimerManager()
  const particles = createParticleSystem(effectsContainer)
  const coinParticles = createParticleSystem(coinContainer)
  const frontParticles = createParticleSystem(frontEffectsContainer)
  const screenShake = createScreenShake(container)

  // Animation state
  let isAnimating = false
  let isFadingOut = false
  let animationStartTime = 0
  let animationTimeline: gsap.core.Timeline | null = null
  let canvasWidth = 0
  let canvasHeight = 0
  let targetAmount = 0
  let currentDisplayAmount = 0
  let lightBeams: Graphics | null = null

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
      if (isAnimating && !isFadingOut) {
        playGenericUISound()
        hide()
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
    gsap.to(skipButton, { alpha: 1, duration: 0.3, delay: 1.0 })
  }

  // UI elements
  let titleSprite: Sprite | null = null
  let titleText: Text | null = null
  let numberContainer: Container | null = null
  let goldenFrame: Graphics | null = null

  /**
   * Helper function to properly center and scale a number container
   * Ensures the number never overflows the viewport
   */
  function centerAndScaleNumberContainer(
    container: Container,
    maxWidth: number,
    targetX: number,
    targetY: number,
    maxHeight?: number
  ): number {
    // Get original bounds before scaling
    const originalBounds = container.getBounds()

    // Calculate scale needed to fit within maxWidth
    let finalScale = 1
    if (originalBounds.width > maxWidth) {
      finalScale = maxWidth / originalBounds.width
    }

    // Also check height constraint if provided
    if (maxHeight && originalBounds.height * finalScale > maxHeight) {
      finalScale = maxHeight / originalBounds.height
    }

    // Apply additional safety margin to ensure we never overflow
    // Cap the scale to prevent extremely large numbers
    const maxScale = Math.min(finalScale, 1.5)
    finalScale = Math.min(finalScale, maxScale)

    // Apply scale
    container.scale.set(finalScale)

    // Center the container properly after scaling
    const bounds = container.getBounds()
    container.x = targetX - bounds.width / 2
    container.y = targetY - bounds.height / 2

    return finalScale
  }

  /**
   * Create epic background with animated gradient
   */
  function createBackground(cWidth: number, cHeight: number): void {
    backgroundContainer.removeChildren()

    // Deep purple/gold gradient background
    createDarkOverlay(backgroundContainer, cWidth, cHeight, 0x0a0515, 0.95, false)

    // Add animated vignette
    const vignette = new Graphics()
    const gradient = vignette
    gradient.circle(cWidth / 2, cHeight / 2, Math.max(cWidth, cHeight) * 0.8)
    gradient.fill({ color: 0x1a0a2a, alpha: 0.5 })
    backgroundContainer.addChild(vignette)
  }

  /**
   * Create rotating light beams
   */
  function createLightBeams(cWidth: number, cHeight: number): void {
    lightBeamsContainer.removeChildren()

    const centerX = cWidth / 2
    const centerY = cHeight / 2
    const beamCount = 16
    const maxRadius = Math.max(cWidth, cHeight)

    lightBeams = new Graphics()

    for (let i = 0; i < beamCount; i++) {
      const angle = (i / beamCount) * Math.PI * 2
      const nextAngle = ((i + 0.15) / beamCount) * Math.PI * 2

      lightBeams.moveTo(centerX, centerY)
      lightBeams.lineTo(
        centerX + Math.cos(angle) * maxRadius,
        centerY + Math.sin(angle) * maxRadius
      )
      lightBeams.lineTo(
        centerX + Math.cos(nextAngle) * maxRadius,
        centerY + Math.sin(nextAngle) * maxRadius
      )
      lightBeams.closePath()
      lightBeams.fill({ color: 0xFFD700, alpha: 0.15 })
    }

    lightBeams.blendMode = BLEND_MODES.ADD as any
    lightBeams.alpha = 0
    lightBeamsContainer.addChild(lightBeams)

    // Fade in and start rotation (skip rotation on mobile for performance)
    gsap.to(lightBeams, { alpha: IS_MOBILE ? 0.5 : 1, duration: 0.5, ease: 'power2.out' })
    if (!IS_MOBILE) {
      gsap.to(lightBeams, { rotation: Math.PI * 2, duration: 20, repeat: -1, ease: 'none' })
    }
  }

  /**
   * Create golden pulsing frame
   */
  function createGoldenFrame(cWidth: number, cHeight: number): void {
    frameContainer.removeChildren()

    const padding = 30
    const cornerRadius = 20
    const frameThickness = 8

    // Outer glow layers
    for (let i = 4; i >= 0; i--) {
      const glow = new Graphics()
      const thickness = frameThickness + i * 6
      const alpha = 0.1 - i * 0.015

      glow.roundRect(padding - thickness/2, padding - thickness/2,
                     cWidth - padding * 2 + thickness, cHeight - padding * 2 + thickness,
                     cornerRadius + i * 2)
      glow.stroke({ color: 0xFFD700, width: thickness, alpha })
      glow.blendMode = BLEND_MODES.ADD as any
      frameContainer.addChild(glow)
    }

    // Main golden frame
    goldenFrame = new Graphics()
    goldenFrame.roundRect(padding, padding, cWidth - padding * 2, cHeight - padding * 2, cornerRadius)
    goldenFrame.stroke({ color: 0xFFD700, width: frameThickness })
    goldenFrame.alpha = 0
    frameContainer.addChild(goldenFrame)

    // Animate frame in
    gsap.to(goldenFrame, { alpha: 1, duration: 0.3, ease: 'power2.out' })

    // Pulse animation
    gsap.to(goldenFrame, {
      alpha: 0.7,
      duration: 0.5,
      yoyo: true,
      repeat: -1,
      ease: 'sine.inOut',
      delay: 0.5
    })
  }

  /**
   * Create epic title with glow effects
   */
  function createTitle(centerX: number, centerY: number, cWidth: number): void {
    textContainer.removeChildren()

    titleSprite = createAssetSprite(winTotalImage)
    if (!titleSprite) {
      console.warn('No title sprite available for jackpot result overlay')
      return
    }
    const maxTitleWidth = cWidth * 0.8 // 80% canvas width

    // Calculate scale to fit 80% canvas width
    const targetScale = maxTitleWidth / titleSprite.width

    // Position at center (anchor will handle centering)
    titleSprite.x = centerX
    titleSprite.y = centerY * 0.4

    // Reset for entrance animation
    titleSprite.scale.set(0)
    textContainer.addChild(titleSprite)

    // Epic entrance - scale up with overshoot
    gsap.to(titleSprite.scale, {
      x: targetScale,
      y: targetScale,
      duration: 0.8,
      ease: 'elastic.out(1.2, 0.4)'
    })
  }

  /**
   * Spawn massive coin fountain from bottom
   */
  function spawnCoinFountain(cWidth: number, cHeight: number): void {
    const coinTextures = ['coin1', 'coin2', 'coin3', 'coin4', 'coin5']
      .map(key => ASSETS.loadedImages?.[key])
      .filter(Boolean)
      .map(img => img instanceof Texture ? img : Texture.from(img as string))

    if (coinTextures.length === 0) return

    const sessionId = timers.getSessionId()

    // Initial burst from center-bottom (fewer bursts on mobile)
    const burstCount = IS_MOBILE ? 3 : 5
    for (let i = 0; i < burstCount; i++) {
      timers.setTimeout(() => {
        coinParticles.spawnCoins({
          x: cWidth / 2 + (Math.random() - 0.5) * 200,
          y: cHeight + 50,
          count: PARTICLE_COUNTS.initialCoinBurst,
          textures: coinTextures,
          spread: 150,
          speed: 18
        })
      }, i * 100, sessionId)
    }

    // Continuous fountain
    timers.setInterval(() => {
      coinParticles.spawnCoins({
        x: cWidth / 2 + (Math.random() - 0.5) * 300,
        y: cHeight + 30,
        count: PARTICLE_COUNTS.continuousCoinCount,
        textures: coinTextures,
        spread: 100,
        speed: 15
      })
    }, TIMING.coinFountainInterval, sessionId)

    // Side fountains
    timers.setInterval(() => {
      // Left fountain
      coinParticles.spawnCoins({
        x: cWidth * 0.15,
        y: cHeight + 30,
        count: PARTICLE_COUNTS.sideCoinCount,
        textures: coinTextures,
        spread: 60,
        speed: 12
      })
      // Right fountain
      coinParticles.spawnCoins({
        x: cWidth * 0.85,
        y: cHeight + 30,
        count: PARTICLE_COUNTS.sideCoinCount,
        textures: coinTextures,
        spread: 60,
        speed: 12
      })
    }, TIMING.sideFountainInterval, sessionId)
  }

  /**
   * Spawn rainbow fireworks display
   */
  function spawnRainbowFireworks(cWidth: number, cHeight: number): void {
    const sessionId = timers.getSessionId()
    const colors = [0xFF0000, 0xFF7F00, 0xFFFF00, 0x00FF00, 0x0000FF, 0x8B00FF, 0xFFD700]

    // Initial fireworks burst (fewer on mobile)
    const initialFireworks = IS_MOBILE ? 4 : 8
    for (let i = 0; i < initialFireworks; i++) {
      timers.setTimeout(() => {
        const x = cWidth * 0.15 + Math.random() * cWidth * 0.7
        const y = cHeight * 0.1 + Math.random() * cHeight * 0.35
        particles.spawnFirework(x, y, PARTICLE_COUNTS.fireworkParticles)

        // Extra colored particles
        frontParticles.spawnParticles({
          count: PARTICLE_COUNTS.fireworkExtraParticles,
          x, y,
          colors: [colors[i % colors.length], 0xFFFFFF],
          sizeRange: [4, 12],
          speedRange: [6, 14],
          gravity: 0.2,
          maxLife: 2
        })
      }, i * 150, sessionId)
    }

    // Continuous fireworks
    timers.setInterval(() => {
      const x = cWidth * 0.1 + Math.random() * cWidth * 0.8
      const y = cHeight * 0.1 + Math.random() * cHeight * 0.4
      particles.spawnFirework(x, y, PARTICLE_COUNTS.fireworkParticles)
    }, TIMING.fireworkInterval, sessionId)

    // Extra sparkle bursts
    timers.setInterval(() => {
      frontParticles.spawnParticles({
        count: PARTICLE_COUNTS.sparkleParticles,
        x: cWidth * 0.2 + Math.random() * cWidth * 0.6,
        y: cHeight * 0.2 + Math.random() * cHeight * 0.3,
        colors: colors,
        sizeRange: [3, 8],
        speedRange: [4, 10],
        gravity: 0.1,
        maxLife: 1.5,
        shape: 'diamond'
      })
    }, TIMING.sparkleInterval, sessionId)
  }

  /**
   * Create explosive shockwave sequence
   */
  function createExplosiveEntrance(centerX: number, centerY: number): void {
    const sessionId = timers.getSessionId()

    // Shockwave burst (fewer on mobile)
    const shockColors = IS_MOBILE
      ? [0xFFD700, 0xFF6B00, 0xFFFFFF]
      : [0xFFD700, 0xFF6B00, 0xFFFFFF, 0xFF00FF, 0x00FFFF]

    shockColors.forEach((color, i) => {
      timers.setTimeout(() => {
        particles.spawnShockwave(centerX, centerY, color, 800 + i * 100)
      }, i * 80, sessionId)
    })

    // Particle explosion
    particles.spawnParticles({
      count: PARTICLE_COUNTS.explosionParticles,
      x: centerX,
      y: centerY,
      colors: [0xFFD700, 0xFF6B00, 0xFFFFFF, 0xFF00FF],
      sizeRange: [5, 20],
      speedRange: [8, 20],
      gravity: 0.15,
      maxLife: 3,
      shape: 'diamond'
    })

    // Screen shake (less intense on mobile)
    screenShake.start(IS_MOBILE ? 15 : 25, IS_MOBILE ? 600 : 1000)
  }

  /**
   * Start fade out animation
   */
  function startFadeOut(): void {
    if (isFadingOut) return
    isFadingOut = true
    animationStartTime = Date.now()

    // Stop audio immediately when starting fade out
    stopWinningAnnouncement()
  }

  /**
   * Show the ultimate jackpot celebration
   */
  function show(totalAmount: number, cWidth: number, cHeight: number): void {
    const sessionId = timers.newSession()

    canvasWidth = cWidth
    canvasHeight = cHeight
    targetAmount = totalAmount
    currentDisplayAmount = 0
    container.visible = true
    container.alpha = 1
    isAnimating = true
    isFadingOut = false
    animationStartTime = Date.now()

    const centerX = cWidth / 2
    const centerY = cHeight / 2

    // Audio
    audioManager.pause()
    playWinningAnnouncement()

    // Clear all containers
    backgroundContainer.removeChildren()
    lightBeamsContainer.removeChildren()
    frameContainer.removeChildren()
    effectsContainer.removeChildren()
    coinContainer.removeChildren()
    textContainer.removeChildren()
    // Properly destroy amountContainer children to preserve textures
    while (amountContainer.children.length > 0) {
      const child = amountContainer.children[0]
      amountContainer.removeChild(child)
      destroyPreservingTextures(child)
    }
    frontEffectsContainer.removeChildren()
    skipButtonContainer.removeChildren()
    particles.clear()
    coinParticles.clear()
    frontParticles.clear()

    // Kill existing timeline
    if (animationTimeline) {
      animationTimeline.kill()
    }

    // Create animation timeline
    animationTimeline = gsap.timeline()

    // Phase 0: Instant background
    createBackground(cWidth, cHeight)

    // Create skip button
    createSkipButton(cWidth, cHeight)

    // Phase 1: EXPLOSIVE ENTRANCE
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return
      createExplosiveEntrance(centerX, centerY)
      createLightBeams(cWidth, cHeight)
      createGoldenFrame(cWidth, cHeight)
    }, null, 0.1)

    // Phase 2: Title appears with shockwave
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return
      createTitle(centerX, centerY, cWidth)
      particles.spawnShockwave(centerX, centerY * 0.4, 0xFFD700, 500)
      screenShake.start(15, 400)
    }, null, 0.5)

    // Phase 3: Amount display (title is already shown in Phase 2)
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return

      const maxWidth = cWidth * 0.85 // 85% canvas width
      const maxHeight = cHeight * 0.2 // 20% canvas height for numbers
      // Use a more conservative base font size that scales with smaller dimension
      const baseFontSize = Math.min(cWidth, cHeight) * 0.2

      numberContainer = createNumberDisplay(0, { baseFontSize })

      // Use helper function to center and scale properly
      const finalScale = centerAndScaleNumberContainer(
        numberContainer,
        maxWidth,
        centerX,
        centerY * 1.15,
        maxHeight
      )

      // Reset to 0 for entrance animation
      numberContainer.scale.set(0)
      numberContainer.alpha = 0
      amountContainer.addChild(numberContainer)

      gsap.to(numberContainer.scale, { x: finalScale, y: finalScale, duration: 0.5, ease: 'back.out(2)' })
      gsap.to(numberContainer, { alpha: 1, duration: 0.3 })
    }, null, 0.8)

    // Phase 4: EPIC continuous effects
    animationTimeline.call(() => {
      if (sessionId !== timers.getSessionId()) return
      spawnCoinFountain(cWidth, cHeight)
      spawnRainbowFireworks(cWidth, cHeight)

      // Continuous confetti
      particles.spawnConfetti(cWidth, cHeight, PARTICLE_COUNTS.initialConfetti)
      timers.setInterval(() => {
        particles.spawnConfetti(cWidth, cHeight, PARTICLE_COUNTS.continuousConfetti)
      }, TIMING.confettiInterval, sessionId)

      // Periodic screen shake (skip on mobile for performance)
      if (!IS_MOBILE) {
        timers.setInterval(() => {
          screenShake.start(8, 200)
        }, 2000, sessionId)
      }
    }, null, 1.2)

    // Auto-close after animation
    const counterDuration = getCounterDuration(totalAmount)
    const totalDisplayTime = (INTRO_DURATION_S + counterDuration) * 1000 + DISPLAY_TIME_AFTER_COUNTER_MS

    timers.setTimeout(() => {
      if (isAnimating && !isFadingOut) {
        startFadeOut()
      }
    }, totalDisplayTime, sessionId)
  }

  /**
   * Hide the overlay
   */
  function hide(): void {
    container.visible = false
    isAnimating = false
    isFadingOut = false

    // Audio
    stopWinningAnnouncement()
    // Don't resume here - let exitFreeSpinMode handle music transition
    // audioManager.resume() is removed to prevent jackpot music from continuing

    // Clear timers
    timers.clearAll()
    screenShake.stop()

    // Kill timeline and animations
    if (animationTimeline) {
      animationTimeline.kill()
      animationTimeline = null
    }
    if (lightBeams) {
      gsap.killTweensOf(lightBeams)
    }
    if (goldenFrame) {
      gsap.killTweensOf(goldenFrame)
    }

    // Clear UI
    titleSprite = null
    titleText = null
    numberContainer = null
    lightBeams = null
    goldenFrame = null

    // Clear particles
    particles.clear()
    coinParticles.clear()
    frontParticles.clear()

    // Properly destroy all children and kill GSAP tweens to prevent memory leaks
    clearContainer(backgroundContainer)
    clearContainer(lightBeamsContainer)
    clearContainer(frameContainer)
    clearContainer(effectsContainer)
    clearContainer(coinContainer)
    clearContainer(textContainer)
    clearContainer(amountContainer)
    clearContainer(frontEffectsContainer)
    clearContainer(skipButtonContainer)

    // Notify game store
    gameStore.completeFinalJackpotResult()
  }

  /**
   * Update loop - handles counter animation and effects
   */
  function update(_timestamp: number): void {
    if (!isAnimating || !container.visible) return

    const elapsed = (Date.now() - animationStartTime) / 1000

    // Handle fade out
    if (isFadingOut) {
      const fadeProgress = Math.min(elapsed / (FADE_DURATION_MS / 1000), 1)
      container.alpha = 1 - fadeProgress

      if (fadeProgress >= 1) {
        hide()
      }
      return
    }

    screenShake.update()
    particles.update(canvasHeight)
    coinParticles.update(canvasHeight)
    frontParticles.update(canvasHeight)

    // Title pulse with color glow effect (skip on mobile for performance)
    if (!IS_MOBILE) {
      if (titleSprite && elapsed >= 1.0) {
        const maxTitleWidth = canvasWidth * 0.8
        const baseScale = maxTitleWidth / (titleSprite.texture.width || 1)
        const pulse = 1 + Math.sin((elapsed - 1.0) * 3) * 0.05
        titleSprite.scale.set(baseScale * pulse)
      }
      if (titleText && elapsed >= 1.0) {
        const maxTitleWidth = canvasWidth * 0.8
        const baseScale = titleText.width > maxTitleWidth ? maxTitleWidth / titleText.width : 1
        const pulse = 1 + Math.sin((elapsed - 1.0) * 3) * 0.05
        titleText.scale.set(baseScale * pulse)
      }
    }

    // Counter animation with milestone effects
    const counterStart = INTRO_DURATION_S
    const counterDuration = getCounterDuration(targetAmount)
    const counterEnd = counterStart + counterDuration

    if (elapsed >= counterStart && elapsed < counterEnd && numberContainer) {
      const counterProgress = (elapsed - counterStart) / counterDuration
      const easeProgress = 1 - Math.pow(1 - counterProgress, 3)
      const newAmount = Math.floor(targetAmount * easeProgress)

      // Check for milestone (every 25%) - skip on mobile for performance
      if (!IS_MOBILE) {
        const oldMilestone = Math.floor(currentDisplayAmount / (targetAmount * 0.25))
        const newMilestone = Math.floor(newAmount / (targetAmount * 0.25))

        if (newMilestone > oldMilestone && newMilestone > 0) {
          // Milestone celebration!
          screenShake.start(10, 200)
          frontParticles.spawnParticles({
            count: PARTICLE_COUNTS.milestoneParticles,
            x: canvasWidth / 2,
            y: canvasHeight * 0.6,
            colors: [0xFFD700, 0xFFFFFF],
            sizeRange: [5, 12],
            speedRange: [5, 12],
            gravity: 0.2
          })
        }
      }

      // On mobile, only update display when value changes significantly (every 5% or at least 1 unit change)
      // This reduces expensive container rebuilds
      const shouldUpdate = IS_MOBILE
        ? (newAmount !== currentDisplayAmount && (
            Math.abs(newAmount - currentDisplayAmount) >= Math.max(1, targetAmount * 0.05) ||
            newAmount === targetAmount
          ))
        : newAmount !== currentDisplayAmount

      if (shouldUpdate) {
        currentDisplayAmount = newAmount

        // Rebuild number display with auto-scaling
        const oldY = numberContainer.y + numberContainer.height / 2 // Get center Y
        const oldAlpha = numberContainer.alpha
        // Properly destroy old children to free resources
        while (amountContainer.children.length > 0) {
          const child = amountContainer.children[0]
          amountContainer.removeChild(child)
          destroyPreservingTextures(child)
        }

        const maxWidth = canvasWidth * 0.85 // 85% canvas width
        const maxHeight = canvasHeight * 0.2 // 20% canvas height
        const baseFontSize = Math.min(canvasWidth, canvasHeight) * 0.2
        numberContainer = createNumberDisplay(currentDisplayAmount, { baseFontSize })

        // Use helper function to center and scale properly
        centerAndScaleNumberContainer(
          numberContainer,
          maxWidth,
          canvasWidth / 2,
          oldY,
          maxHeight
        )

        numberContainer.alpha = oldAlpha
        amountContainer.addChild(numberContainer)
      }
    } else if (elapsed >= counterEnd && numberContainer && currentDisplayAmount !== targetAmount) {
      // Final amount with celebration burst
      currentDisplayAmount = targetAmount
      const oldY = numberContainer.y + numberContainer.height / 2 // Get center Y
      // Properly destroy old children to free resources
      while (amountContainer.children.length > 0) {
        const child = amountContainer.children[0]
        amountContainer.removeChild(child)
        destroyPreservingTextures(child)
      }

      const maxWidth = canvasWidth * 0.85 // 85% canvas width
      const maxHeight = canvasHeight * 0.2 // 20% canvas height
      const baseFontSize = Math.min(canvasWidth, canvasHeight) * 0.2
      numberContainer = createNumberDisplay(currentDisplayAmount, { baseFontSize })

      // Use helper function to center and scale properly
      centerAndScaleNumberContainer(
        numberContainer,
        maxWidth,
        canvasWidth / 2,
        oldY,
        maxHeight
      )

      amountContainer.addChild(numberContainer)

      // Final celebration
      screenShake.start(IS_MOBILE ? 10 : 15, IS_MOBILE ? 300 : 500)
      frontParticles.spawnParticles({
        count: PARTICLE_COUNTS.finalCelebration,
        x: canvasWidth / 2,
        y: canvasHeight * 0.6,
        colors: [0xFFD700, 0xFF6B00, 0xFFFFFF],
        sizeRange: [8, 20],
        speedRange: [8, 16],
        gravity: 0.15
      })
    }
  }

  /**
   * Build/rebuild for canvas resize
   */
  function build(cWidth: number, cHeight: number): void {
    canvasWidth = cWidth
    canvasHeight = cHeight
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
