/**
 * Win Animation Constants
 * Shared constants for all win animation types
 * Centralized for easy maintenance and consistency
 */

/**
 * Layout constants - positioning and sizing
 */
export const WIN_ANIMATION_LAYOUT = {
  /** Maximum width as percentage of game view (0-1) */
  MAX_WIDTH_PERCENT: 0.9,
  /** Gap between win text and amount display as percentage of game view height (0-1) */
  TEXT_AMOUNT_GAP_PERCENT: 0.1,
  /** Mobile breakpoint width in pixels */
  MOBILE_BREAKPOINT: 600
} as const

/**
 * Amount display scale constants - unified scale, still smaller than image
 */
export const AMOUNT_DISPLAY_SCALE = {
  SMALL: {
    MOBILE: 0.9,
    PC: 1.0
  },
  MEDIUM: {
    MOBILE: 0.9,
    PC: 1.0
  },
  BIG: {
    MOBILE: 0.9,
    PC: 1.0
  },
  MEGA: {
    MOBILE: 0.9,
    PC: 1.0
  }
} as const

/**
 * Win text target height percentages - UNIFIED size for all win types
 * All win images display at the same size regardless of win type
 */
export const WIN_TEXT_HEIGHT_PERCENT = {
  SMALL: 0.18,
  MEDIUM: 0.18,
  BIG: 0.18,
  MEGA: 0.18,
  JACKPOT: 0.18
} as const

/**
 * Unified win image sizing configuration
 * Image will be sized to be prominent and visible
 */
export const WIN_IMAGE_SIZING = {
  /** Maximum width as percentage of viewport */
  MAX_WIDTH_PERCENT: 0.9,
  /** Maximum width in pixels (absolute limit) */
  MAX_WIDTH_PX: 600,
  /** Maximum height as percentage of viewport */
  MAX_HEIGHT_PERCENT: 0.35,
  /** Minimum padding from edges */
  EDGE_PADDING: 20
} as const

/**
 * Animation hold durations - how long each animation stays visible (seconds)
 */
export const WIN_ANIMATION_DURATION = {
  SMALL: 1.5,
  MEDIUM: 2.0,
  BIG: 2.5,
  MEGA: 3.0,
  JACKPOT: 3.5
} as const

/**
 * Win image configuration keys
 */
export const WIN_IMAGE_KEYS = {
  SMALL: 'win_small.webp',
  MEDIUM: 'win_medium.webp',
  BIG: 'win_big.webp',
  MEGA: 'win_mega.webp',
  JACKPOT: 'win_jackpot.webp'
} as const

/**
 * Win text fallback strings (used when image is not available)
 */
export const WIN_TEXT_FALLBACK = {
  SMALL: 'WIN!',
  MEDIUM: 'BIG WIN!',
  BIG: 'BIG WIN!',
  MEGA: 'MASSIVE WIN!',
  JACKPOT: 'JACKPOT!'
} as const

/**
 * Color constants for win animations
 */
export const WIN_COLORS = {
  GOLD: 0xffd700,
  YELLOW: 0xffff00,
  DARK_GOLD: 0xb8860b,
  ORANGE: 0xff6600,
  DARK_ORANGE: 0xcc6600,
  RED: 0xff0000,
  BLACK: 0x000000
} as const

/**
 * Helper function to check if current viewport is mobile
 */
export function isMobileViewport(width: number): boolean {
  return width < WIN_ANIMATION_LAYOUT.MOBILE_BREAKPOINT
}

/**
 * Helper function to get amount display scale based on win type and viewport
 */
export function getAmountDisplayScale(
  winType: 'SMALL' | 'MEDIUM' | 'BIG' | 'MEGA',
  isMobile: boolean
): number {
  const scaleConfig = AMOUNT_DISPLAY_SCALE[winType]
  return isMobile ? scaleConfig.MOBILE : scaleConfig.PC
}

/**
 * Helper function to calculate unified win image scale
 * Ensures all win images are displayed at the same size regardless of win type
 * and always fit nicely within the viewport
 */
export function calculateWinImageScale(
  spriteWidth: number,
  spriteHeight: number,
  canvasWidth: number,
  canvasHeight: number
): number {
  // Calculate max dimensions like settings menu sizing
  const maxWidth = Math.min(canvasWidth * WIN_IMAGE_SIZING.MAX_WIDTH_PERCENT, WIN_IMAGE_SIZING.MAX_WIDTH_PX)
  const maxHeight = canvasHeight * WIN_IMAGE_SIZING.MAX_HEIGHT_PERCENT

  // Calculate scale to fit within bounds
  const scaleByWidth = maxWidth / spriteWidth
  const scaleByHeight = maxHeight / spriteHeight

  // Use the smaller scale to ensure image fits
  return Math.min(scaleByWidth, scaleByHeight)
}

/**
 * Bonus overlay layout constants
 */
export const BONUS_OVERLAY_LAYOUT = {
  /** Panel max width as percentage of canvas width */
  PANEL_MAX_WIDTH_PERCENT: 0.85,
  /** Panel max width in pixels (absolute limit) */
  PANEL_MAX_WIDTH_PX: 400,
  /** Panel height as percentage of canvas height */
  PANEL_HEIGHT_PERCENT: 0.20,
  /** Panel center Y position as percentage of canvas height */
  PANEL_CENTER_Y_PERCENT: 0.45,
  /** Button max width as percentage of canvas width */
  BUTTON_MAX_WIDTH_PERCENT: 0.40,
  /** Button max height as percentage of canvas height */
  BUTTON_MAX_HEIGHT_PERCENT: 0.10,
  /** Button center Y position as percentage of canvas height */
  BUTTON_CENTER_Y_PERCENT: 0.80,
  /** Number display scale relative to panel */
  NUMBER_SCALE_PERCENT: 0.75,
  /** Label max width as percentage of canvas width */
  LABEL_MAX_WIDTH_PERCENT: 0.70
} as const

