import { Container, Sprite, Graphics, Text, TextStyle } from 'pixi.js'
import gsap from 'gsap'
import { getWinningHighValueSprite, getGlyphSprite } from '@/config/spritesheet'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'

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
 * Symbols that have win card animations (front/back images)
 */
export const SYMBOLS_WITH_CARDS = ['fa', 'zhong', 'bai', 'bawan', 'liangsuo', 'liangtong', 'wusuo', 'wutong'] as const
export type CardSymbol = typeof SYMBOLS_WITH_CARDS[number]

// Track last shown symbol to skip consecutive same-symbol animations
let lastShownSymbol: string | null = null

// Session-only preference to skip all card animations (resets on page reload)
let skipAllCards = false

// Symbol to image key mapping (front and back)
const SYMBOL_IMAGE_KEYS: Record<string, { front: string; back: string }> = {
  fa: { front: 'won_fa_front.webp', back: 'won_fa_back.webp' },
  zhong: { front: 'won_zhong_front.webp', back: 'won_zhong_back.webp' },
  bai: { front: 'won_bai_front.webp', back: 'won_bai_back.webp' },
  bawan: { front: 'won_bawan_front.webp', back: 'won_bawan_back.webp' },
  liangsuo: { front: 'won_liangsuo_front.webp', back: 'won_liangsuo_back.webp' },
  liangtong: { front: 'won_liangtong_front.webp', back: 'won_liangtong_back.webp' },
  wusuo: { front: 'won_wusuo_front.webp', back: 'won_wusuo_back.webp' },
  wutong: { front: 'won_wutong_front.webp', back: 'won_wutong_back.webp' },
}

/**
 * Grid rect for positioning the image over the reels
 */
export interface GridRect {
  x: number
  y: number
  w: number
  h: number
}

export interface SymbolWinAnimation {
  container: Container
  show: (symbol: string, canvasWidth: number, canvasHeight: number, gridRect?: GridRect) => Promise<void>
  hide: () => void
  update: () => void
  isShowing: () => boolean
  getCurrentSymbol: () => string | null
  setOnHide: (callback: () => void) => void
  resetLastShown: () => void
}

/**
 * Reset the last shown symbol tracker
 * Call this when starting a completely new game session
 */
export function resetSymbolWinAnimationState(): void {
  lastShownSymbol = null
}

/**
 * Check if a symbol has card animation images
 */
export function hasCardImages(symbol: string): boolean {
  return symbol in SYMBOL_IMAGE_KEYS
}

// Keep for backward compatibility
export type HighValueSymbol = CardSymbol
export const HIGH_VALUE_SYMBOLS = SYMBOLS_WITH_CARDS
export function isHighValueSymbol(symbol: string): symbol is CardSymbol {
  return SYMBOLS_WITH_CARDS.includes(symbol as CardSymbol)
}
export function hasSymbolAnimation(symbol: string): boolean {
  return hasCardImages(symbol)
}

/**
 * Creates a symbol win animation effect with 3D flip card
 * Shows front image, user can click to flip to back
 * Skip button to dismiss
 */
export function createSymbolWinAnimation(): SymbolWinAnimation {
  const container = new Container()
  container.visible = false
  container.zIndex = 50

  let isActive = false
  let currentSymbol: string | null = null
  let frontSprite: Sprite | null = null
  let backSprite: Sprite | null = null
  let dimmedOverlay: Graphics | null = null
  let skipButton: Sprite | null = null
  let checkboxContainer: Container | null = null
  let onHideCallback: (() => void) | null = null
  let currentTween: gsap.core.Tween | gsap.core.Timeline | null = null
  let isFlipped = false
  let resolvePromise: (() => void) | null = null
  let cardScale = 1

  /**
   * Create skip button at bottom center
   */
  function createSkipButton(width: number, height: number): void {
    if (skipButton) {
      skipButton.destroy()
      skipButton = null
    }

    skipButton = getGlyphSprite('glyph_skip_button.webp')
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
      if (isActive) {
        // Play generic UI sound
        playGenericUISound()
        hide()
        if (resolvePromise) {
          resolvePromise()
          resolvePromise = null
        }
      }
    })

    const btn = skipButton
    const baseScale = buttonScale
    ;(skipButton as any).on('pointerover', () => {
      gsap.to(btn.scale, { x: baseScale * 1.1, y: baseScale * 1.1, duration: 0.15 })
    })
    ;(skipButton as any).on('pointerout', () => {
      gsap.to(btn.scale, { x: baseScale, y: baseScale, duration: 0.15 })
    })

    container.addChild(skipButton)
    skipButton.alpha = 0
    gsap.to(skipButton, { alpha: 1, duration: 0.3, delay: 0.3 })
  }

  /**
   * Create checkbox for "Don't show again" option
   */
  function createCheckbox(width: number, height: number): void {
    if (checkboxContainer) {
      checkboxContainer.destroy({ children: true })
      checkboxContainer = null
    }

    checkboxContainer = new Container()

    const isMobile = width < 600
    const boxSize = isMobile ? 20 : 24
    const fontSize = isMobile ? 12 : 14
    const spacing = 8

    // Create checkbox box
    const checkbox = new Graphics()
    const isChecked = skipAllCards

    // Draw checkbox background
    checkbox.roundRect(0, 0, boxSize, boxSize, 4)
    checkbox.fill({ color: 0x333333, alpha: 0.8 })
    checkbox.stroke({ color: 0xffffff, width: 2, alpha: 0.8 })

    // Draw checkmark if checked
    if (isChecked) {
      checkbox.moveTo(5, boxSize / 2)
      checkbox.lineTo(boxSize / 2 - 2, boxSize - 6)
      checkbox.lineTo(boxSize - 4, 6)
      checkbox.stroke({ color: 0x00ff00, width: 3 })
    }

    checkbox.eventMode = 'static'
    checkbox.cursor = 'pointer'

    // Toggle on click
    ;(checkbox as any).on('pointerdown', () => {
      skipAllCards = !skipAllCards
      // Redraw checkbox
      checkbox.clear()
      checkbox.roundRect(0, 0, boxSize, boxSize, 4)
      checkbox.fill({ color: 0x333333, alpha: 0.8 })
      checkbox.stroke({ color: 0xffffff, width: 2, alpha: 0.8 })
      if (skipAllCards) {
        checkbox.moveTo(5, boxSize / 2)
        checkbox.lineTo(boxSize / 2 - 2, boxSize - 6)
        checkbox.lineTo(boxSize - 4, 6)
        checkbox.stroke({ color: 0x00ff00, width: 3 })
      }
    })

    checkboxContainer.addChild(checkbox)

    // Create label text with shadow effect
    const labelStyle = new TextStyle({
      fontFamily: 'Arial, sans-serif',
      fontSize: fontSize,
      fill: 0xffffff,
      fontWeight: 'normal',
      dropShadow: {
        alpha: 0.8,
        angle: Math.PI / 6,
        blur: 4,
        color: 0x000000,
        distance: 2
      }
    })
    const label = new Text({ text: "Don't show cards", style: labelStyle })
    label.anchor.set(0, 0.5)
    label.x = boxSize + spacing
    label.y = boxSize / 2

    // Make label clickable too
    label.eventMode = 'static'
    label.cursor = 'pointer'
    
    // Add hover effects for label
    ;(label as any).on('pointerover', () => {
      gsap.to(label.scale, { x: 1.05, y: 1.05, duration: 0.2 })
      gsap.to(label, { alpha: 0.8, duration: 0.2 })
    })
    
    ;(label as any).on('pointerout', () => {
      gsap.to(label.scale, { x: 1, y: 1, duration: 0.2 })
      gsap.to(label, { alpha: 1, duration: 0.2 })
    })
    
    ;(label as any).on('pointerdown', () => {
      skipAllCards = !skipAllCards
      // Redraw checkbox
      checkbox.clear()
      checkbox.roundRect(0, 0, boxSize, boxSize, 4)
      checkbox.fill({ color: 0x333333, alpha: 0.8 })
      checkbox.stroke({ color: 0xffffff, width: 2, alpha: 0.8 })
      if (skipAllCards) {
        checkbox.moveTo(5, boxSize / 2)
        checkbox.lineTo(boxSize / 2 - 2, boxSize - 6)
        checkbox.lineTo(boxSize - 4, 6)
        checkbox.stroke({ color: 0x00ff00, width: 3 })
      }
    })

    checkboxContainer.addChild(label)

    // Position checkbox container between card and skip button
    const totalWidth = boxSize + spacing + label.width
    checkboxContainer.x = (width - totalWidth) / 2
    // Position at midpoint between card bottom and skip button top
    const cardBottom = height / 2 - 40 + (height * 0.7 / 2)
    const skipButtonTop = height - 150
    checkboxContainer.y = (cardBottom + skipButtonTop) / 2

    container.addChild(checkboxContainer)
    checkboxContainer.alpha = 0
    gsap.to(checkboxContainer, { alpha: 1, duration: 0.3, delay: 0.4 })
  }

  /**
   * Flip the card with 3D animation
   */
  function flipCard(): void {
    if (!frontSprite || !backSprite || !isActive) return
    if (currentTween) {
      currentTween.kill()
    }

    const flipDuration = 0.4
    const currentFront = isFlipped ? backSprite : frontSprite
    const currentBack = isFlipped ? frontSprite : backSprite

    currentTween = gsap.timeline()

    // First half: scale X to 0 (card turning away)
    currentTween.to(currentFront.scale, {
      x: 0,
      duration: flipDuration / 2,
      ease: 'power2.in',
      onComplete: () => {
        currentFront.visible = false
        currentBack.visible = true
        currentBack.scale.x = 0
      }
    })

    // Second half: scale X back to normal (card turning toward)
    currentTween.to(currentBack.scale, {
      x: cardScale,
      duration: flipDuration / 2,
      ease: 'power2.out'
    })

    isFlipped = !isFlipped
  }

  /**
   * Show the animation for a winning symbol
   * Returns a Promise that resolves when user clicks skip
   */
  function show(symbol: string, canvasWidth: number, canvasHeight: number, gridRect?: GridRect): Promise<void> {
    return new Promise((resolve) => {
      // Skip if user checked "Don't show cards"
      if (skipAllCards) {
        resolve()
        return
      }

      // If already showing this symbol, don't restart
      if (isActive && currentSymbol === symbol) {
        resolve()
        return
      }

      // Skip if same symbol was just shown
      if (lastShownSymbol === symbol) {
        resolve()
        return
      }

      // Check if symbol has card images
      const imageKeys = SYMBOL_IMAGE_KEYS[symbol]
      if (!imageKeys) {
        console.warn(`[SymbolWinAnimation] No images for symbol: ${symbol}`)
        resolve()
        return
      }

      // Hide any existing animation first
      if (isActive) {
        hide()
      }

      const frontSpriteNew = getWinningHighValueSprite(imageKeys.front)
      const backSpriteNew = getWinningHighValueSprite(imageKeys.back)

      if (!frontSpriteNew || !backSpriteNew) {
        console.warn(`[SymbolWinAnimation] Could not load sprites for: ${symbol}`)
        resolve()
        return
      }

      // Track this symbol as the last shown
      lastShownSymbol = symbol
      resolvePromise = resolve

      isActive = true
      currentSymbol = symbol
      frontSprite = frontSpriteNew
      backSprite = backSpriteNew
      isFlipped = false
      container.visible = true

      // Create dimmed overlay
      dimmedOverlay = new Graphics()
      dimmedOverlay.rect(0, 0, canvasWidth, canvasHeight)
      dimmedOverlay.fill({ color: 0x000000, alpha: 0.6 })
      dimmedOverlay.alpha = 0
      container.addChild(dimmedOverlay)

      // Setup front sprite
      frontSprite.anchor.set(0.5)
      frontSprite.x = Math.round(canvasWidth / 2)
      frontSprite.y = Math.round(canvasHeight / 2 - 80)

      // Size: max-width 90vw or 600px, max-height 70%
      const maxWidth = Math.min(canvasWidth * 0.9, 600)
      const maxHeight = canvasHeight * 0.7
      const scaleByWidth = maxWidth / frontSprite.width
      const scaleByHeight = maxHeight / frontSprite.height
      cardScale = Math.min(scaleByWidth, scaleByHeight)

      frontSprite.scale.set(0)
      frontSprite.alpha = 0
      frontSprite.visible = true

      // Setup back sprite (hidden initially)
      backSprite.anchor.set(0.5)
      backSprite.x = Math.round(canvasWidth / 2)
      backSprite.y = Math.round(canvasHeight / 2 - 80)
      backSprite.scale.set(cardScale)
      backSprite.visible = false

      // Make cards clickable for flip
      frontSprite.eventMode = 'static'
      frontSprite.cursor = 'pointer'
      ;(frontSprite as any).on('pointerdown', flipCard)

      backSprite.eventMode = 'static'
      backSprite.cursor = 'pointer'
      ;(backSprite as any).on('pointerdown', flipCard)

      container.addChild(frontSprite)
      container.addChild(backSprite)

      // Create skip button and checkbox
      createSkipButton(canvasWidth, canvasHeight)
      createCheckbox(canvasWidth, canvasHeight)

      // Animate in
      currentTween = gsap.timeline()

      // Play card transition sound when card appears
      audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'card_transition', volume: 0.7 })

      // Fade in dimmed overlay
      currentTween.to(dimmedOverlay, {
        alpha: 1,
        duration: 0.2,
        ease: 'power1.out'
      }, 0)

      // Pop in front sprite
      currentTween.to(frontSprite, {
        alpha: 1,
        duration: 0.15,
        ease: 'power1.out'
      }, 0.05)

      currentTween.to(frontSprite.scale, {
        x: cardScale,
        y: cardScale,
        duration: 0.35,
        ease: 'back.out(1.4)'
      }, 0.05)

      // No auto-close - user must click skip
    })
  }

  /**
   * Hide the animation
   */
  function hide(): void {
    if (!isActive) return

    // Kill any running tweens
    if (currentTween) {
      currentTween.kill()
      currentTween = null
    }

    // Remove dimmed overlay
    if (dimmedOverlay) {
      if (dimmedOverlay.parent) {
        container.removeChild(dimmedOverlay)
      }
      dimmedOverlay.destroy()
      dimmedOverlay = null
    }

    // Remove front sprite
    if (frontSprite) {
      ;(frontSprite as any).off('pointerdown')
      if (frontSprite.parent) {
        container.removeChild(frontSprite)
      }
      frontSprite.destroy()
      frontSprite = null
    }

    // Remove back sprite
    if (backSprite) {
      ;(backSprite as any).off('pointerdown')
      if (backSprite.parent) {
        container.removeChild(backSprite)
      }
      backSprite.destroy()
      backSprite = null
    }

    // Remove skip button
    if (skipButton) {
      gsap.killTweensOf(skipButton)
      ;(skipButton as any).off('pointerdown')
      ;(skipButton as any).off('pointerover')
      ;(skipButton as any).off('pointerout')
      if (skipButton.parent) {
        container.removeChild(skipButton)
      }
      skipButton.destroy()
      skipButton = null
    }

    // Remove checkbox
    if (checkboxContainer) {
      gsap.killTweensOf(checkboxContainer)
      if (checkboxContainer.parent) {
        container.removeChild(checkboxContainer)
      }
      checkboxContainer.destroy({ children: true })
      checkboxContainer = null
    }

    isActive = false
    currentSymbol = null
    isFlipped = false
    container.visible = false

    // Call onHide callback
    if (onHideCallback) {
      onHideCallback()
    }
  }

  /**
   * Set callback to be called when animation hides
   */
  function setOnHide(callback: () => void): void {
    onHideCallback = callback
  }

  /**
   * Update animation (called each frame if needed)
   */
  function update(): void {
    // GSAP handles the animation
  }

  /**
   * Check if animation is currently showing
   */
  function isShowing(): boolean {
    return isActive
  }

  /**
   * Get currently showing symbol
   */
  function getCurrentSymbol(): string | null {
    return currentSymbol
  }

  /**
   * Reset the last shown symbol tracker
   */
  function resetLastShown(): void {
    lastShownSymbol = null
  }

  return {
    container,
    show,
    hide,
    update,
    isShowing,
    getCurrentSymbol,
    setOnHide,
    resetLastShown
  }
}
