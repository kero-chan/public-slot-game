// @ts-nocheck
import { Container, Sprite, Rectangle } from 'pixi.js'
import gsap from 'gsap'
import { useAudioEffects } from '@/composables/useAudioEffects'
import { audioManager } from '@/composables/audioManager'
import {
  createDarkOverlay,
  createNumberDisplay,
  type BaseOverlay
} from './base'
import { getGlyphSprite, getWinAnnouncementSprite, getBackgroundSprite } from '@/config/spritesheet'
import { BONUS_OVERLAY_LAYOUT, WIN_ANIMATION_LAYOUT } from './winAnimationConstants'

/**
 * Bonus overlay interface
 */
export interface BonusOverlay extends BaseOverlay {
  show: (freeSpinsCount: number, cWidth: number, cHeight: number, onStart: () => void) => void
}

/**
 * Simplified constants for bonus overlay
 */
const BUTTON_SCALE_MULTIPLIER = 2.5 // Increase button size significantly
const PULSE_AMOUNT = 0.02
const PULSE_START_TIME = 0.8
const BACKGROUND_PULSE_SPEED = 1.2
const CONTENT_PULSE_SPEED = 2.5
// Lower the panel by this percentage (only on bonus overlay, not during free spin session)
const PANEL_Y_OFFSET_PERCENT = 0.18

/**
 * Creates a bonus trigger overlay that displays when 3+ bonus tiles appear
 * Shows congratulatory message and number of free spins awarded
 */
export function createBonusOverlay(): BonusOverlay {
  const { playEffect } = useAudioEffects()
  const container = new Container()
  container.visible = false
  container.zIndex = 1100

  // Sub-containers for layering
  const backgroundContainer = new Container()
  const panelContainer = new Container()
  const contentContainer = new Container()

  container.addChild(backgroundContainer)
  container.addChild(panelContainer)
  container.addChild(contentContainer)

  // State
  let isAnimating = false
  let onStartCallback: (() => void) | null = null
  let canvasWidth = 600
  let canvasHeight = 800
  let numberContainer: Container | null = null
  let bgImage: Sprite | null = null
  let panelBgSprite: Sprite | null = null
  let baseContentScale = 1.0
  let currentHorizontalContainer: Container | null = null

  /**
   * Create background with dark overlay
   */
  function createBackground(width: number, height: number): void {
    backgroundContainer.removeChildren()
    createDarkOverlay(backgroundContainer, width, height, 0x000000, 0.6, false)

    try {
      bgImage = getWinAnnouncementSprite('free_spins_overlay_bg.webp')
      if (bgImage) {
        bgImage.x = width / 2
        bgImage.y = height / 2
        const scaleX = width / bgImage.width
        const scaleY = height / bgImage.height
        const scale = Math.max(scaleX, scaleY)
        bgImage.scale.set(scale)
        backgroundContainer.addChild(bgImage)
      }
    } catch (error) {
      console.warn('Failed to load bonus overlay background:', error)
    }
  }

  /**
   * Calculate panel dimensions based on background image, ensuring it fits within game view
   */
  function calculatePanelDimensions(
    cWidth: number,
    cHeight: number
  ): { width: number; height: number; scale: number } {
    const bgSprite = getBackgroundSprite('free_spin_background.webp')
    if (!bgSprite) {
      // Fallback if image not found
      return { width: 300, height: 150, scale: 1 }
    }

    // Get max dimensions from constants
    const maxWidth = Math.min(
      cWidth * BONUS_OVERLAY_LAYOUT.PANEL_MAX_WIDTH_PERCENT,
      BONUS_OVERLAY_LAYOUT.PANEL_MAX_WIDTH_PX,
      cWidth * WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT
    )
    const maxHeight = cHeight * BONUS_OVERLAY_LAYOUT.PANEL_HEIGHT_PERCENT

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
   * Create content container with number positioned at 60% of background height
   * Number size is 10% of background height
   */
  function createNumberContent(
    freeSpinsCount: number,
    backgroundHeight: number,
    cWidth: number
  ): Container {
    const container = new Container()

    // Create number display
    numberContainer = createNumberDisplay(freeSpinsCount, { scale: 1.0 })
    const numberBounds = numberContainer.getBounds()

    // Calculate scale: number height should be 55% of background height
    const targetNumberHeight = backgroundHeight * 0.55
    const scale = targetNumberHeight / numberBounds.height
    
    // Ensure doesn't exceed game view max width
    const maxContentWidth = cWidth * WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT
    if (numberBounds.width * scale > maxContentWidth) {
      const widthScale = maxContentWidth / numberBounds.width
      baseContentScale = widthScale
    } else {
      baseContentScale = scale
    }

    // Apply scale to number
    numberContainer.scale.set(baseContentScale)
    const scaledWidth = numberBounds.width * baseContentScale
    const scaledHeight = numberBounds.height * baseContentScale
    numberContainer.pivot.set(scaledWidth / 2, scaledHeight / 2)
    numberContainer.x = -40 // Shift left slightly
    numberContainer.y = 0 // Position will be handled by container
    container.addChild(numberContainer)

    return container
  }

  /**
   * Create panel with background image
   */
  function createPanelBackground(
    centerX: number,
    centerY: number,
    width: number,
    height: number,
    scale: number
  ): void {
    panelContainer.removeChildren()
    
    const bgSprite = getBackgroundSprite('free_spin_background.webp')
    if (!bgSprite) {
      console.warn('Free spin background sprite not found')
      return
    }

    panelBgSprite = bgSprite
    panelBgSprite.anchor.set(0.5)
    panelBgSprite.scale.set(scale)
    panelBgSprite.x = centerX
    panelBgSprite.y = centerY
    panelContainer.addChild(panelBgSprite)
  }

  /**
   * Create start button with proper bounds checking and increased size
   */
  function createStartButton(cWidth: number, cHeight: number, centerX: number): Container | null {
    const buttonSprite = getGlyphSprite('glyph_start_button.webp')
    if (!buttonSprite) {
      console.warn('Start button sprite not found')
      return null
    }

    // Calculate max button size (increased significantly)
    // Use larger percentage of canvas for button
    const maxButtonWidth = cWidth * BONUS_OVERLAY_LAYOUT.BUTTON_MAX_WIDTH_PERCENT * BUTTON_SCALE_MULTIPLIER
    const maxButtonHeight = cHeight * BONUS_OVERLAY_LAYOUT.BUTTON_MAX_HEIGHT_PERCENT * BUTTON_SCALE_MULTIPLIER

    // Calculate scale based on larger dimensions
    let buttonScale = Math.min(maxButtonWidth / buttonSprite.width, maxButtonHeight / buttonSprite.height)

    // Ensure button doesn't exceed game view bounds (but allow larger than before)
    const maxAllowedWidth = cWidth * WIN_ANIMATION_LAYOUT.MAX_WIDTH_PERCENT * 1.3 // Allow larger button
    if (buttonSprite.width * buttonScale > maxAllowedWidth) {
      buttonScale = maxAllowedWidth / buttonSprite.width
    }

    // Position button at 15% from bottom of screen
    const buttonY = cHeight * 0.85

    // Create button container
    const startButton = new Container()
    startButton.x = centerX
    startButton.y = buttonY

    buttonSprite.anchor.set(0.5)
    buttonSprite.scale.set(buttonScale)
    startButton.addChild(buttonSprite)

    // Setup interactivity
    startButton.eventMode = 'static'
    startButton
    const spriteBounds = buttonSprite.getBounds()
    startButton.hitArea = new Rectangle(
      -spriteBounds.width / 2,
      -spriteBounds.height / 2,
      spriteBounds.width,
      spriteBounds.height
    )

    // Store button scale for hover animation
    const finalButtonScale = buttonScale

    startButton.on('pointerover', () => {
      gsap.to(startButton.scale, { x: finalButtonScale * 1.05, y: finalButtonScale * 1.05, duration: 0.2 })
    })

    startButton.on('pointerout', () => {
      gsap.to(startButton.scale, { x: finalButtonScale, y: finalButtonScale, duration: 0.2 })
    })

    startButton.on('pointerdown', () => {
      audioManager.switchToJackpotMusic()
      if (onStartCallback) {
        onStartCallback()
      }
      hide()
    })

    startButton.scale.set(0)
    startButton.alpha = 0

    return startButton
  }

  /**
   * Setup entrance animations
   */
  function setupAnimations(
    startButton: Container | null
  ): void {
    const tl = gsap.timeline()

    // Panel entrance
    panelContainer.scale.set(0.8)
    panelContainer.alpha = 0
    tl.to(panelContainer, { scale: 1, alpha: 1, duration: 0.4, ease: 'back.out(1.5)' }, 0)

    // Content entrance
    if (currentHorizontalContainer) {
      tl.to(currentHorizontalContainer, { alpha: 1, duration: 0.1 }, 0.2)
      tl.to(
        currentHorizontalContainer.scale,
        {
          x: baseContentScale * 1.05,
          y: baseContentScale * 1.05,
          duration: 0.3,
          ease: 'back.out(2)'
        },
        0.2
      )
      tl.to(
        currentHorizontalContainer.scale,
        { x: baseContentScale, y: baseContentScale, duration: 0.2, ease: 'power2.out' },
        0.5
      )
    }

    // Button entrance
    if (startButton) {
      // Get the final button scale from the sprite inside
      const buttonSprite = startButton.children[0] as Sprite
      const finalButtonScale = buttonSprite ? buttonSprite.scale.x : 1
      tl.to(startButton, { alpha: 1, duration: 0.3 }, 0.7)
      tl.to(startButton.scale, { x: finalButtonScale, y: finalButtonScale, duration: 0.3, ease: 'back.out(1.5)' }, 0.7)
    }
  }

  /**
   * Show the overlay
   */
  function show(freeSpinsCount: number, cWidth: number, cHeight: number, onStart: () => void): void {
    // Validate inputs
    if (!freeSpinsCount || freeSpinsCount <= 0 || isNaN(freeSpinsCount)) {
      console.error('Invalid freeSpinsCount:', freeSpinsCount)
      return
    }
    if (!cWidth || !cHeight || cWidth <= 0 || cHeight <= 0) {
      console.error('Invalid canvas dimensions:', cWidth, cHeight)
      return
    }

    try {
      canvasWidth = cWidth
      canvasHeight = cHeight
      container.visible = true
      isAnimating = true
      onStartCallback = onStart

      // Reset references
      currentHorizontalContainer = null

      // Play sound
      playEffect('reach_bonus')

      // Clear containers
      backgroundContainer.removeChildren()
      panelContainer.removeChildren()
      contentContainer.removeChildren()

      const centerX = cWidth / 2
      // Lower the panel position on this overlay only
      const panelCenterY = cHeight * (BONUS_OVERLAY_LAYOUT.PANEL_CENTER_Y_PERCENT + PANEL_Y_OFFSET_PERCENT)

      // Create background
      createBackground(cWidth, cHeight)

      // Calculate panel dimensions based on background image
      const { width: panelWidth, height: panelHeight, scale: panelScale } = calculatePanelDimensions(
        cWidth,
        cHeight
      )

      // Create panel with background image
      createPanelBackground(centerX, panelCenterY, panelWidth, panelHeight, panelScale)

      // Create number content - size is 10% of background height
      const numberContent = createNumberContent(freeSpinsCount, panelHeight, cWidth)

      // Position number at 60% from top of background image
      // Background is centered at panelCenterY, so 60% from top = panelCenterY - panelHeight/2 + panelHeight*0.6
      const numberY = panelCenterY - panelHeight / 2 + panelHeight * 0.6
      // Move number to the right by 16% of background width
      const numberOffsetX = panelWidth * 0.16
      const numberOffsetY = panelHeight * 0.16
      numberContent.x = centerX + numberOffsetX
      numberContent.y = numberY - numberOffsetY
      numberContent.scale.set(0)
      numberContent.alpha = 0
      contentContainer.addChild(numberContent)

      currentHorizontalContainer = numberContent

      // Create start button
      const startButton = createStartButton(cWidth, cHeight, centerX)
      if (startButton) {
        contentContainer.addChild(startButton)
      }

      // Setup animations
      setupAnimations(startButton)
    } catch (error) {
      console.error('Error showing bonus overlay:', error)
      container.visible = true
      isAnimating = true
    }
  }

  /**
   * Hide the overlay
   */
  function hide(): void {
    gsap.killTweensOf(panelContainer)
    gsap.killTweensOf(numberContainer)
    gsap.killTweensOf(currentHorizontalContainer)

    container.visible = false
    isAnimating = false

    backgroundContainer.removeChildren()
    panelContainer.removeChildren()
    contentContainer.removeChildren()

    bgImage = null
    panelBgSprite = null
    numberContainer = null
    currentHorizontalContainer = null
  }

  /**
   * Update animation
   */
  function update(timestamp: number): void {
    if (!isAnimating || !container.visible) return

    const elapsed = (Date.now() - timestamp) / 1000

    // Background pulse
    if (bgImage && bgImage.texture) {
      const scaleX = canvasWidth / bgImage.texture.width
      const scaleY = canvasHeight / bgImage.texture.height
      const baseScale = Math.max(scaleX, scaleY)
      const pulse = baseScale * (1 + Math.sin(elapsed * BACKGROUND_PULSE_SPEED) * 0.01)
      bgImage.scale.set(pulse)
    }

    // Content pulse after entrance
    if (currentHorizontalContainer && elapsed > PULSE_START_TIME) {
      const pulse = baseContentScale * (1 + Math.sin(elapsed * CONTENT_PULSE_SPEED) * PULSE_AMOUNT)
      currentHorizontalContainer.scale.set(Math.min(pulse, baseContentScale))
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
    isShowing: () => isAnimating || container.visible
  }
}
