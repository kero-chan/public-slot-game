// @ts-nocheck
import { Container, Graphics, Sprite, Text } from 'pixi.js'
import gsap from 'gsap'
import { getGameModeSprite, getGlyphSprite, getIconSprite } from '@/config/spritesheet'
import {
  createTimerManager,
  createDarkOverlay,
  type BaseOverlay
} from './base'
import { audioEvents, AUDIO_EVENTS } from '@/composables/audioEventBus'
import { howlerAudio } from '@/composables/useHowlerAudio'

/**
 * Play generic UI sound with pitch randomization (0.6-1.4)
 */
function playGenericUISound(): void {
  const howl = howlerAudio.getHowl('generic_ui')
  if (!howl) return

  // Apply pitch randomization 0.6-1.4
  const randomPitch = 0.6 + Math.random() * 0.8
  howl.rate(randomPitch)

  audioEvents.emit(AUDIO_EVENTS.EFFECT_PLAY, { audioKey: 'generic_ui', volume: 0.6 })
}

/**
 * Game mode option configuration
 */
interface GameModeOption {
  id: string
  cost: number
  description?: string
}

/**
 * Game mode selection overlay interface
 */
export interface GameModeSelectionOverlay extends BaseOverlay {
  show: (cWidth: number, cHeight: number, onSelect: (modeId: string) => void) => void
}

// ============================================================================
// CONFIGURATION - Easy to modify for future changes
// ============================================================================

/**
 * Game mode items configuration
 * To add/remove items: modify this array
 */
const GAME_MODES: GameModeOption[] = [
  {
    id: 'bonus_spin_trigger',
    cost: 750,
    description: 'Once unsealed, the flames will answer only with abundance.'
  }
]

/**
 * Menu background configuration
 */
const MENU_CONFIG = {
  backgroundImage: 'game_mode_menu_background.webp',
  maxWidthPercent: 0.85, // Max 85% of canvas width
  maxHeightPercent: 0.8, // Max 80% of canvas height
  closeButton: {
    image: 'icon_close.webp',
    sizePercent: 0.06, // 6% of menu height
    position: {
      xOffset: 0.085, // 8.5% from right edge
      yOffset: 0.05 // 5% from top edge
    }
  }
} as const

/**
 * Menu item layout configuration
 * To change positions of child elements: modify these values
 */
const MENU_ITEM_LAYOUT = {
  backgroundImage: 'game_mode_menu_item_background.webp',
  // Item sizing
  heightPercent: 0.24, // Item height = 24% of menu height
  maxWidthPercent: 0.68, // Max 68% of menu width

  // Cost display positioning (relative to item center)
  costDisplay: {
    // Scales relative to item scale
    costTextScaleMultiplier: 0.8, // Cost text fontSize multiplier (relative to coinsTextScale)
    costNumberScale: 0.30, // Cost sprite scale (when useSprites is true)
    coinBgScale: 0.22,
    coinsTextScale: 30,
    useSprites: false, // true = use sprites, false = use text (default)
    // Position at bottom right corner
    position: {
      paddingXPercent: 0.05, // 5% padding from right edge
      paddingYPercent: 0.05 // 5% padding from bottom edge
    }
  }
} as const

/**
 * Items list layout configuration
 */
const ITEMS_LIST_LAYOUT = {
  gapPercent: 0.01, // Gap between items = 1% of menu height
  startPositionPercent: 0.11, // Start position = 11% from top of menu background
  verticalAlignment: 'top' as const // 'center' | 'top' | 'bottom'
} as const

/**
 * Cost display layout configuration
 */
const COST_DISPLAY_LAYOUT = {
  coinBackgroundImage: 'game_mode_menu_item_coin_background.webp',
  coinsText: {
    text: 'COINS',
    style: {
      fontFamily: 'Arial Black, sans-serif',
      fontWeight: 'bold' as const,
      fill: '#FFD700', // Gold color
      stroke: { color: '#000000', width: 2 },
      dropShadow: { color: '#000000', blur: 4, distance: 2, alpha: 0.8 }
    }
  },
  // Number positioning relative to coin background center
  numberPosition: {
    yOffset: -0.3 // Negative = above center
  },
  // COINS text positioning relative to coin background center
  coinsTextPosition: {
    yOffset: 1.5 // Below number
  }
} as const

/**
 * Animation configuration
 */
const ANIMATION_CONFIG = {
  itemHover: {
    scale: 1.03,
    duration: 0.15,
    ease: 'back.out(2)'
  },
  itemSelect: {
    scale: 1.08,
    duration: 0.1,
    yoyo: true,
    repeat: 1,
    ease: 'power2.inOut'
  },
  itemEntry: {
    slideUpDistance: 100, // Distance to slide up from (in pixels)
    duration: 0.4,
    delayPerItem: 0.15,
    ease: 'back.out(1.2)'
  },
  menuEntry: {
    slideUpDistance: 150, // Distance to slide up from (in pixels)
    duration: 0.5,
    ease: 'back.out(1.2)',
    backgroundFadeDuration: 0.3,
    itemsStartDelay: 0.2 // Delay before items start animating (after background appears)
  },
  selectDelay: 250 // ms before calling onSelect callback
} as const

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Calculate item scale based on menu dimensions
 * Scales to ensure width reaches maxWidthPercent, height will adjust proportionally
 */
function calculateItemScale(
  itemBg: Sprite,
  menuWidth: number,
  menuHeight: number
): number {
  const originalWidth = itemBg.texture.width
  const maxWidth = menuWidth * MENU_ITEM_LAYOUT.maxWidthPercent

  // Scale based on width to reach maxWidthPercent
  return maxWidth / originalWidth
}

/**
 * Calculate items list positioning
 * menuCenterY: center Y position of menu (relative to menuContainer, usually 0)
 */
function calculateItemsListPosition(
  itemHeights: number[],
  menuCenterY: number,
  menuHeight: number
): { startY: number; gap: number } {
  const gap = menuHeight * ITEMS_LIST_LAYOUT.gapPercent

  // Calculate start position from top of menu background
  // menuCenterY is center (usually 0), so top is menuCenterY - menuHeight / 2
  const menuTop = menuCenterY - menuHeight / 2
  const startOffset = menuHeight * ITEMS_LIST_LAYOUT.startPositionPercent

  // Start Y position: from top of menu + offset
  // First item's center will be at menuTop + startOffset + itemHeights[0] / 2
  const startY = menuTop + startOffset

  return { startY, gap }
}

/**
 * Calculate Y position for a specific item in the list
 * listStartY is the top position where items should start
 */
function calculateItemYPosition(
  index: number,
  itemHeights: number[],
  gap: number,
  listStartY: number
): number {
  // Start from listStartY, add half height of first item to center it
  let cumulativeHeight = itemHeights[0] / 2

  // For subsequent items, add previous items and gaps
  for (let i = 0; i < index; i++) {
    cumulativeHeight += itemHeights[i] + gap
  }

  // For items after the first, adjust by half the difference between current and first item height
  if (index > 0) {
    cumulativeHeight += (itemHeights[index] - itemHeights[0]) / 2
  }

  return listStartY + cumulativeHeight
}

/**
 * Calculate cost display position at bottom right corner
 */
function calculateCostDisplayPosition(
  costDisplay: Container,
  itemWidth: number,
  itemHeight: number
): { x: number; y: number } {
  const bounds = costDisplay.getBounds()
  const paddingX = itemWidth * MENU_ITEM_LAYOUT.costDisplay.position.paddingXPercent
  const paddingY = itemHeight * MENU_ITEM_LAYOUT.costDisplay.position.paddingYPercent

  return {
    x: itemWidth / 2 - paddingX - bounds.width / 2,
    y: itemHeight / 2 - paddingY - bounds.height / 2
  }
}

// ============================================================================
// MAIN COMPONENT CREATION FUNCTIONS
// ============================================================================

/**
 * Create cost display with coin background and "coins" text
 * Layout: vertical stack (number on top, "COINS" below), centered on coin background
 */
function createCostDisplay(
  cost: number,
  scale: number,
  coinBgScale: number,
  textScale: number,
  useSprites: boolean = false
): Container {
  const costContainer = new Container()

  // Coin background
  const coinBg = getGameModeSprite(COST_DISPLAY_LAYOUT.coinBackgroundImage)
  if (coinBg) {
    coinBg.anchor.set(0.5)
    coinBg.scale.set(coinBgScale)
    costContainer.addChild(coinBg)
  }

  // Create cost number - either as text or sprites
  if (useSprites) {
    // Create number sprites
    const digits = cost.toLocaleString().split('')
    let totalWidth = 0
    const sprites: Sprite[] = []

    for (const char of digits) {
      let sprite: Sprite | null = null
      if (char === ',') {
        sprite = getGlyphSprite('glyph_comma.webp')
        if (sprite) {
          sprite.scale.set(scale * 0.5)
          sprite.anchor.set(0, 1)
        }
      } else if (char >= '0' && char <= '9') {
        sprite = getGlyphSprite(`glyph_${char}.webp`)
        if (sprite) {
          sprite.scale.set(scale)
          sprite.anchor.set(0, 0.5)
        }
      }

      if (sprite) {
        sprites.push(sprite)
        totalWidth += sprite.width
      }
    }

    // Position number sprites
    const digitSprite = sprites.find((s, idx) => digits[idx] !== ',')
    const digitHeight = digitSprite ? digitSprite.height : 0
    let offsetX = -totalWidth / 2
    const numberY = textScale * COST_DISPLAY_LAYOUT.numberPosition.yOffset

    for (let i = 0; i < sprites.length; i++) {
      const sprite = sprites[i]
      const char = digits[i]
      sprite.x = offsetX

      if (char === ',') {
        sprite.y = numberY + digitHeight * 0.7
      } else {
        sprite.y = numberY - digitHeight * 0.2
      }

      costContainer.addChild(sprite)
      offsetX += sprite.width
    }
  } else {
    // Create cost number text (similar style to coins text)
    const costText = new Text({
      text: cost.toLocaleString(),
      style: {
        ...COST_DISPLAY_LAYOUT.coinsText.style,
        fontSize: textScale * 1.6, // Scale based on costTextScaleMultiplier
        align: 'center'
      }
    })
    costText.anchor.set(0.5, 0.5)
    costText.x = 0
    costText.y = textScale * COST_DISPLAY_LAYOUT.numberPosition.yOffset

    costContainer.addChild(costText)
  }

  // Add "COINS" text
  const coinsText = new Text({
    text: COST_DISPLAY_LAYOUT.coinsText.text,
    style: {
      ...COST_DISPLAY_LAYOUT.coinsText.style,
      fontSize: textScale,
      align: 'center'
    }
  })
  coinsText.anchor.set(0.5, 0.5)
  coinsText.x = 0
  coinsText.y = textScale * COST_DISPLAY_LAYOUT.coinsTextPosition.yOffset

  costContainer.addChild(coinsText)

  return costContainer
}

/**
 * Create menu item with all child elements
 */
function createMenuItem(
  mode: GameModeOption,
  menuWidth: number,
  menuHeight: number,
  index: number,
  onSelect: (modeId: string) => void
): Container {
  const itemContainer = new Container()

  // Item background
  const itemBg = getGameModeSprite(MENU_ITEM_LAYOUT.backgroundImage)
  if (!itemBg) {
    return itemContainer
  }

  // Calculate scale
  const finalScale = calculateItemScale(itemBg, menuWidth, menuHeight)
  itemBg.anchor.set(0.5, 0.5)
  itemBg.scale.set(finalScale)
  itemBg.x = 0
  itemBg.y = 0
  itemContainer.addChild(itemBg)

  // Get scaled dimensions
  const scaledWidth = itemBg.texture.width * finalScale
  const scaledHeight = itemBg.texture.height * finalScale

  // Cost display
  const costDisplay = createCostDisplay(
    mode.cost,
    MENU_ITEM_LAYOUT.costDisplay.useSprites
      ? finalScale * MENU_ITEM_LAYOUT.costDisplay.costNumberScale
      : MENU_ITEM_LAYOUT.costDisplay.costTextScaleMultiplier,
    finalScale * MENU_ITEM_LAYOUT.costDisplay.coinBgScale,
    finalScale * MENU_ITEM_LAYOUT.costDisplay.coinsTextScale,
    MENU_ITEM_LAYOUT.costDisplay.useSprites
  )

  const costPosition = calculateCostDisplayPosition(costDisplay, scaledWidth, scaledHeight)
  costDisplay.x = costPosition.x
  costDisplay.y = costPosition.y
  itemContainer.addChild(costDisplay)

  // Store item height for positioning
  ;(itemContainer as any)._itemHeight = scaledHeight

  // Interaction
  itemContainer.eventMode = 'static'
  itemContainer.cursor = 'pointer'

  itemContainer.on('pointerover', () => {
    gsap.to(itemContainer.scale, {
      x: ANIMATION_CONFIG.itemHover.scale,
      y: ANIMATION_CONFIG.itemHover.scale,
      duration: ANIMATION_CONFIG.itemHover.duration,
      ease: ANIMATION_CONFIG.itemHover.ease
    })
  })

  itemContainer.on('pointerout', () => {
    gsap.to(itemContainer.scale, { x: 1, y: 1, duration: ANIMATION_CONFIG.itemHover.duration })
  })

  itemContainer.on('pointerdown', (e) => {
    e.stopPropagation()
    onSelect(mode.id)
  })

  return itemContainer
}

/**
 * Create close button
 */
function createCloseButton(
  menuBg: Sprite,
  menuWidth: number,
  menuHeight: number,
  onClose: () => void
): Sprite | null {
  const closeBtn = getIconSprite(MENU_CONFIG.closeButton.image)
  if (!closeBtn) return null

  closeBtn.anchor.set(0.5)
  const closeBtnScale = (menuHeight * MENU_CONFIG.closeButton.sizePercent) / closeBtn.height
  closeBtn.scale.set(closeBtnScale)
  closeBtn.x = menuBg.x + menuWidth / 2 - menuWidth * MENU_CONFIG.closeButton.position.xOffset
  closeBtn.y = menuBg.y - menuHeight / 2 + menuHeight * MENU_CONFIG.closeButton.position.yOffset
  closeBtn.eventMode = 'static'
  closeBtn.cursor = 'pointer'

  closeBtn.on('pointerover', () => {
    gsap.to(closeBtn.scale, {
      x: closeBtnScale * 1.1,
      y: closeBtnScale * 1.1,
      duration: 0.15
    })
  })

  closeBtn.on('pointerout', () => {
    gsap.to(closeBtn.scale, { x: closeBtnScale, y: closeBtnScale, duration: 0.15 })
  })

  closeBtn.on('pointerdown', (e) => {
    e.stopPropagation()
    // Play generic UI sound
    playGenericUISound()
    onClose()
  })

  return closeBtn
}

/**
 * Apply entry animation to menu item - slide up from below
 */
function animateItemEntry(item: Container, index: number, finalY: number): void {
  // Start position: below final position
  const startY = finalY + ANIMATION_CONFIG.itemEntry.slideUpDistance
  item.y = startY
  item.alpha = 0

  // Calculate delay: background delay + per-item delay
  const delay = ANIMATION_CONFIG.menuEntry.itemsStartDelay + (ANIMATION_CONFIG.itemEntry.delayPerItem * index)

  // Animate to final position
  gsap.to(item, {
    y: finalY,
    alpha: 1,
    duration: ANIMATION_CONFIG.itemEntry.duration,
    delay: delay,
    ease: ANIMATION_CONFIG.itemEntry.ease
  })
}

/**
 * Apply entry animation to menu container - slide up from below
 * Background fades in first, then menu slides up
 * finalY: final Y position of menuContainer (relative to parent container)
 */
function animateMenuEntry(
  menuContainer: Container,
  finalY: number,
  backgroundContainer: Container
): void {
  // Start position: below final position
  const startY = finalY + ANIMATION_CONFIG.menuEntry.slideUpDistance
  menuContainer.y = startY
  menuContainer.alpha = 0

  // Background fades in first
  backgroundContainer.alpha = 0
  gsap.to(backgroundContainer, {
    alpha: 1,
    duration: ANIMATION_CONFIG.menuEntry.backgroundFadeDuration,
    ease: 'power2.out'
  })

  // Menu slides up after background starts appearing
  gsap.to(menuContainer, {
    y: finalY,
    alpha: 1,
    duration: ANIMATION_CONFIG.menuEntry.duration,
    delay: ANIMATION_CONFIG.menuEntry.backgroundFadeDuration * 0.3, // Start sliding while background is fading
    ease: ANIMATION_CONFIG.menuEntry.ease
  })
}

// ============================================================================
// MAIN OVERLAY FUNCTION
// ============================================================================

/**
 * Creates a game mode selection overlay using image assets
 */
export function createGameModeSelectionOverlay(): GameModeSelectionOverlay {
  const container = new Container()
  container.visible = false
  container.zIndex = 1200

  const backgroundContainer = new Container()
  const menuContainer = new Container()
  const itemsContainer = new Container()

  container.addChild(backgroundContainer)
  container.addChild(menuContainer)
  menuContainer.addChild(itemsContainer) // Items should move with menu container

  const timers = createTimerManager()

  let isAnimating = false
  let onSelectCallback: ((modeId: string) => void) | null = null
  let canvasWidth = 600
  let canvasHeight = 800
  let itemContainers: Container[] = []

  /**
   * Select mode
   */
  function selectMode(modeId: string): void {
    if (!isAnimating) return

    const selectedIndex = GAME_MODES.findIndex(m => m.id === modeId)
    if (selectedIndex >= 0 && itemContainers[selectedIndex]) {
      const selectedItem = itemContainers[selectedIndex]
      gsap.to(selectedItem.scale, {
        x: ANIMATION_CONFIG.itemSelect.scale,
        y: ANIMATION_CONFIG.itemSelect.scale,
        duration: ANIMATION_CONFIG.itemSelect.duration,
        yoyo: ANIMATION_CONFIG.itemSelect.yoyo,
        repeat: ANIMATION_CONFIG.itemSelect.repeat,
        ease: ANIMATION_CONFIG.itemSelect.ease
      })
    }

    setTimeout(() => {
      if (onSelectCallback) {
        onSelectCallback(modeId)
      }
      hide()
    }, ANIMATION_CONFIG.selectDelay)
  }

  /**
   * Show overlay
   */
  function show(cWidth: number, cHeight: number, onSelect: (modeId: string) => void): void {
    timers.newSession()

    canvasWidth = cWidth
    canvasHeight = cHeight
    container.visible = true
    isAnimating = true
    onSelectCallback = onSelect

    backgroundContainer.removeChildren()
    menuContainer.removeChildren()
    itemsContainer.removeChildren()
    itemContainers = []

    const centerX = cWidth / 2
    const centerY = cHeight / 2

    // Dark overlay background
    createDarkOverlay(backgroundContainer, cWidth, cHeight, 0x000000, 0.7, false)

    // Make background clickable to close
    const bgHitArea = new Graphics()
    bgHitArea.rect(0, 0, cWidth, cHeight)
    bgHitArea.fill({ color: 0x000000, alpha: 0.01 })
    bgHitArea.eventMode = 'static'
    bgHitArea.cursor = 'pointer'
    bgHitArea.on('pointerdown', () => {
      // Play generic UI sound
      playGenericUISound()
      hide()
    })
    backgroundContainer.addChild(bgHitArea)

    // Menu background
    const menuBg = getGameModeSprite(MENU_CONFIG.backgroundImage)
    if (menuBg) {
      menuBg.anchor.set(0.5)
      const maxWidth = cWidth * MENU_CONFIG.maxWidthPercent
      const maxHeight = cHeight * MENU_CONFIG.maxHeightPercent
      const scaleX = maxWidth / menuBg.width
      const scaleY = maxHeight / menuBg.height
      const menuScale = Math.min(scaleX, scaleY, 1.0)
      menuBg.scale.set(menuScale)

      // Position menuBg relative to menuContainer (at center)
      menuBg.x = 0
      menuBg.y = 0

      menuBg.eventMode = 'static'
      menuBg.on('pointerdown', (e) => {
        e.stopPropagation()
      })

      menuContainer.addChild(menuBg)

      const menuWidth = menuBg.width
      const menuHeight = menuBg.height

      // Position menuContainer at center of canvas (will be animated)
      menuContainer.x = centerX
      // menuContainer.y will be set by animateMenuEntry

      // Add itemsContainer back to menuContainer (it was removed by removeChildren)
      menuContainer.addChild(itemsContainer)

      // Close button
      const closeBtn = createCloseButton(menuBg, menuWidth, menuHeight, hide)
      if (closeBtn) {
        menuContainer.addChild(closeBtn)
      }

      // Pre-calculate item heights
      const itemHeights: number[] = []
      GAME_MODES.forEach((mode, index) => {
        const tempItem = createMenuItem(mode, menuWidth, menuHeight, index, selectMode)
        const itemHeight = (tempItem as any)._itemHeight || menuHeight * 0.22
        itemHeights.push(itemHeight)
        tempItem.destroy({ children: true })
      })

      // Calculate list positioning (relative to menuContainer center)
      const { startY: listStartY, gap: itemGap } = calculateItemsListPosition(
        itemHeights,
        0, // menuContainer center is at 0 relative to itself
        menuHeight
      )

      // Create and position items
      GAME_MODES.forEach((mode, index) => {
        const item = createMenuItem(mode, menuWidth, menuHeight, index, selectMode)
        item.x = 0 // Relative to menuContainer center
        const finalY = calculateItemYPosition(index, itemHeights, itemGap, listStartY) // Use calculated listStartY

        // Items will be animated by animateItemEntry
        itemsContainer.addChild(item)
        itemContainers.push(item)

        // Animate item entry (slide up from below)
        animateItemEntry(item, index, finalY)
      })

      // Animate menu entry (slide up from below, background appears first)
      animateMenuEntry(menuContainer, centerY, backgroundContainer) // centerY is the final Y position on canvas
    }
  }

  /**
   * Hide overlay
   */
  function hide(): void {
    gsap.killTweensOf(menuContainer)
    gsap.killTweensOf(menuContainer.scale)
    gsap.killTweensOf(backgroundContainer)
    itemContainers.forEach(item => {
      gsap.killTweensOf(item)
      gsap.killTweensOf(item.scale)
    })

    container.visible = false
    isAnimating = false
    timers.clearAll()

    backgroundContainer.removeChildren()
    menuContainer.removeChildren()
    itemsContainer.removeChildren()

    itemContainers = []
  }

  /**
   * Update
   */
  function update(timestamp: number): void {
    // No particle effects needed
  }

  /**
   * Build/rebuild
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
