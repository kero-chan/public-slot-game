/**
 * Game Helper Utilities
 * Display-only functions for visual elements (NOT game logic)
 */

import { ASSETS } from '@/config/assets'
import { CONFIG } from '@/config/constants'
import { isBonusTile } from './tileHelpers'
import type { Grid } from '@/features/spin/types'

/**
 * Options for getRandomSymbol function
 */
export interface RandomSymbolOptions {
  /** Column index (0-4) */
  col?: number
  /** Visual row index (0-5) */
  visualRow?: number
  /** Whether gold variants are allowed */
  allowGold?: boolean
  /** Probability of gold (0-1), default 0.15 */
  goldChance?: number
  /** Probability of bonus (0-1), default 0.03 */
  bonusChance?: number
  /** Whether bonus tiles are allowed, default true */
  allowBonus?: boolean
}

/**
 * ⚠️ DISPLAY-ONLY FUNCTION - NOT GAME LOGIC
 *
 * This function generates random symbols for VISUAL DISPLAY ONLY:
 * - Initial grid display before first spin
 * - Reel strip animations
 * - Buffer rows (not visible to player)
 *
 * ❌ THIS IS NOT USED FOR GAME LOGIC:
 * - Backend generates ALL game grids
 * - Backend controls ALL RNG for actual gameplay
 * - Backend determines ALL win outcomes
 *
 * See /specs/09-security-architecture.md for details.
 *
 * @param options - Optional configuration
 * @returns Symbol name, possibly with "_gold" suffix
 *
 * @example
 * // Get random symbol with default settings
 * const symbol = getRandomSymbol()
 *
 * @example
 * // Get random symbol with gold allowed for column 2
 * const symbol = getRandomSymbol({ col: 2, allowGold: true })
 */
export function getRandomSymbol(options: RandomSymbolOptions = {}): string {
  const {
    col,
    allowGold = false,
    goldChance = 0.15,
    bonusChance = 0.03,
    allowBonus = true
  } = options

  // SECURITY: Symbol pool hardcoded to ONLY valid tile symbols (paytable removed from frontend)
  // Backend handles actual symbol spawning and game logic
  // These are the ONLY valid mahjong tile symbols that can appear on the reels
  const VALID_TILE_SYMBOLS = [
    'liangtong',
    'fa',
    'wusuo',
    'bai',
    'wutong',
    'bawan',
    'zhong',
    'liangsuo'
  ]
  const pool = VALID_TILE_SYMBOLS

  if (pool.length === 0) return 'fa'

  // Never generate wild tiles on the client for display

  // Small chance for bonus tile (3% by default)
  if (allowBonus && Math.random() < bonusChance) {
    return 'bonus'
  }

  let symbol = pool[Math.floor(Math.random() * pool.length)]

  // Check if we should make this a gold variant
  // Gold variants only appear in columns 1, 2, 3 (middle 3 columns of 0-4)
  // Not restricted by row - can appear in any row
  // Note: symbol from pool is already guaranteed to not be wild or bonus
  if (allowGold && Math.random() < goldChance) {
    const GOLD_ALLOWED_COLS = [1, 2, 3]

    const isAllowedColumn = col === undefined || GOLD_ALLOWED_COLS.includes(col)

    if (isAllowedColumn) {
      symbol = symbol + '_gold'
    }
  }

  return symbol
}

/**
 * Calculate dynamic counter duration based on number magnitude
 * Small numbers count faster, large numbers take more time
 * Uses logarithmic scale for natural feel
 *
 * @param amount - The target amount to count to
 * @returns Duration in seconds
 *
 * @example
 * getCounterDuration(10) // ~0.6s
 * getCounterDuration(1000) // ~1.8s
 * getCounterDuration(100000) // ~3.0s
 */
export function getCounterDuration(amount: number): number {
  // Logarithmic scale: small numbers are very fast, scales smoothly with magnitude
  // log10(10) = 1.0 → 0.6s
  // log10(100) = 2.0 → 1.2s
  // log10(1000) = 3.0 → 1.8s
  // log10(10000) = 4.0 → 2.4s
  // log10(100000) = 5.0 → 3.0s
  const duration = Math.log10(amount + 1) * 0.6
  return Math.max(0.3, Math.min(duration, 4.5)) // Min 0.3s, max 4.5s
}

/**
 * Create an empty grid with random symbols for initial display
 * Enforces bonus tile limits (max 1 per column in visible rows)
 *
 * @returns Grid with all rows (buffer + visible)
 *
 * @example
 * const grid = createEmptyGrid()
 * // Returns 5×10 grid with random symbols
 */
export function createEmptyGrid(): Grid {
  const grid: Grid = []
  const totalRows = CONFIG.reels.rows // Now 10 (includes buffer rows)
  const bufferRows = CONFIG.reels.bufferRows || 0
  const fullyVisibleRows = CONFIG.reels.fullyVisibleRows || 4

  // Fully visible rows start at bufferRows (row 4)
  const fullyVisibleStart = bufferRows
  const fullyVisibleEnd = bufferRows + fullyVisibleRows - 1

  for (let col = 0; col < CONFIG.reels.count; col++) {
    grid[col] = []
    let bonusCountInVisibleRows = 0

    // Create all rows (buffer + visible)
    for (let row = 0; row < totalRows; row++) {
      // Convert grid row to visual row for gold rules
      const visualRow = row - bufferRows

      // Check if we're in fully visible rows
      const isVisibleRow = row >= fullyVisibleStart && row <= fullyVisibleEnd

      // If we already have a bonus in this column's visible rows, don't allow more
      const allowBonus = !(isVisibleRow && bonusCountInVisibleRows >= 1)

      const symbol = getRandomSymbol({ col, visualRow, allowGold: true, allowBonus })
      grid[col][row] = symbol

      // Track bonus tiles in visible rows
      if (isBonusTile(symbol) && isVisibleRow) {
        bonusCountInVisibleRows++
      }
    }
  }
  return grid
}

/**
 * Create reel strips for animation purposes
 *
 * @param count - Number of reels
 * @param length - Length of each reel strip
 * @returns 2D array of symbol strings
 *
 * @example
 * const strips = createReelStrips(5, 30)
 * // Returns 5 strips, each with 30 symbols
 */
export function createReelStrips(count: number, length: number): string[][] {
  const strips: string[][] = []
  for (let c = 0; c < count; c++) {
    const strip: string[] = []
    for (let i = 0; i < length; i++) {
      // Allow gold in reel strips based on column
      // Visual row not specified since strips rotate
      // Bonus tiles allowed in strips but will be rare due to bonusChance
      strip.push(getRandomSymbol({ col: c, allowGold: true, allowBonus: true }))
    }
    strips.push(strip)
  }
  return strips
}

/**
 * Get the offset for buffer rows (gameRow = gridRow - bufferOffset)
 *
 * @returns Number of buffer rows
 *
 * @example
 * const offset = getBufferOffset() // Returns 4 for current config
 */
export function getBufferOffset(): number {
  return CONFIG.reels.bufferRows || 0
}

/**
 * Fill buffer rows with random symbols (called before win evaluation)
 * Mutates the grid in place
 *
 * @param grid - The grid to fill
 *
 * @example
 * fillBufferRows(grid)
 * // Buffer rows (0-3) now filled with random symbols
 */
export function fillBufferRows(grid: Grid): void {
  const bufferRows = CONFIG.reels.bufferRows || 0
  if (bufferRows === 0) return

  for (let col = 0; col < CONFIG.reels.count; col++) {
    for (let row = 0; row < bufferRows; row++) {
      grid[col][row] = getRandomSymbol()
    }
  }
}

/**
 * Enforce max 1 bonus tile per column in fully visible rows
 * Mutates the grid in place
 *
 * @param grid - The grid to validate and fix
 *
 * @example
 * enforceBonusLimit(grid)
 * // Grid now has at most 1 bonus per column in visible rows
 */
export function enforceBonusLimit(grid: Grid): void {
  const bufferRows = CONFIG.reels.bufferRows || 0
  const totalRows = CONFIG.reels.rows // Now 10 (includes buffer rows)
  const fullyVisibleRows = CONFIG.reels.fullyVisibleRows || 4

  // Fully visible rows start at bufferRows (row 4)
  const fullyVisibleStart = bufferRows
  const fullyVisibleEnd = bufferRows + fullyVisibleRows - 1

  for (let col = 0; col < CONFIG.reels.count; col++) {
    const bonusPositions: number[] = []

    // Find all bonus tiles in fully visible rows
    for (let row = 0; row < grid[col].length; row++) {
      const isVisibleRow = row >= fullyVisibleStart && row <= fullyVisibleEnd

      if (isVisibleRow && isBonusTile(grid[col][row])) {
        bonusPositions.push(row)
      }
    }

    // If more than 1 bonus tile, replace extras with random symbols
    if (bonusPositions.length > 1) {
      // Keep the first one, replace the rest
      for (let i = 1; i < bonusPositions.length; i++) {
        const row = bonusPositions[i]
        const visualRow = row - bufferRows
        // Replace with a random non-bonus symbol
        grid[col][row] = getRandomSymbol({
          col,
          visualRow,
          allowGold: true,
          allowBonus: false
        })
      }
    }
  }
}
