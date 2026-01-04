// @ts-nocheck
import { Container, Graphics, Sprite } from 'pixi.js'
import { GlowFilter } from 'pixi-filters'
import { useGameStore, useSettingsStore, useUIStore } from '@/stores'
import type { UseGameState } from '@/composables/slotMachine/useGameState'
import gsap from 'gsap'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'

import type { FooterRect, FooterHandlers } from './modules/types'
import { formatNumber } from './modules/textureUtils'
import {
  createParticleState,
  updateSpinButtonParticles,
  updateJackpotParticles,
  updateLightning,
  type ParticleState
} from './modules/particleEffects'
import {
  createSpinButtonState,
  buildSpinButton,
  setSpinButtonSpinning,
  updateSpinButton,
  type SpinButtonState
} from './modules/spinButton'
import { createGlyphNumber, type GlyphNumberDisplay } from './modules/glyphNumber'
import { getBackgroundSprite, getBackgroundTexture, getGlyphSprite, getIconSprite } from '@/config/spritesheet'

export type { FooterRect, FooterHandlers }

/**
 * High-value symbol types for frame theming
 */
export type FrameTheme = 'default' | 'fa' | 'zhong' | 'bai' | 'bawan'

export interface UseFooter {
  container: Container
  build: (rect: FooterRect) => void
  setHandlers: (handlers: FooterHandlers) => void
  update: (timestamp?: number) => void
  updateValues: () => void
  /** Get balance/wallet position for particle fly-to animation */
  getWalletPosition: () => { x: number; y: number } | null
  /** Called when particles reach the wallet - triggers pending balance update */
  onParticlesReachedWallet: () => void
  /** Set whether particles are currently active (delays balance updates) */
  setParticlesActive: (active: boolean) => void
  /** Set the frame theme (for high-value symbol wins) */
  setFrameTheme: (theme: FrameTheme) => void
}

/**
 * Play a UI sound effect with optional pitch randomization
 * @param audioKey - The audio key to play
 * @param pitchRange - Optional [min, max] range for pitch randomization (e.g., [0.6, 1.4])
 */
function playUISound(audioKey: string, pitchRange?: [number, number]): void {
  const howl = howlerAudio.getHowl(audioKey)
  if (!howl) return

  // Apply pitch randomization if specified
  if (pitchRange) {
    const [min, max] = pitchRange
    const randomPitch = min + Math.random() * (max - min)
    howl.rate(randomPitch)
  }

  audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey, volume: 0.6 })
}

export function useFooter(gameState: UseGameState): UseFooter {
  const gameStore = useGameStore()
  const settingsStore = useSettingsStore()
  const uiStore = useUIStore()

  const container = new Container()
  const menuContainer = new Container()
  const jackpotContainer = new Container()

  let handlers: FooterHandlers = { spin: () => {}, increaseBet: () => {}, decreaseBet: () => {} }
  const amountDisplays: Record<string, GlyphNumberDisplay> = {}
  const amountAlignments: Record<string, 'center' | 'left'> = {}  // Track alignment for each display
  const btns: Record<string, Sprite> = {}

  let balance = 0
  let betAmount = 0
  let winAmount = 0

  // Particle-synced balance update
  let particlesActive = false
  let pendingBalance: number | null = null  // Balance waiting for particles to reach wallet
  let x = 0, y = 0, w = 0, h = 0
  let inFreeSpinMode = false
  let spins = 0
  let lastTs = 0
  let displayedSpins = 0 // For animated count
  let isCountAnimating = false
  let countAnimationTween: gsap.core.Tween | null = null
  let numberSprites: Sprite[] = [] // Track number sprites for animation

  // State objects
  let spinButtonState: SpinButtonState = createSpinButtonState()
  let particleState: ParticleState = createParticleState()

  // UI containers
  let amountContainer: Container
  let mainMenuContainer: Container
  let btnHover: Graphics
  let footerBgSprite: Sprite | null = null
  let currentFrameTheme: FrameTheme = 'default'
  let turboLabelSprite: Sprite | null = null

  /**
   * Get the footer frame texture key based on theme
   */
  function getFooterTextureKey(theme: FrameTheme, isFreespin: boolean): string {
    if (isFreespin) {
      return 'background_footer_freespin.webp'
    }
    if (theme === 'default') return 'background_footer.webp'
    return `${theme}_background_footer.webp`
  }

  /**
   * Set the frame theme (for high-value symbol wins)
   */
  function setFrameTheme(theme: FrameTheme): void {
    if (theme === currentFrameTheme) return
    currentFrameTheme = theme

    if (footerBgSprite) {
      const isFreespin = gameState.inFreeSpinMode.value
      const textureKey = getFooterTextureKey(theme, isFreespin)
      const texture = getBackgroundTexture(textureKey)
      if (texture) {
        footerBgSprite.texture = texture
      }
    }
  }

  function setHandlers(h: FooterHandlers): void {
    handlers = { ...handlers, ...h }
  }

  function drawHoverCircle(sprite: Sprite, radius = 0, xPos: number | null = null): void {
    // Get sprite's global position and convert to footer container's local coords
    const globalPos = sprite.getGlobalPosition()
    // btnHover is a child of the footer container, so we need local coords relative to footer
    const localPos = container.toLocal(globalPos)
    const xPosition = xPos ?? localPos.x
    btnHover.clear()
    btnHover.circle(xPosition, localPos.y, radius).fill({ color: 0xffffff, alpha: 0.2 })
  }

  function startSpin(): void {
    if (spinButtonState.spinHoverCircle) {
      spinButtonState.spinHoverCircle.visible = false
    }
  }

  function switchMenuMode(): void {
    const centerX = x + Math.floor(w / 2)
    if (gameState.inFreeSpinMode.value) {
      amountContainer.visible = true  // Keep amount displays visible
      menuContainer.visible = false
      jackpotContainer.visible = false  // Free spin count is now shown via overlay
      // Switch to freespin footer background
      if (footerBgSprite) {
        const textureKey = getFooterTextureKey(currentFrameTheme, true)
        const tex = getBackgroundTexture(textureKey)
        if (tex) {
          footerBgSprite.texture = tex
          // Scale to fit width + 10% larger
          const scale = (w / tex.width) * 1.1
          footerBgSprite.scale.set(scale)
        }
        // Lower the freespin background
        const footerBgOffset = Math.round(h * -0.08)
        footerBgSprite.position.set(centerX, y + h + footerBgOffset)
      }
      // Reposition amounts for freespin footer holes
      repositionAmountsForFreespin(true)
    } else {
      // Amount container is already positioned in build()
      amountContainer.visible = true
      jackpotContainer.visible = false
      menuContainer.visible = true
      // Switch to themed footer background (use current theme, not always default)
      if (footerBgSprite) {
        const textureKey = getFooterTextureKey(currentFrameTheme, false)
        const tex = getBackgroundTexture(textureKey)
        if (tex) {
          footerBgSprite.texture = tex
          // Recalculate scale for new texture to maintain consistent size
          const scale = (w / tex.width) * 1.1
          footerBgSprite.scale.set(scale)
        }
        // Normal position for regular footer
        const footerBgOffset = Math.round(h * 0.01)
        footerBgSprite.position.set(centerX, y + h + footerBgOffset)
      }
      // Restore normal amount positions
      repositionAmountsForFreespin(false)
    }
  }

  function repositionAmountsForFreespin(isFreespin: boolean): void {
    if (!amountDisplays['bet'] || !amountDisplays['balance'] || !amountDisplays['win']) return

    if (isFreespin) {
      // Glyph height proportional to footer width for freespin mode
      const glyphHeight = Math.floor(w * 0.020) // 2.0% of footer width
      const freeSpinRemainingGlyphHeight = Math.floor(w * 0.05) // 10% of footer width - increased size
      currentGlyphHeight = glyphHeight
      freeSpinGlyphHeight = freeSpinRemainingGlyphHeight  // Store for updates

      // Freespin footer: Hide bet, show only balance and win
      // The freespin background is lowered, so we need to adjust positions accordingly
      const freespinFooterY = y + h * 0.08  // Match the lowered footer offset

      // Hide bet display in free spin mode
      if (amountDisplays['bet']) {
        amountDisplays['bet'].container.visible = false
      }

      // Balance - 70% from left edge, 45% from top
      const balanceX = x + w * 0.70  // 70% from left
      const balanceY = freespinFooterY + h * 0.465  // 46.5% from top
      amountDisplays['balance'].container.position.set(balanceX, balanceY)
      amountDisplays['balance'].container.visible = true
      // Update glyph height for balance with left alignment
      amountAlignments['balance'] = 'left'
      amountDisplays['balance'].update(formatNumber(balance), glyphHeight, 'left')

      // Win - 70% from left edge, 72% from top
      const winX = x + w * 0.70  // 70% from left
      const winY = freespinFooterY + h * 0.635  // 63.5% from top
      amountDisplays['win'].container.position.set(winX, winY)
      amountDisplays['win'].container.visible = true
      // Update glyph height for win with left alignment
      amountAlignments['win'] = 'left'
      amountDisplays['win'].update(formatNumber(winAmount), glyphHeight, 'left')

      // Free spin counter - 45% from left edge, 50% from top
      if (!amountDisplays['freespins']) {
        // Create free spin display if it doesn't exist
        const freeSpinDisplay = createGlyphNumber()
        amountDisplays['freespins'] = freeSpinDisplay
        amountContainer.addChild(freeSpinDisplay.container)
      }
      const freeSpinX = x + w * 0.47  // 47% from left
      const freeSpinY = freespinFooterY + h * 0.545  // 54.5% from top
      amountDisplays['freespins'].container.position.set(freeSpinX, freeSpinY)
      amountDisplays['freespins'].container.visible = true
      // Update glyph height for free spins with center alignment
      amountAlignments['freespins'] = 'center'
      amountDisplays['freespins'].update(String(gameState.freeSpins.value), freeSpinRemainingGlyphHeight, 'center')
    } else {
      // Normal footer: Use glyph height proportional to footer height (same as in build())
      const glyphHeight = Math.floor(h * 0.06)  // 6% of footer height
      currentGlyphHeight = glyphHeight

      // Normal footer positions (same as in build())
      // Show all displays
      if (amountDisplays['bet']) {
        amountDisplays['bet'].container.visible = true
      }
      if (amountDisplays['balance']) {
        amountDisplays['balance'].container.visible = true
      }
      if (amountDisplays['win']) {
        amountDisplays['win'].container.visible = true
      }
      // Hide free spin counter in normal mode
      if (amountDisplays['freespins']) {
        amountDisplays['freespins'].container.visible = false
      }

      const leftSlotX = x + w * 0.23
      const leftSlotY = y + h * 0.495
      const middleSlotX = x + w * 0.52
      const middleSlotY = y + h * 0.92
      const rightSlotX = x + w * 0.8
      const rightSlotY = y + h * 0.495

      amountDisplays['bet'].container.position.set(leftSlotX, leftSlotY)
      amountDisplays['balance'].container.position.set(middleSlotX, middleSlotY)
      amountDisplays['win'].container.position.set(rightSlotX, rightSlotY)

      // Update all displays with new glyph height and current values (center-aligned)
      amountAlignments['bet'] = 'center'
      amountAlignments['balance'] = 'center'
      amountAlignments['win'] = 'center'
      amountDisplays['bet'].update(formatNumber(betAmount), glyphHeight)
      amountDisplays['balance'].update(formatNumber(balance), glyphHeight)
      amountDisplays['win'].update(formatNumber(winAmount), glyphHeight)
    }
  }

  function buildJackpotContainer(): void {
    const particlesTemp = particleState.jackpotParticlesContainer
    jackpotContainer.removeChildren()
    numberSprites = [] // Reset tracked sprites

    // Use displayedSpins for animated display
    const freeSpins = isCountAnimating ? displayedSpins : gameState.freeSpins.value
    const menuHeight = h
    // Position between multiplier and grid - just above the footer area
    // y is footer top position, so we go slightly above it
    jackpotContainer.position.set(x, y - h * 0.35)
    container.addChild(jackpotContainer)

    if (!particleState.jackpotParticlesContainer) {
      particleState.jackpotParticlesContainer = new Container()
    } else {
      particleState.jackpotParticlesContainer = particlesTemp
    }
    jackpotContainer.addChild(particleState.jackpotParticlesContainer!)

    // Only show the free spins number (no text labels)
    if (freeSpins <= 0) {
      // No display when free spins are exhausted
      return
    }

    // Scale factor for display
    const scaleFactor = 0.6
    const targetHeight = menuHeight * 0.4 * scaleFactor

    const digits = String(freeSpins).split('')
    let totalWidth = 0
    const digitSprites: Sprite[] = []

    // First pass: create sprites and calculate total width
    for (const d of digits) {
      if (d === '.' || d === ',') continue

      const sprite = getGlyphSprite(`glyph_${d}.webp`)
      if (!sprite) continue
      // Use linear filtering with mipmaps for high-quality rendering on retina displays
      if (sprite.texture?.source) {
        sprite.texture.source.scaleMode = 'linear'
        sprite.texture.source.autoGenerateMipmaps = true
      }
      sprite.anchor.set(0)
      sprite.scale.set(targetHeight / sprite.height)
      digitSprites.push(sprite)
      totalWidth += sprite.width
    }

    // Second pass: position sprites centered
    let offsetX = (w - totalWidth) / 2
    for (const sprite of digitSprites) {
      sprite.x = offsetX
      sprite.y = 0
      jackpotContainer.addChild(sprite)
      numberSprites.push(sprite)
      offsetX += sprite.width
    }
  }

  /**
   * Animate the free spins count increasing
   */
  function animateFreeSpinsCount(fromCount: number, toCount: number): void {
    if (isCountAnimating && countAnimationTween) {
      countAnimationTween.kill()
    }

    isCountAnimating = true
    displayedSpins = fromCount

    // Create animation object for tweening
    const animObj = { count: fromCount }

    countAnimationTween = gsap.to(animObj, {
      count: toCount,
      duration: 1.2,
      ease: 'power2.out',
      onUpdate: () => {
        displayedSpins = Math.round(animObj.count)
        buildJackpotContainer()

        // Add pulse effect to number sprites
        numberSprites.forEach((sprite, index) => {
          const scale = 0.8 * (jackpotContainer.children[1] as Sprite)?.height / sprite.texture.height || 1
          const pulse = 1 + Math.sin(Date.now() * 0.01 + index * 0.5) * 0.1
          sprite.scale.set(scale * pulse)
        })
      },
      onComplete: () => {
        isCountAnimating = false
        displayedSpins = toCount
        buildJackpotContainer()

        // Final pop animation
        numberSprites.forEach((sprite, index) => {
          gsap.fromTo(sprite.scale,
            { x: sprite.scale.x * 1.3, y: sprite.scale.y * 1.3 },
            { x: sprite.scale.x, y: sprite.scale.y, duration: 0.3, ease: 'back.out(2)', delay: index * 0.05 }
          )
        })
      }
    })
  }

  function buildMenuContainer(): void {
    menuContainer.removeChildren()
    const menuHeight = 0.5 * h
    menuContainer.position.set(x, y + menuHeight + 0.05 * h)
    container.addChild(menuContainer)

    // Add mask to clip menu content - extend upward to allow spin button to show
    const menuMask = new Graphics()
    menuMask.rect(0, -menuHeight * 0.5, w, menuHeight * 1.5)  // Extended upward for spin button
    menuMask.fill(0xffffff)
    menuContainer.mask = menuMask
    menuContainer.addChild(menuMask)

    mainMenuContainer = new Container()
    menuContainer.addChild(mainMenuContainer)

    const targetButtonHeightPer = 0.308  // Scaled up by 10% to better fill holes
    // Position buttons in the menu area - lowered to sit in holes
    const yPosition = 0.347 * menuHeight
    const spinBtnY = 0.13 * menuHeight  // Positive value to lower spin button

    // Button positions relative to center (spin button at center)
    const centerX = w * 0.5

     // Spin button (center) - raised up
    buildSpinButton(
      mainMenuContainer,
      spinButtonState,
      particleState,
      menuHeight,
      centerX,
      spinBtnY,
      handlers,
      gameState,
      startSpin
    )

    // Positions based on footer background hole layout
    // 4 holes around center spin: outer-left, inner-left, inner-right, outer-right
    const isSmartphone = w < 600
    const innerHoleOffset = w * 0.182        // Distance from center to inner holes (minus/plus)
    const outerHoleOffset = w * 0.31        // Distance from center to outer holes (game mode/bet setting)

    // Most left hole - Turbo/Fast Spin toggle button
    buildHoleButton(mainMenuContainer, 'icon_btn_toggle_turbo_mode.webp', centerX - outerHoleOffset + (isSmartphone ? 2 : 3), yPosition - (isSmartphone ? 1 : 2), menuHeight, targetButtonHeightPer, () => {
      // Play generic UI sound with pitch randomization (0.6-1.4)
      playUISound('generic_ui', [0.6, 1.4])
      // Toggle turbo/fast spin mode
      settingsStore.toggleFastSpin()
    })

    // Inner left hole (left of spin) - Minus button
    buildHoleButton(mainMenuContainer, 'icon_decrease_bet_amount.webp', centerX - innerHoleOffset + (isSmartphone ? 1 : 1.5), yPosition - (isSmartphone ? 1 : 2), menuHeight, targetButtonHeightPer, () => {
      // Play decrease bet sound
      playUISound('decrease_bet')
      handlers.decreaseBet()
    })

    // Inner right hole (right of spin) - Plus button
    buildHoleButton(mainMenuContainer, 'icon_increase_bet_amount.webp', centerX + innerHoleOffset - (isSmartphone ? 2 : 3.5), yPosition - (isSmartphone ? 1 : 2), menuHeight, targetButtonHeightPer, () => {
      // Play increase bet sound
      playUISound('increase_bet')
      handlers.increaseBet()
    })

    // Most right hole - Game Mode selection button (opens game mode overlay)
    buildHoleButton(mainMenuContainer, 'icon_btn_game_mode_setting.webp', centerX + outerHoleOffset - (isSmartphone ? 2 : 4), yPosition - (isSmartphone ? 1 : 2), menuHeight, targetButtonHeightPer, () => {
      // Play generic UI sound with pitch randomization (0.6-1.4)
      playUISound('generic_ui', [0.6, 1.4])
      if (handlers.onGameModeClick) {
        handlers.onGameModeClick()
      }
    })
  }

  /**
   * Build a button for the footer holes using the new button icons
   */
  function buildHoleButton(
    container: Container,
    iconKey: string,
    xPos: number,
    yPos: number,
    menuHeight: number,
    heightPer: number,
    onClick?: () => void
  ): void {
    const sprite = getIconSprite(iconKey)
    if (!sprite) return

    // Draw circle background behind the button
    const circleRadius = menuHeight * heightPer * 0.55
    const circleBg = new Graphics()
    circleBg.circle(xPos, yPos, circleRadius)
    circleBg.stroke({ width: 2, color: 0xffffff, alpha: 0.3 })
    container.addChild(circleBg)

    sprite.anchor.set(0.5)
    sprite.position.set(xPos, yPos)
    sprite.scale.set(menuHeight * heightPer / sprite.height)
    sprite.eventMode = 'static'
    sprite

    const hoverRadius = menuHeight * heightPer * 0.5

    sprite.on('pointerover', () => {
      drawHoverCircle(sprite, hoverRadius)
      btnHover.visible = true
    })
    sprite.on('pointerout', () => {
      btnHover.visible = false
    })
    if (onClick) {
      sprite.on('pointerdown', () => {
        btnHover.visible = false
        onClick()
      })
    }

    container.addChild(sprite)
    btns[iconKey] = sprite
  }

  function build(rect: FooterRect): void {
    container.removeChildren()
    x = rect.x
    y = rect.y
    w = rect.w
    h = rect.h

    // Clear all amount displays to rebuild them
    Object.keys(amountDisplays).forEach(key => {
      delete amountDisplays[key]
    })

    const setRectHeight = Math.floor(h * 0.5)
    const centerX = x + Math.floor(w / 2)

    // Footer background - anchor at bottom center so it aligns with screen bottom
    // Use freespin background during free spin mode
    const bgKey = getFooterTextureKey(currentFrameTheme, gameState.inFreeSpinMode.value)
    const isSmartphone = w < 600
    footerBgSprite = getBackgroundSprite(bgKey)
    if (footerBgSprite) {
      footerBgSprite.anchor.set(0.5, 1.0)  // Anchor at bottom center
      // Scale to fit width + 10% larger, keeping aspect ratio
      const scale = (w / footerBgSprite.width) * 1.1
      footerBgSprite.scale.set(scale)
      // Position at bottom of footer rect with offset
      // Freespin background sits lower than normal footer
      const footerBgOffset = gameState.inFreeSpinMode.value
        ? Math.round(h * 0.15)  // Freespin: lower position
        : Math.round(h * 0.01)  // Normal: original position
      footerBgSprite.position.set(centerX, y + h + footerBgOffset)

      container.addChild(footerBgSprite)
    }

    // Turbo mode label - shown in bottom left of footer when turbo mode is active
    turboLabelSprite = getIconSprite('icon_turbo_label.webp')
    if (turboLabelSprite) {
      turboLabelSprite.anchor.set(0, 1)  // Anchor at bottom left
      // Position in bottom left corner of footer frame
      const footerLeftEdge = x + w * 0.05  // 5% from left edge
      const footerBottom = y + h - h * 0.05  // 5% from bottom
      turboLabelSprite.position.set(footerLeftEdge, footerBottom)
      // Scale to fit nicely within the footer
      const labelScale = (h * 0.12) / turboLabelSprite.height
      turboLabelSprite.scale.set(labelScale)
      // Initially hidden - will be shown/hidden in update based on fastSpin state
      turboLabelSprite.visible = settingsStore.fastSpin
      container.addChild(turboLabelSprite)
    }

    // Amount displays - positioned in the new footer slots
    // Left green area: balance, Center top: bet, Right cyan area: win
    amountContainer = new Container()
    container.addChild(amountContainer)

    balance = gameState.credits.value
    betAmount = gameState.bet.value
    winAmount = gameState.currentWin.value

    // Build glyph displays at initial positions (will be repositioned by switchMenuMode)
    // Positions based on normal footer layout - aligned with input holes
    const leftSlotX = x + w * 0.23     // Left slot
    const leftSlotY = y + h * 0.49     // Y position for left slot
    const middleSlotX = x + w * 0.52   // Middle slot (center)
    const middleSlotY = y + h * 0.92   // Y position for middle slot
    const rightSlotX = x + w * 0.8     // Right slot
    const rightSlotY = y + h * 0.49    // Y position for right slot

    const glyphHeight = Math.floor(h * 0.06)  // Height for glyph numbers

    buildGlyphAmount('bet', formatNumber(betAmount), leftSlotX, leftSlotY, glyphHeight)        // Bet on left
    buildGlyphAmount('balance', formatNumber(balance), middleSlotX, middleSlotY, glyphHeight)  // Balance in middle
    buildGlyphAmount('win', formatNumber(winAmount), rightSlotX, rightSlotY, glyphHeight)      // Win on right

    buildMenuContainer()
    buildJackpotContainer()

    inFreeSpinMode = gameState.inFreeSpinMode.value
    spins = gameState.freeSpins.value
    switchMenuMode()

    btnHover = new Graphics()
    btnHover.visible = false
    container.addChild(btnHover)
  }

  // Store glyph height for updates
  let currentGlyphHeight = 20
  let freeSpinGlyphHeight = 50  // Separate size for freespin counter

  function buildGlyphAmount(key: string, text: string, centerX: number, yPos: number, glyphHeight: number): void {
    currentGlyphHeight = glyphHeight
    const display = createGlyphNumber()
    display.update(text, glyphHeight)
    display.container.position.set(centerX, yPos)
    amountDisplays[key] = display
    amountAlignments[key] = 'center'  // Default to center alignment
    amountContainer.addChild(display.container)
  }

  function tweenGlyphNumber(display: GlyphNumberDisplay, start: number, end: number, duration = 0.5, align?: 'center' | 'left'): void {
    const obj = { value: start }
    gsap.to(obj, {
      value: end,
      duration,
      ease: 'power1.out',
      onUpdate: () => {
        display.update(formatNumber(obj.value), currentGlyphHeight, align)
      }
    })
  }

  function updateValues(): void {
    const newCredits = gameState.credits.value

    // Check if balance increased (win scenario)
    const isBalanceIncrease = newCredits > balance

    if (newCredits !== balance) {
      if (isBalanceIncrease && particlesActive) {
        // Balance increased and particles are flying - queue the update
        pendingBalance = newCredits
      } else {
        // Balance decreased (bet deduction) or no particles - update immediately
        tweenGlyphNumber(amountDisplays['balance'], balance, newCredits, 0.5, amountAlignments['balance'])
        balance = newCredits
        pendingBalance = null
      }
    }

    if (gameState.bet.value !== betAmount) {
      amountDisplays['bet'].update(formatNumber(gameState.bet.value), currentGlyphHeight, amountAlignments['bet'])
      betAmount = gameState.bet.value
    }

    let accumulatedWinAmount = gameState.accumulatedWinAmount.value
    if (gameState.inFreeSpinMode.value) {
      accumulatedWinAmount = gameState.freeSpinSessionWinAmount.value
    }

    if (accumulatedWinAmount !== winAmount) {
      if (winAmount < accumulatedWinAmount) {
        tweenGlyphNumber(amountDisplays['win'], winAmount, accumulatedWinAmount, 0.5, amountAlignments['win'])
      } else {
        amountDisplays['win'].update(formatNumber(accumulatedWinAmount), currentGlyphHeight, amountAlignments['win'])
      }
      winAmount = accumulatedWinAmount
    }

    // Win amount notification removed - using glyph numbers in footer instead
    if (gameState.isShowAmountNotification.value) {
      gameStore.setShowAmountNotification(false)
    }

    // Update free spins counter in footer if in free spin mode
    if (gameState.inFreeSpinMode.value && amountDisplays['freespins']) {
      amountDisplays['freespins'].update(String(gameState.freeSpins.value), freeSpinGlyphHeight, 'center')
    }
  }

  function update(timestamp = 0): void {
    const dt = lastTs ? Math.max(0, (timestamp - lastTs) / 1000) : 0
    lastTs = timestamp

    if (inFreeSpinMode !== gameState.inFreeSpinMode.value || spins !== gameState.freeSpins.value) {
      const previousSpins = spins
      const newSpins = gameState.freeSpins.value
      inFreeSpinMode = gameState.inFreeSpinMode.value
      spins = newSpins
      switchMenuMode()
      startSpin()

      // Animate free spins count increasing (retrigger case)
      if (inFreeSpinMode && newSpins > previousSpins && previousSpins > 0) {
        animateFreeSpinsCount(previousSpins, newSpins)
      }
    }

    // Spin button
    const isSpinning = !!gameState.isSpinning?.value
    setSpinButtonSpinning(spinButtonState, isSpinning, !!gameState.canSpin?.value)
    updateSpinButton(spinButtonState, particleState, isSpinning, dt)

    // Update turbo label visibility based on fast spin setting
    if (turboLabelSprite) {
      turboLabelSprite.visible = settingsStore.fastSpin && !gameState.inFreeSpinMode.value
    }

    // Particle effects
    updateSpinButtonParticles(particleState)
    updateLightning(particleState, isSpinning, spinButtonState.spinBtnSprite!, timestamp)
    updateJackpotParticles(particleState, gameState.inFreeSpinMode?.value || false, w, h)
  }

  /**
   * Get the balance/wallet display position for particle fly-to animation
   */
  function getWalletPosition(): { x: number; y: number } | null {
    const balanceDisplay = amountDisplays['balance']
    if (!balanceDisplay) return null

    // Return the global position of the balance container
    const globalPos = balanceDisplay.container.getGlobalPosition()
    return {
      x: globalPos.x,
      y: globalPos.y
    }
  }

  /**
   * Called when particles reach the wallet - triggers pending balance update
   */
  function onParticlesReachedWallet(): void {
    if (pendingBalance !== null) {
      // Particles have arrived - now animate the balance increase
      tweenGlyphNumber(amountDisplays['balance'], balance, pendingBalance, 0.5, amountAlignments['balance'])
      balance = pendingBalance
      pendingBalance = null
    }
    particlesActive = false
  }

  /**
   * Set whether particles are currently active
   */
  function setParticlesActive(active: boolean): void {
    particlesActive = active
  }

  return { container, build, setHandlers, update, updateValues, getWalletPosition, onParticlesReachedWallet, setParticlesActive, setFrameTheme }
}
